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

type ShowCoinsByDenomination struct {
	Servant
}

type ShowCoinsByDenominationOutput struct {
  SerialNumbers []uint32 `json:"sns"`
}

type ShowCoinsByDenominationRawOutput struct {
  SerialNumbers map[uint32][]int `json:"sns"`
}

func NewShowCoinsByDenomination(progressChannel chan interface{}) *ShowCoinsByDenomination {
	return &ShowCoinsByDenomination{
		*NewServant(progressChannel),
	}
}

func (v *ShowCoinsByDenomination) ShowCoinsByDenomination(ctx context.Context, cc *cloudcoin.CloudCoin) (*ShowCoinsByDenominationOutput, error) {
	logger.L(ctx).Debugf("ShowCoinsByDenominationing for maxDenomination %d ", maxDenomination)

  so := &ShowCoinsByDenominationOutput{}
	params := make([][]byte, v.Raida.TotalServers())
  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

  sn := cc.Sn
  // Iterate over raida servers
	for idx, _ := range params {
  	params[idx] = v.GetHeaderSky(COMMAND_SHOW_COINS_BY_DENOMINATION, idx, cce)
    encb := make([]byte, 0)
    encb = append(encb, v.GetChallenge()...)

    cbytes := utils.ExplodeSn(sn)
    an, _ := hex.DecodeString(cc.Ans[idx])

    encb = append(encb, cbytes...)
    encb = append(encb, an...)

    // denomination
    encb = append(encb, byte(0x1))

    // Page
    encb = append(encb, byte(0x0))

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
  pownArray, rdata := v.ProcessGenericResponsesCommon(ctx, nil, results, v.SCBDSuccessFunction, cce)

  logger.L(ctx).Debugf("pown array %v", pownArray)
  if !v.IsQuorumCollected(pownArray) {
    details := v.GetErrorsFromResults(results)
    logger.L(ctx).Errorf("Not enough valid responses. Quorum is not reached")
    return nil, perror.NewWithDetails(perror.ERROR_RAIDA_QUORUM, "Not enough valid responses from the RAIDA servers", details)
  }

  sns := make(map[int]int, 0)
  var rsn int
  // Calculating best SNs
  for idx, item := range(rdata) {
    if item == nil {
      continue
    }
    bitem := item.([]byte)
    // SN is three bytes long
    if (len(bitem) % 3) != 0 {
      logger.L(ctx).Debugf("Incorrect response from RAIDA%d. Datalength must be multiple of 3. It is %d", idx, len(bitem))
      continue
    }

    for i := 0; i < len(bitem); i += 3 {
      rsn = int(utils.GetUint24(bitem[i:i+3]))

      logger.L(ctx).Debugf("RAIDA%d SN %d %d,%d,%d", idx, rsn, bitem[i], bitem[i+1], bitem[i+2])

      sns[rsn]++
    }
  }

  for sn, nums := range(sns) {
    if nums < config.MIN_QUORUM_COUNT {
      continue
    }

    logger.L(ctx).Debugf("Got SN %d", sn)
    so.SerialNumbers = append(so.SerialNumbers, uint32(sn))
  }

  logger.L(ctx).Debugf("d %v", rdata)
  logger.L(ctx).Debugf("d %v", sns)

  return so, nil
}

func (v *ShowCoinsByDenomination) ShowCoinsByDenominationRaw(ctx context.Context, cc *cloudcoin.CloudCoin) (*ShowCoinsByDenominationRawOutput, error) {
	logger.L(ctx).Debugf("ShowCoinsByDenominationingAll for maxDenomination %d ", maxDenomination)

  so := &ShowCoinsByDenominationRawOutput{}
  so.SerialNumbers = make(map[uint32][]int, 0)

	params := make([][]byte, v.Raida.TotalServers())
  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

  sn := cc.Sn
  // Iterate over raida servers
	for idx, _ := range params {
  	params[idx] = v.GetHeaderSky(COMMAND_SHOW_COINS_BY_DENOMINATION, idx, cce)
    encb := make([]byte, 0)
    encb = append(encb, v.GetChallenge()...)

    cbytes := utils.ExplodeSn(sn)
    an, _ := hex.DecodeString(cc.Ans[idx])

    encb = append(encb, cbytes...)
    encb = append(encb, an...)

    // denomination
    encb = append(encb, byte(0x1))

    // Page
    encb = append(encb, byte(0x0))
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
  pownArray, rdata := v.ProcessGenericResponsesCommon(ctx, nil, results, v.SCBDSuccessFunction, cce)

  logger.L(ctx).Debugf("pown array %v", pownArray)
  var rsn uint32
  // Calculating best SNs
  for idx, item := range(rdata) {
    if item == nil {
      continue
    }
    bitem := item.([]byte)
    // SN is three bytes long
    if (len(bitem) % 3) != 0 {
      logger.L(ctx).Debugf("Incorrect response from RAIDA%d. Datalength must be multiple of 3. It is %d", idx, len(bitem))
      continue
    }

    for i := 0; i < len(bitem); i += 3 {
      rsn = utils.GetUint24(bitem[i:i+3])

      _, ok := so.SerialNumbers[rsn]
      if !ok {
        // New SN
        so.SerialNumbers[rsn] = make([]int, v.Raida.TotalNumber)
        for i := 0; i < v.Raida.TotalNumber; i++ {
          so.SerialNumbers[rsn][i] = config.HEALTH_CHECK_STATUS_NOT_PRESENT
        }
      }

      so.SerialNumbers[rsn][idx] = config.HEALTH_CHECK_STATUS_PRESENT
      
      logger.L(ctx).Debugf("RAIDA%d SN %d %d,%d,%d", idx, rsn, bitem[i], bitem[i+1], bitem[i+2])
    }
  }

  // Fix network and counterfeits
  for sn, _ := range(so.SerialNumbers) {
    for idx, ps := range(pownArray) {
      switch ps {
      case config.RAIDA_STATUS_NORESPONSE:
        so.SerialNumbers[sn][idx] = config.HEALTH_CHECK_STATUS_NETWORK
      case config.RAIDA_STATUS_FAIL:
        so.SerialNumbers[sn][idx] = config.HEALTH_CHECK_STATUS_COUNTERFEIT
      case config.RAIDA_STATUS_UNTRIED:
        so.SerialNumbers[sn][idx] = config.HEALTH_CHECK_STATUS_UNKNOWN
      case config.RAIDA_STATUS_ERROR:
        so.SerialNumbers[sn][idx] = config.HEALTH_CHECK_STATUS_UNKNOWN
      }

      logger.L(ctx).Debugf("heath raida %d sn %d status %d (ps=%d)", idx, sn, so.SerialNumbers[sn][idx], ps)

    }
  }

  return so, nil
}

func (v *Servant) SCBDSuccessFunction(ctx context.Context, idx int, status int, coins []cloudcoin.CloudCoin, rdata []byte) (int, interface{}) {
  cmdName := v.Raida.DetectionAgents[idx].CurrentContext
  if (status == RESPONSE_STATUS_SUCCESS || status == RESPONSE_STATUS_PAGE_NOT_FOUND) {
    logger.L(ctx).Debugf("RAIDA%d (command %s) Success Status %d, rest data len %d", idx, cmdName, status, len(rdata))
    v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_PASS)
    return config.RAIDA_STATUS_PASS, rdata[config.RAIDA_RESPONSE_HEADER_SIZE:]
  }

  if (status == RESPONSE_STATUS_FAIL || status == RESPONSE_STATUS_FAILED_AUTH) {
    logger.L(ctx).Debugf("Raida%d Fail Status %d", idx, status)
    v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_FAIL)
    return config.RAIDA_STATUS_FAIL, nil
  } 
  
  logger.L(ctx).Debugf("Raida%d Unknown Status %d", idx, status)
  v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_ERROR)

  return config.RAIDA_STATUS_ERROR, nil
}


