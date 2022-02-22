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


type Detect struct {
	Servant
}

type DetectOutput struct {
  TotalCoins int `json:"total"`
  TotalAuthentic int `json:"authentic"`
  TotalFracked int `json:"fracked"`
  TotalLimbo int `json:"limbo"`
  TotalCounterfeit int `json:"counterfeit"`
  TotalUnknown int `json:"unknown"`
  Coins []CoinOutput `json:"coins"`
  Details [][]string `json:"details"`
}

func NewDetect(progressChannel chan interface{}) *Detect {
	return &Detect{
		*NewServant(progressChannel),
	}
}

func (v *Detect) GetStrideSize() int {
  // packetSize = 1024 - header(22) - challenge(16) - signature(2)
  coinsLen := 1024 - 22 - 16 - 2

  // Sn + An
  coinItemSize := 3 + 16

  return coinsLen / coinItemSize
}

func (v *Detect) Detect(ctx context.Context, coins []cloudcoin.CloudCoin) (*DetectOutput, error) {
  stride := v.GetStrideSize()
	logger.L(ctx).Debugf("Detecting %d notes (%d notes per packet)", len(coins), stride)

  if len(coins) == 0 {
    logger.L(ctx).Errorf("No coins to detect")
    return nil, perror.New(perror.ERROR_NO_COINS, "No coins to detect")
  }

  var do = &DetectOutput{}
  for i := 0; i < len(coins); i += stride {
    max := i + stride
    if max > len(coins) {
      max = len(coins)
    }

    logger.L(ctx).Debugf("coins slice %d:%d, Processing %d notes", i, max, (max - i))
    response, err := v.ProcessDetect(ctx, coins[i:max])
    if err != nil {
      logger.L(ctx).Errorf("Failed to process batch: %s", err.Error())
      for _, lcc := range(coins[i:max]) {
        do.TotalUnknown += lcc.GetDenomination()
      }

      perr := err.(*perror.ProgramError)
      do.Details = append(do.Details, perr.Details)
      continue
    }

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

func (v *Detect) ProcessDetect(ctx context.Context, coins []cloudcoin.CloudCoin) (*DetectOutput, error) {
  // All coins must have the same coinID

  isSky := false
  for _, cc:= range(coins) {
    logger.L(ctx).Debugf("Processing coin %d (isSky %v)",cc.Sn, cc.IsIDCoin())
    isSky = cc.IsIDCoin()
  }
  logger.L(ctx).Debugf("IsSky Detect %v", isSky)

	params := make([][]byte, v.Raida.TotalServers())

  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

  // Iterate over raida servers
	for idx, _ := range params {
    if isSky {
    	params[idx] = v.GetHeaderSky(COMMAND_DETECT, idx, cce)
    } else {
    	params[idx] = v.GetHeader(COMMAND_DETECT, idx, cce)
    }

    /*
    params[idx] = append(params[idx], v.GetChallenge()...)

    // Put coins in the array for this raida
    for _, cc := range(coins) {
      sn := cc.Sn
      cbyte := utils.ExplodeSn(sn)
      data, _ := hex.DecodeString(cc.Ans[idx])
      params[idx] = append(params[idx], cbyte...)
      params[idx] = append(params[idx], data...)
    }

    params[idx] = append(params[idx], v.GetSignature()...)
  */

    encb := make([]byte, 0)
    encb = append(encb, v.GetChallenge()...)
    for _, cc := range(coins) {
      sn := cc.Sn
      cbyte := utils.ExplodeSn(sn)
      data, _ := hex.DecodeString(cc.Ans[idx])
      encb = append(encb, cbyte...)
      encb = append(encb, data...)
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

  logger.L(ctx).Debug("Detect Results")
  var a, f, c, u, l, total int
  coinResults := make([]CoinOutput, len(coins))
  for idx, cc := range (coins) {
    coins[idx].Grade()

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

    logger.L(ctx).Debugf("Coin #%d: %s (%s)", coins[idx].Sn, coins[idx].GetGradeStatusString(), coins[idx].GetPownString())
    coinResults[idx] = CoinOutput{
      Sn: cc.Sn,
      PownString: coins[idx].GetPownString(),
      Result: coins[idx].GetGradeStatusString(),
    }
  }

  dr := &DetectOutput{
    TotalCoins: total,
    TotalAuthentic: a,
    TotalFracked: f,
    TotalCounterfeit: c,
    TotalLimbo: l,
    TotalUnknown: u,
  }

  dr.Coins = coinResults

  return dr, err
}
