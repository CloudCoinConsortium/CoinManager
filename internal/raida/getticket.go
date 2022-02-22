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


type GetTicket struct {
	Servant
}

type GetTicketOutput struct {
  // Array of batches of 25 tickets
  Tickets [][]string `json:"tickets"`
  Details [][]string `json:"details,omitempty"`
}

func NewGetTicket(progressChannel chan interface{}) *GetTicket {
	return &GetTicket{
		*NewServant(progressChannel),
	}
}

func (v *GetTicket) GetStrideSize() int {
  // packetSize = 1024 - header(22) - challenge(16) - signature(2)
  coinsLen := 1024 - 22 - 16 - 2

  // Sn + An
  coinItemSize := 3 + 16

  return coinsLen / coinItemSize
}


func (v *GetTicket) GetTicket(ctx context.Context, coins []cloudcoin.CloudCoin) (*GetTicketOutput, error) {
  stride := v.GetStrideSize()
	logger.L(ctx).Debugf("Getting ticketis %d notes (%d notes per packet)", len(coins), stride)

  finalResult := make([][]string, 0)
  var do = &GetTicketOutput{}
  for i := 0; i < len(coins); i += stride {
    max := i + stride
    if max > len(coins) {
      max = len(coins)
    }

    logger.L(ctx).Debugf("coins slice %d:%d, Processing %d notes", i, max, (max - i))
    response, err := v.ProcessGetTicket(ctx, coins[i:max])
    if err != nil {
      logger.L(ctx).Errorf("Failed to process batch: %s", err.Error())
      perr := err.(*perror.ProgramError)
      do.Details = append(do.Details, perr.Details)
      return nil, err
    }

    finalResult = append(finalResult, response)
  }

  do.Tickets = finalResult

  return do, nil
}

func (v *GetTicket) GetSingleTicketSky(ctx context.Context, cc *cloudcoin.CloudCoin, ridx int) (string, error) {
  logger.L(ctx).Debugf("Getting ticket for coin %d on raida %d", cc.Sn, ridx)

	params := make([][]byte, v.Raida.TotalServers())
  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

  params[ridx] = v.GetHeaderSky(COMMAND_GETTICKET, ridx, cce)
  encb := make([]byte, 0)
  encb = append(encb, v.GetChallenge()...)

  sn := cc.Sn
  cbyte := utils.ExplodeSn(sn)
  data, _ := hex.DecodeString(cc.Ans[ridx])

  encb = append(encb, cbyte...)
  encb = append(encb, data...)

  encb = append(encb, v.GetSignature()...)
  encb, err = v.EncryptIfRequired(ctx, cce, ridx, encb)
  if err != nil {
    logger.L(ctx).Debugf("Failed to encrypt body for R%d: %s", ridx, err.Error())
    return "", err
  }
  params[ridx] = append(params[ridx], encb...)

  v.UpdateHeaderUdpPackets(params)
	results := v.Raida.SendRequestToSpecificRS(ctx, params, 0, []int{ridx})
  pownArray, rdata := v.ProcessGenericResponsesCommon(ctx, nil, results, v.MixedGetTicketSuccessFunction, cce)

  status := pownArray[ridx]
  if status != config.RAIDA_STATUS_PASS {
    logger.L(ctx).Errorf("Failed to get ticket from raida %d. Status %d", ridx, status)
    return "", perror.New(perror.ERROR_GET_TICKET, "Invalid status from RAIDA")
  }

  bticket := rdata[ridx].([]byte)
  ticket := hex.EncodeToString(bticket)

  logger.L(ctx).Debugf("Got ticket %s for RAIDA%d", ticket, ridx)

  return ticket, nil
}

func (v *GetTicket) ProcessGetTicket(ctx context.Context, coins []cloudcoin.CloudCoin) ([]string, error) {

  isSky := false
  for _, cc:= range(coins) {
    logger.L(ctx).Debugf("Processing coins %d isSky (%v)",cc.Sn, cc.IsIDCoin())
    isSky = cc.IsIDCoin()
  }

  logger.L(ctx).Debugf("is sky %v", isSky)

	params := make([][]byte, v.Raida.TotalServers())
  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

  // Iterate over raida servers
	for idx, _ := range params {
    if isSky {
    	params[idx] = v.GetHeaderSky(COMMAND_GETTICKET, idx, cce)
    } else {
    	params[idx] = v.GetHeader(COMMAND_GETTICKET, idx, cce)
    }

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
  pownArray, rdata := v.ProcessGenericResponsesCommon(ctx, coins, results, v.MixedGetTicketSuccessFunction, cce)

  err = nil
  if !v.IsQuorumCollected(pownArray) {
    details := v.GetErrorsFromResults(results)
    logger.L(ctx).Errorf("Not enough valid responses. Quorum is not reached")
    err = perror.NewWithDetails(perror.ERROR_RAIDA_QUORUM, "Not enough valid responses from the RAIDA servers", details)
    return nil, err
  }

  rv := make([]string, config.TOTAL_RAIDA_NUMBER)
  for idx, v := range(rdata) {
    if v == nil {
      continue
    }

    bytes := v.([]byte)
    rv[idx] = hex.EncodeToString(bytes)
  }

  logger.L(ctx).Debug("GetTicket Results")

  logger.L(ctx).Debugf("v %v", rv)

  return rv, nil
}

func (v *GetTicket) ReadTicket(ctx context.Context, idx int, data []byte) (int, []byte) {
  needMoreBytes := 4
  receivedExtraBytes := len(data) 

  logger.L(ctx).Debugf("Ticket. Need more %d bytes to process. Received %d extra bytes", needMoreBytes, receivedExtraBytes)

  if receivedExtraBytes < needMoreBytes {
    diffBytes := needMoreBytes - receivedExtraBytes  
    logger.L(ctx).Debugf("Will download more %d bytes", diffBytes)

    result := v.Raida.ReadBytesFromDA(ctx, idx, diffBytes)
    if result.ErrCode == config.REMOTE_RESULT_ERROR_NONE {
      logger.L(ctx).Debugf("Read Ticket bytes: %d", len(result.Data))
      data = append(data, result.Data...)
    } else if result.ErrCode == config.REMOTE_RESULT_ERROR_TIMEOUT {
      logger.L(ctx).Errorf("Timeout while reading ticket bytes")
      return config.RAIDA_STATUS_NORESPONSE, []byte{}
    } else {
      logger.L(ctx).Errorf("Error while reading ticket bytes" )
      return config.RAIDA_STATUS_ERROR, []byte{}
    }
  }
    
  logger.L(ctx).Debugf("Processing extra data %d bytes", len(data))
  if (len(data) < 4) {
    logger.L(ctx).Errorf("Invalid ticket length %d", len(data))
    return config.RAIDA_STATUS_ERROR, []byte{}
  }

  return config.RAIDA_STATUS_PASS, data[:4]
}

func (v *GetTicket) MixedGetTicketSuccessFunction(ctx context.Context, idx int, status int, coins []cloudcoin.CloudCoin, rdata []byte) (int, interface{}) {
  if (status == RESPONSE_STATUS_ALL_PASS) {
    logger.L(ctx).Debugf("Raida%d AllPass Status %d", idx, status)
    
    status, ticket := v.ReadTicket(ctx, idx, rdata[config.RAIDA_REQUEST_HEADER_SIZE:])
    if status != config.RAIDA_STATUS_PASS {
      logger.L(ctx).Errorf("Failed to read ticket for mixed results. Raida %d", idx)
      return status, nil
    }

    logger.L(ctx).Debugf("Got ticket %v", ticket)
    return config.RAIDA_STATUS_PASS, ticket
  } 
  
  if (status == RESPONSE_STATUS_MIX) {
    logger.L(ctx).Debugf("Raida%d Mix Status %d. Reading ticket firts", idx, status)

    status, ticket := v.ReadTicket(ctx, idx, rdata[config.RAIDA_REQUEST_HEADER_SIZE:])
    if status != config.RAIDA_STATUS_PASS {
      logger.L(ctx).Errorf("Failed to read ticket for mixed results. Raida %d", idx)
      return status, nil
    }

    logger.L(ctx).Debugf("Got ticket %v", ticket)

    result := v.ReadMixedResultsAndUpdateCoins(ctx, idx, coins, rdata[config.RAIDA_RESPONSE_HEADER_SIZE + 4:])

    return result, ticket
  }

  if (status == RESPONSE_STATUS_ALL_FAIL) {
    logger.L(ctx).Errorf("Raida%d AllFail Status %d", idx, status)

    return config.RAIDA_STATUS_FAIL, nil
  } 
  
  logger.L(ctx).Errorf("Raida%d Invalid Status %d", idx, status)

  return config.RAIDA_STATUS_ERROR, nil
}
