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


type Break struct {
	Servant
}

type BreakOutput struct {
}

func NewBreak(progressChannel chan interface{}) *Break {
	return &Break{
		*NewServant(progressChannel),
	}
}

func (v *Break) Break(ctx context.Context, cc *cloudcoin.CloudCoin, sns []uint32, pgs []string) (*BreakOutput, error) {
	logger.L(ctx).Debugf("Breaking Coin %d", cc.Sn)

  bo := &BreakOutput{}
	params := make([][]byte, v.Raida.TotalServers())
  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

  psn := uint32(config.PUBLIC_CHANGE_SN)
  sn := cc.Sn
  // Iterate over raida servers
	for idx, _ := range params {
  	params[idx] = v.GetHeaderSky(COMMAND_BREAK, idx, cce)

    encb := make([]byte, 0)
    encb = append(encb, v.GetChallenge()...)

    cbytes := utils.ExplodeSn(sn)

    encb = append(encb, cbytes...)

    ans, err := hex.DecodeString(cc.Ans[idx])
		if err != nil {
		  logger.L(ctx).Debugf("Failed to decode ANS for coin %d", cc.Sn)
			return nil, err
		}

    bpgs, err := hex.DecodeString(pgs[idx])
		if err != nil {
		  logger.L(ctx).Debugf("Failed to decode ANS for coin %d", cc.Sn)
			return nil, err
		}

    logger.L(ctx).Debugf("pgs %s %s %v", cc.Ans[idx], pgs[idx], bpgs)
    encb = append(encb, ans...)
    encb = append(encb, bpgs...)

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

func (v *Break) BreakCoins(ctx context.Context, cc *cloudcoin.CloudCoin, coins []cloudcoin.CloudCoin) (*BreakOutput, error) {
	logger.L(ctx).Debugf("Breaking Coin %d", cc.Sn)

  bo := &BreakOutput{}
	params := make([][]byte, v.Raida.TotalServers())
  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

  psn := uint32(config.PUBLIC_CHANGE_SN)
  sn := cc.Sn
  // Iterate over raida servers
	for idx, _ := range params {
  	params[idx] = v.GetHeader(COMMAND_BREAK, idx, cce)

    encb := make([]byte, 0)
    encb = append(encb, v.GetChallenge()...)

    // SN of the coin that will break
    cbytes := utils.ExplodeSn(sn)
    encb = append(encb, cbytes...)

    // ANs of this coin
    ans, err := hex.DecodeString(cc.Ans[idx])
		if err != nil {
		  logger.L(ctx).Debugf("Failed to decode ANS for coin %d", cc.Sn)
			return nil, err
		}
    encb = append(encb, ans...)

    // Owner ID (Public Change)
    pbytes := utils.ExplodeSn(psn)
    encb = append(encb, pbytes...)

    // Add Pans 
    for _, ccc := range(coins) {
      csn := ccc.Sn

      ccbyte := utils.ExplodeSn(csn)
      cdata, _ := hex.DecodeString(ccc.Pans[idx])

      encb = append(encb, ccbyte...)
      encb = append(encb, cdata...)
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
  pownArray, _ := v.ProcessGenericResponses(ctx, coins, results, v.CommonSuccessFunction)

  logger.L(ctx).Debugf("pown array %s", pownArray)
  if !v.IsQuorumCollected(pownArray) {
    details := v.GetErrorsFromResults(results)
    logger.L(ctx).Errorf("Not enough valid responses. Quorum is not reached")
    return nil, perror.NewWithDetails(perror.ERROR_RAIDA_QUORUM, "Not enough valid responses from the RAIDA servers", details)
  }

  return bo, nil
}
