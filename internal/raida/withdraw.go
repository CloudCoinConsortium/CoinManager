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


type Withdraw struct {
	Servant
}

type WithdrawOutput struct {
  TotalCoins int `json:"total"`
  TotalAuthentic int `json:"authentic"`
  TotalFracked int `json:"fracked"`
  TotalLimbo int `json:"limbo"`
  TotalCounterfeit int `json:"counterfeit"`
  TotalUnknown int `json:"unknown"`
  Coins []CoinOutput `json:"coins"`
  Details [][]string `json:"details"`
}

func NewWithdraw(progressChannel chan interface{}) *Withdraw {
	return &Withdraw{
		*NewServant(progressChannel),
	}
}

func (v *Withdraw) GetStrideSize() int {
  // packetSize = 1024 - header(22) - challenge(16) - signature(2)
  coinsLen := 1024 - 22 - 16 - 2


  // owcbyte
  coinsLen -= 3

  // owAN
  coinsLen -= 16

  // PGs
  coinsLen -= 16

  // receiptID
  coinsLen -= 16

  // ts
  coinsLen -= 6

  // TY
  coinsLen -= 1

  // RaidType is in memo
  coinsLen -= config.MAX_MEMO_LENGTH

  coinItemSize := 3 

  return coinsLen / coinItemSize
}

func (v *Withdraw) Withdraw(ctx context.Context, ownerCC *cloudcoin.CloudCoin, coins []cloudcoin.CloudCoin, receiptID string, memo string) (*WithdrawOutput, error) {
  stride := v.GetStrideSize()
	logger.L(ctx).Debugf("Withdrawing %d notes receipt %s (%d notes per packet)", len(coins), receiptID, stride)

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

  // Generating PGs
  pgs := make([]string, 0)
  for i := 0; i < config.TOTAL_RAIDA_NUMBER; i++ {
    pg, _ := utils.GeneratePG()
    pgs = append(pgs, pg)
  }

  logger.L(ctx).Debugf("ts and memos %v %v pgs %v", bts, memos, pgs)

  var do = &WithdrawOutput{}
  do.Details = make([][]string, 0)
  for i := 0; i < len(coins); i += stride {
    max := i + stride
    if max > len(coins) {
      max = len(coins)
    }

    logger.L(ctx).Debugf("coins slice %d:%d, Processing %d notes", i, max, (max - i))
    response, err := v.ProcessWithdraw(ctx, *ownerCC, coins[i:max], breceiptID, bts, memos, pgs)
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

func (v *Withdraw) ProcessWithdraw(ctx context.Context, ownerCC cloudcoin.CloudCoin, coins []cloudcoin.CloudCoin, breceiptID []byte, bts []byte, memos []string, pgs []string) (*WithdrawOutput, error) {
	params := make([][]byte, v.Raida.TotalServers())
  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

  // Iterate over raida servers
	for idx, _ := range params {
  	params[idx] = v.GetHeaderSky(COMMAND_WITHDRAW, idx, cce)
    encb := make([]byte, 0)
    encb = append(encb, v.GetChallenge()...)

    owCbyte := utils.ExplodeSn(ownerCC.Sn)
    owAn, _ := hex.DecodeString(ownerCC.Ans[idx])

    encb = append(encb, owCbyte...)
    encb = append(encb, owAn...)
    // Put coins in the array for this raida
    for _, cc := range(coins) {
      sn := cc.Sn

      cbyte := utils.ExplodeSn(sn)
      encb = append(encb, cbyte...)
    }

    // PG
    pgbyte, _ := hex.DecodeString(pgs[idx])

    logger.L(ctx).Debugf("bt %d %v", len(bts), pgbyte)

    encb = append(encb, pgbyte...)
    encb = append(encb, breceiptID...)

    // Timestamp is 6 bytes 
    encb = append(encb, bts...)

    // TY = 1 here
    encb = append(encb, 0x1)

    // Memo
    // RAID TYPE
    encb = append(encb, config.STATEMENT_RAID_TYPE_STRIPE)

    // Memo itself
    bsmemo := []byte(memos[idx])
    encb = append(encb, bsmemo...)

    // Trailing zeroes
    zcnt := config.MAX_MEMO_LENGTH - 1 - len(bsmemo)
    logger.L(ctx).Debugf("a%d len=%d r=%d", zcnt, len(bsmemo), idx)
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
  pownArray, _ := v.ProcessGenericResponsesCommon(ctx, coins, results, v.CommonMixedSuccessFunction, cce)

  err = nil
  if !v.IsQuorumCollected(pownArray) {
    details := v.GetErrorsFromResults(results)
    logger.L(ctx).Errorf("Not enough valid responses. Quorum is not reached")
    err = perror.NewWithDetails(perror.ERROR_RAIDA_QUORUM, "Not enough valid responses from the RAIDA servers", details)
  }
  

  logger.L(ctx).Debug("Withdraw Results")
  var a, f, c, u, l, total int
  coinResults := make([]CoinOutput, len(coins))
  for idx, cc := range (coins) {
    coins[idx].Grade()
    for i := 0; i < v.Raida.TotalNumber; i++ {

      an := cloudcoin.GetANFromPG(i, coins[idx].Sn, pgs[i])

      logger.L(ctx).Debugf("Setting an%d for sn%d to %s", i, coins[idx].Sn, coins[idx].Ans[i])
      coins[idx].SetAn(i, an)
    }

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

  pr := &WithdrawOutput{
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

