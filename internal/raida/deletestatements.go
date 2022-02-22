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

type DeleteAllStatements struct {
	Servant
}

func NewDeleteAllStatements(progressChannel chan interface{}) *DeleteAllStatements {
	return &DeleteAllStatements{
		*NewServant(progressChannel),
	}
}

func (v *DeleteAllStatements) DeleteAllStatements(ctx context.Context, cc *cloudcoin.CloudCoin) error {
	logger.L(ctx).Debugf("DeleteAllStatementsing for maxDenomination %d ", maxDenomination)

	params := make([][]byte, v.Raida.TotalServers())
  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

  sn := cc.Sn
  // Iterate over raida servers
	for idx, _ := range params {
  	params[idx] = v.GetHeaderSky(COMMAND_DELETE_ALL_STATEMENTS, idx, cce)
    encb := make([]byte, 0)
    encb = append(encb, v.GetChallenge()...)

    cbytes := utils.ExplodeSn(sn)
    an, _ := hex.DecodeString(cc.Ans[idx])

    encb = append(encb, cbytes...)
    encb = append(encb, an...)
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
  pownArray, _ := v.ProcessGenericResponsesCommon(ctx, nil, results, v.DeleteSuccessFunction, cce)

  if !v.IsQuorumCollected(pownArray) {
    details := v.GetErrorsFromResults(results)
    logger.L(ctx).Errorf("Not enough valid responses. Quorum is not reached")
    return perror.NewWithDetails(perror.ERROR_RAIDA_QUORUM, "Not enough valid responses from the RAIDA servers", details)
  }

  logger.L(ctx).Debugf("Deleted successully")

  return nil
}

func (v *Servant) DeleteSuccessFunction(ctx context.Context, idx int, status int, coins []cloudcoin.CloudCoin, rdata []byte) (int, interface{}) {
  cmdName := v.Raida.DetectionAgents[idx].CurrentContext
  if (status == RESPONSE_STATUS_SUCCESS || status == RESPONSE_STATUS_NO_STATEMENTS || status == RESPONSE_STATUS_STATEMENTS_DELETED) {
    logger.L(ctx).Debugf("RAIDA%d (command %s) Success Status %d, rest data len %d", idx, cmdName, status, len(rdata))
    return config.RAIDA_STATUS_PASS, rdata[config.RAIDA_RESPONSE_HEADER_SIZE:]
  }

  if (status == RESPONSE_STATUS_FAIL || status == RESPONSE_STATUS_FAILED_AUTH) {
    logger.L(ctx).Debugf("Raida%d Fail Status %d", idx, status)
    return config.RAIDA_STATUS_FAIL, nil
  } 
  
  logger.L(ctx).Debugf("Raida%d Unknown Status %d", idx, status)

  return config.RAIDA_STATUS_ERROR, nil
}
