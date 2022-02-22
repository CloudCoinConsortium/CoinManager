package raida

import (
	"encoding/hex"
  "context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
)


type BreakInBank struct {
	Servant
}

type BreakInBankOutput struct {
}

func NewBreakInBank(progressChannel chan interface{}) *BreakInBank {
	return &BreakInBank{
		*NewServant(progressChannel),
	}
}

func (v *BreakInBank) BreakInBank(ctx context.Context, owcc *cloudcoin.CloudCoin, snToBreak uint32, sns []uint32) (*BreakInBankOutput, error) {
	logger.L(ctx).Debugf("BreakInBanking Coin %d", snToBreak)

  bo := &BreakInBankOutput{}
	params := make([][]byte, v.Raida.TotalServers())
  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

  psn := uint32(config.PUBLIC_CHANGE_SN)
  // Iterate over raida servers
	for idx, _ := range params {
  	params[idx] = v.GetHeaderSky(COMMAND_BREAK_IN_BANK, idx, cce)
 
    encb := make([]byte, 0)
    encb = append(encb, v.GetChallenge()...)

    owcbytes := utils.ExplodeSn(owcc.Sn)
    encb = append(encb, owcbytes...)

    ans, err := hex.DecodeString(owcc.Ans[idx])
		if err != nil {
		  logger.L(ctx).Debugf("Failed to decode ANS for coin %d", owcc.Sn)
			return nil, err
		}
    encb = append(encb, ans...)

    cbytes := utils.ExplodeSn(snToBreak)
    encb = append(encb, cbytes...)

    pbytes := utils.ExplodeSn(psn)
    encb = append(encb, pbytes...)

    for _, csn := range(sns) {
      rbytes := utils.ExplodeSn(csn)

      encb = append(encb, rbytes...)
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
  pownArray, _ := v.ProcessGenericResponsesCommon(ctx, nil, results, v.CommonSuccessFunction, cce)

  logger.L(ctx).Debugf("pown array %s", pownArray)
  if !v.IsQuorumCollected(pownArray) {
    details := v.GetErrorsFromResults(results)
    logger.L(ctx).Errorf("Not enough valid responses. Quorum is not reached")
    return nil, perror.NewWithDetails(perror.ERROR_RAIDA_QUORUM, "Not enough valid responses from the RAIDA servers", details)
  }

  return bo, nil
}

