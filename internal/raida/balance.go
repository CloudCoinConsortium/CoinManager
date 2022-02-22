package raida

import (
	"encoding/binary"
	"encoding/hex"
  "context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
)


type Balance struct {
	Servant
}

type BalanceOutput struct {
  Total int `json:"total"`
  D250 int `json:"d250"`
  D100 int `json:"d100"`
  D25 int `json:"d25"`
  D5 int `json:"d5"`
  D1 int `json:"d1"`
}

type BalanceRawOutput struct {
  Total []int `json:"total"`
  D250 []int `json:"d250"`
  D100 []int `json:"d100"`
  D25 []int `json:"d25"`
  D5 []int `json:"d5"`
  D1 []int `json:"d1"`
}



func NewBalance(progressChannel chan interface{}) *Balance {
	return &Balance{
		*NewServant(progressChannel),
	}
}


func (v *Balance) Balance(ctx context.Context, cc *cloudcoin.CloudCoin) (*BalanceOutput, error) {
	logger.L(ctx).Debugf("Getting Balance for %s sn %d", cc.GetSkyName(), cc.Sn)

	params := make([][]byte, v.Raida.TotalServers())
  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }


  // Iterate over raida servers
	for idx, _ := range params {
  	params[idx] = v.GetHeaderSky(COMMAND_BALANCE, idx, cce)

    encb := make([]byte, 0)
    encb = append(encb, v.GetChallenge()...)

    sn := cc.Sn
    cbyte := utils.ExplodeSn(sn)

    data, _ := hex.DecodeString(cc.Ans[idx])
    encb = append(encb, cbyte...)
    encb = append(encb, data...)
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
  pownArray, balances := v.ProcessGenericResponsesCommon(ctx, []cloudcoin.CloudCoin{*cc}, results, v.BalanceSuccessFunction, cce)

  logger.L(ctx).Debugf("balances response before grade %s", cc.GetGradeStatusString())
  cc.Grade()
  logger.L(ctx).Debugf("balances responses after grade %s", cc.GetGradeStatusString())

  if cc.GetGradeStatus() == config.COIN_STATUS_COUNTERFEIT {
    logger.L(ctx).Debugf("ID coin is couterfeit")
    return nil, perror.New(perror.ERROR_COIN_COUTERFEIT, "ID coin is counterfeit")
  }

  if !v.IsQuorumCollected(pownArray) {
    logger.L(ctx).Errorf("Not enough valid responses. Quorum is not reached. It may be that the wallet is new")
    return nil, perror.New(perror.ERROR_RAIDA_QUORUM, "Not enough valid balance responses. It is ok if the wallet is new and has no transactions yet")
  }

  counters := make(map[string]int, 0)
  vals := make(map[string]*BalanceOutput, 0)
  for idx, iv := range(balances) {
    if iv == nil {
      continue
    }

    v := iv.([]byte)
    // We expect 19 bytes (4xTotal and 3bytes per each denomination)
    //if len(v) != 19 {
    // We expect 4 bytes (4xTotal)
    if len(v) != 4 {
      logger.L(ctx).Warnf("Raida%d returned incorrect length for balance call: %d", idx, len(v))
      continue
    }

    key := hex.EncodeToString(v)
    total := int(binary.BigEndian.Uint32(v[0:4]))
/*
    dn250 := int(utils.GetUint24(v[4:7]))
    dn100 := int(utils.GetUint24(v[7:10]))
    dn25 := int(utils.GetUint24(v[10:13]))
    dn5 := int(utils.GetUint24(v[13:16]))
    dn1 := int(utils.GetUint24(v[16:]))

    br := &BalanceOutput{
      Total: total,
      D250: dn250,
      D100: dn100,
      D25: dn25,
      D5: dn5,
      D1: dn1,
    }
*/

    br := &BalanceOutput{
      Total: total,
    }

    //logger.L(ctx).Debugf("Raida%d, total %d, 250s:%d, 100s:%d, 25s:%d, 5s:%d, 1s:%d", idx, total, dn250, dn100, dn25, dn5, dn1)
    logger.L(ctx).Debugf("Raida%d, total %d", idx, total)

    counters[key]++
    vals[key] = br
  }

  for i, v := range(counters) {
    logger.L(ctx).Debugf("i=%s v=%d", i,v)
  }

  topKey, err := PickTopKey(ctx, counters)
  if err != nil {
    return nil, perror.New(perror.ERROR_RAIDA_QUORUM, "Failed to pick top balance")
  }

  brfinal := vals[topKey]
  logger.L(ctx).Debugf("Balance chosen %v", brfinal)

  return brfinal, nil
}

// Function that can return SUCCESS or FAIL
func (v *Servant) BalanceSuccessFunction(ctx context.Context, idx int, status int, coins []cloudcoin.CloudCoin, rdata []byte) (int, interface{}) {
  if (status == RESPONSE_STATUS_SUCCESS) {
    logger.L(ctx).Debugf("Raida%d Success Status %d, rest data len %d", idx, status, len(rdata))
    v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_PASS)
    return config.RAIDA_STATUS_PASS, rdata[config.RAIDA_RESPONSE_HEADER_SIZE:]
  }

  // This is akward but the RAIDA returns FAIL if the wallet is just created
  if (status == RESPONSE_STATUS_FAIL) {
    logger.L(ctx).Debugf("Raida%d Fail Status %d. We assume that the coin is just created", idx, status)
    v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_ERROR)
    return config.RAIDA_STATUS_ERROR, nil
  } 

  if (status == RESPONSE_STATUS_FAILED_AUTH) {
    logger.L(ctx).Debugf("Raida%d Fail Status %d", idx, status)
    v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_FAIL)
    return config.RAIDA_STATUS_FAIL, nil
  } 
  
  logger.L(ctx).Debugf("Raida%d Unknown Status %d", idx, status)
  v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_ERROR)

  return config.RAIDA_STATUS_ERROR, nil
}

func (v *Balance) BalanceRaw(ctx context.Context, cc *cloudcoin.CloudCoin) (*BalanceRawOutput, error) {
	logger.L(ctx).Debugf("Getting Raw Balance for %s sn %d", cc.GetSkyName(), cc.Sn)

	params := make([][]byte, v.Raida.TotalServers())
  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

  // Iterate over raida servers
	for idx, _ := range params {
  	params[idx] = v.GetHeaderSky(COMMAND_BALANCE, idx, cce)

    encb := make([]byte, 0)
    encb = append(encb, v.GetChallenge()...)

    sn := cc.Sn
    cbyte := utils.ExplodeSn(sn)

    data, _ := hex.DecodeString(cc.Ans[idx])
    encb = append(encb, cbyte...)
    encb = append(encb, data...)
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
  _, balances := v.ProcessGenericResponsesCommon(ctx, []cloudcoin.CloudCoin{*cc}, results, v.BalanceSuccessFunction, cce)

  brs := &BalanceRawOutput{
    Total: make([]int, v.Raida.TotalServers()),
    D250: make([]int, v.Raida.TotalServers()),
    D100: make([]int, v.Raida.TotalServers()), 
    D25:  make([]int, v.Raida.TotalServers()), 
    D5: make([]int, v.Raida.TotalServers()),
    D1: make([]int, v.Raida.TotalServers()),
  }

  for idx, iv := range(balances) {
    if iv == nil {
      continue
    }

    v := iv.([]byte)
    // We expect 19 bytes (4xTotal and 3bytes per each denomination)
    if len(v) != 4 {
      logger.L(ctx).Warnf("Raida%d returned incorrect length for balance call: %d", idx, len(v))
      continue
    }

    key := hex.EncodeToString(v)

    logger.L(ctx).Debugf("r%d key %s", idx, key)

    brs.Total[idx] = int(binary.BigEndian.Uint32(v[0:4]))
/*    brs.D250[idx] = int(utils.GetUint24(v[4:7]))
    brs.D100[idx] = int(utils.GetUint24(v[7:10]))
    brs.D25[idx] = int(utils.GetUint24(v[10:13]))
    brs.D5[idx] = int(utils.GetUint24(v[13:16]))
    brs.D1[idx] = int(utils.GetUint24(v[16:]))
    */
  }

  return brs, nil
}
