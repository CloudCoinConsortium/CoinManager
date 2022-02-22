package raida

import (
	"context"
	"encoding/hex"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
)


type Freecoin struct {
	Servant
}

type FreecoinOutput struct {
  Ans []string
}

func NewFreecoin(progressChannel chan interface{}) *Freecoin {
	return &Freecoin{
		*NewServant(progressChannel),
	}
}


func (v *Freecoin) Get(ctx context.Context, sn uint32) (*FreecoinOutput, error) {
	logger.L(ctx).Debugf("Getting freecoin %d from the RAIDA", sn)

  var fco = &FreecoinOutput{}
	params := make([][]byte, v.Raida.TotalServers())
  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

  // Iterate over raida servers
	for idx, _ := range params {
  	params[idx] = v.GetHeaderSky(COMMAND_FREECOIN, idx, cce)
    encb := make([]byte, 0)
    encb = append(encb, v.GetChallenge()...)

    cbyte := utils.ExplodeSn(sn)

    encb = append(encb, cbyte...)
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
  pownArray, ans := v.ProcessGenericResponsesCommon(ctx, nil, results, v.CommonSuccessFunction, cce)

  logger.L(ctx).Debugf("Got ans %v", ans)

  err = nil
  if !v.IsQuorumCollected(pownArray) {
    details := v.GetErrorsFromResults(results)
    logger.L(ctx).Errorf("Not enough valid responses. Quorum is not reached")
    err = perror.NewWithDetails(perror.ERROR_RAIDA_QUORUM, "Not enough valid responses from the RAIDA servers", details)
    return nil, err
  }

  if len(ans) != config.TOTAL_RAIDA_NUMBER {
    logger.L(ctx).Errorf("Invalid length in response %d", len(ans))
    return nil, perror.New(perror.ERROR_INTERNAL, "Invalid response length")
  }

  fco.Ans = make([]string, config.TOTAL_RAIDA_NUMBER)

  for idx, v := range(ans) {
    if v == nil {
      fco.Ans[idx] = "00000000000000000000000000000000"
      continue
    }

    gv := v.([]byte)
    fco.Ans[idx] = hex.EncodeToString(gv)
  }

  return fco, nil

}

