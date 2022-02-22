package raida

import (
	"context"
	"encoding/hex"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
)


type Deposit struct {
	Servant
}

type DepositOutput struct {
  TotalCoins int `json:"total"`
  TotalAuthentic int `json:"authentic"`
  TotalFracked int `json:"fracked"`
  TotalLimbo int `json:"limbo"`
  TotalCounterfeit int `json:"counterfeit"`
  TotalUnknown int `json:"unknown"`
  Coins []CoinOutput `json:"coins"`
  Details [][]string `json:"details"`
}

func NewDeposit(progressChannel chan interface{}) *Deposit {
	return &Deposit{
		*NewServant(progressChannel),
	}
}

func (v *Deposit) GetStrideSize() int {
  // packetSize = 1024 - header(22) - challenge(16) - signature(2)
  coinsLen := 1024 - 22 - 16 - 2

  // owner
  coinsLen -= 3

  // receiptID
  coinsLen -= 16

  // ts
  coinsLen -= 6

  // coinType
  coinsLen -= 1


  // TY
  coinsLen -= 1

  // RAID TYPE is a part of memo
  coinsLen -= config.MAX_MEMO_LENGTH
  
  // Sn + An
  coinItemSize := 3 + 16

  return coinsLen / coinItemSize
}

func (v *Deposit) Deposit(ctx context.Context, coins []cloudcoin.CloudCoin, to uint32, receiptID string, memo string) (*DepositOutput, error) {
  stride := v.GetStrideSize()
	logger.L(ctx).Debugf("Depositing %d notes receipt %s (%d notes per packet)", len(coins), receiptID, stride)


  breceiptID, err := hex.DecodeString(receiptID)
  if err != nil {
    logger.L(ctx).Errorf("Failed to decode receipt id %s", receiptID, err.Error())
    return nil, perror.New(perror.ERROR_DECODE_HEX, "Failed to decode receiptID " + receiptID)
  }

  bts := utils.GetCurrentTsBytes()
  memos := v.GetStripesAndMirrors(memo)
  for _, memo := range(memos) {
    // 1 is RAID Type
    if len(memo) > config.MAX_MEMO_LENGTH - 1 {
      return nil, perror.New(perror.ERROR_MEMO_TOO_LONG, "Memo is too long")
    }
  }

  logger.L(ctx).Debugf("ts and memos %v %v", bts, memos)

  var do = &DepositOutput{}
  do.Details = make([][]string, 0)
  for i := 0; i < len(coins); i += stride {
    max := i + stride
    if max > len(coins) {
      max = len(coins)
    }

    logger.L(ctx).Debugf("coins slice %d:%d, Processing %d notes", i, max, (max - i))
    response, err := v.ProcessDeposit(ctx, coins[i:max], to, breceiptID, bts, memos)
    if err != nil {
      logger.L(ctx).Errorf("Failed to process batch: %s", err.Error())
      perr := err.(*perror.ProgramError)
      do.Details = append(do.Details, perr.Details)
    }

    logger.L(ctx).Debugf("Batch Function %v", v.batchFunction)
    if (v.batchFunction != nil) {
      err := v.batchFunction(ctx, coins[i:max])
      if err != nil {
        logger.L(ctx).Errorf("Failed to call batch function:" + err.Error())
        for _, lcc := range(coins[i:max]) {
          do.TotalUnknown += lcc.GetDenomination()
        }

        errs := v.GetProgramErrors(err)
        do.Details = append(do.Details, errs)
      }
    }

    do.TotalAuthentic += response.TotalAuthentic
    do.TotalCounterfeit += response.TotalCounterfeit
    do.TotalFracked += response.TotalFracked
    do.TotalLimbo += response.TotalLimbo
    do.TotalUnknown += response.TotalUnknown
    do.Coins = append(do.Coins, response.Coins...)

  }

  for _, cc:= range(coins) {
    do.TotalCoins += cc.GetDenomination()
  }

  return do, nil

}

func (v *Deposit) ProcessDeposit(ctx context.Context, coins []cloudcoin.CloudCoin, to uint32, breceiptID []byte, bts []byte, memos []string) (*DepositOutput, error) {
	params := make([][]byte, v.Raida.TotalServers())
  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

  // Iterate over raida servers
	for idx, _ := range params {
  	params[idx] = v.GetHeaderSky(COMMAND_DEPOSIT, idx, cce)
    encb := make([]byte, 0)
    encb = append(encb, v.GetChallenge()...)

    // Put coins in the array for this raida
    for _, cc := range(coins) {
      sn := cc.Sn

      cbyte := utils.ExplodeSn(sn)
      data, _ := hex.DecodeString(cc.Ans[idx])

      encb = append(encb, cbyte...)
      encb = append(encb, data...)
    }

    // Adding new Owner
    cbyte := utils.ExplodeSn(uint32(to))
 
    logger.L(ctx).Debugf("bt %d", len(bts))

    encb = append(encb, cbyte...)
    encb = append(encb, breceiptID...)

    // Timestamp is 6 bytes 
    encb = append(encb, bts...)

    // Coin Type
    encb = append(encb, 0x0)

    // TY 
    encb = append(encb, 0x0)

    // Memo
    // RAID TYPE
    encb = append(encb, config.STATEMENT_RAID_TYPE_STRIPE)

    // Memo itself
    bsmemo := []byte(memos[idx])
    encb = append(encb, bsmemo...)

    // Trailing zeroes
    zcnt := config.MAX_MEMO_LENGTH - 1 - len(bsmemo)
    for i := 0; i < zcnt; i++ {
      encb = append(encb, 0x0)
    }

    encb = append(encb, v.GetSignature()...)
    encb, err = v.EncryptIfRequired(ctx, cce, idx, encb)
    if err != nil {
      logger.L(ctx).Debugf("Failed to encrypt body for R%d: %s", idx, err.Error())
      continue
    }
    params[idx] = append(params[idx], encb...)
	}

  v.UpdateHeaderUdpPackets(params)
	results := v.Raida.SendRequest(ctx, params, 0)
  pownArray, tickets := v.ProcessGenericResponsesCommon(ctx, coins, results, v.CommonMixedSuccessFunction, cce)

  err = nil
  if !v.IsQuorumCollected(pownArray) {
    details := v.GetErrorsFromResults(results)
    logger.L(ctx).Errorf("Not enough valid responses. Quorum is not reached")
    err = perror.NewWithDetails(perror.ERROR_RAIDA_QUORUM, "Not enough valid responses from the RAIDA servers", details)
  }
  

  logger.L(ctx).Debug("Deposit Results")
  var a, f, c, u, l, total int
  coinResults := make([]CoinOutput, len(coins))
  for idx, cc := range (coins) {
    coins[idx].Grade()
    coins[idx].SetAnsToPansIfPassed()

    total += cc.GetDenomination()
    switch (coins[idx].GetGradeStatus()) {
    case config.COIN_STATUS_AUTHENTIC:
      a += cc.GetDenomination()
    case config.COIN_STATUS_COUNTERFEIT:
      c += cc.GetDenomination()
    case config.COIN_STATUS_LIMBO:
      l += cc.GetDenomination()
    case config.COIN_STATUS_FRACKED:
      f += cc.GetDenomination()
    default:
      u += cc.GetDenomination()
    }

    logger.L(ctx).Debugf("Coin #%d: %s (%s)", cc.Sn, coins[idx].GetGradeStatusString(), coins[idx].GetPownString())
    coinResults[idx] = CoinOutput{
      Sn: cc.Sn,
      PownString: coins[idx].GetPownString(),
      Result: coins[idx].GetGradeStatusString(),
    }
  }

  sTickets := make([]string, len(tickets))
  if (tickets != nil) {
    for idx, ticket := range(tickets) {
      if (ticket == nil) {
        sTickets[idx] = ""
        continue
      }
      sTickets[idx] = hex.EncodeToString(ticket.([]byte))
    }

    logger.L(ctx).Debugf("tickets %v", sTickets)
  }

  pr := &DepositOutput{
    TotalCoins: total,
    TotalAuthentic: a,
    TotalFracked: f,
    TotalCounterfeit: c,
    TotalLimbo: l,
    TotalUnknown: u,
  }

  pr.Coins = coinResults

  return pr, err
}

