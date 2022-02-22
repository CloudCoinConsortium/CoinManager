package raida

import (
	"context"
	"encoding/hex"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
)


type Pang struct {
	Servant
}

type PangOutput struct {
  TotalCoins int `json:"total"`
  TotalAuthentic int `json:"authentic"`
  TotalFracked int `json:"fracked"`
  TotalLimbo int `json:"limbo"`
  TotalCounterfeit int `json:"counterfeit"`
  TotalUnknown int `json:"unknown"`
  TotalAlreadyExists int `json:"already_exists"`
  Coins []CoinOutput `json:"coins"`
  Tickets [][]string `json:"tickets"`
  Details [][]string `json:"details"`
}

func NewPang(progressChannel chan interface{}) *Pang {
	return &Pang{
		*NewServant(progressChannel),
	}
}

func (v *Pang) GetStrideSize() int {
  // packetSize = 1024 - header(22) - challenge(16) - signature(2) - pang
  coinsLen := 1024 - 22 - 16 - 2 - 16

  // Sn + An + Pan
  coinItemSize := 3 + 16

  return coinsLen / coinItemSize
}

func (v *Pang) Pang(ctx context.Context, coins []cloudcoin.CloudCoin) (*PangOutput, error) {
  stride := v.GetStrideSize()
	logger.L(ctx).Debugf("Panging %d notes (%d notes per packet)", len(coins), stride)

  var po = &PangOutput{}
  po.Tickets = make([][]string, 0)
  po.Details = make([][]string, 0)
  for i := 0; i < len(coins); i += stride {
    max := i + stride
    if max > len(coins) {
      max = len(coins)
    }

    logger.L(ctx).Debugf("coins slice %d:%d, Processing %d notes", i, max, (max - i))
    response, err := v.ProcessPang(ctx, coins[i:max])
    if err != nil {
      logger.L(ctx).Errorf("Failed to process batch: %s", err.Error())
      perr := err.(*perror.ProgramError)
      po.Details = append(po.Details, perr.Details)
    }

    logger.L(ctx).Debugf("Batch Function %v", v.batchFunction)
    if (v.batchFunction != nil) {
      err := v.batchFunction(ctx, coins[i:max])
      if err != nil {
        logger.L(ctx).Errorf("Failed to call batch function:" + err.Error())
        for _, lcc := range(coins[i:max]) {
          po.TotalUnknown += lcc.GetDenomination()
        }

        errs := v.GetProgramErrors(err)
        po.Details = append(po.Details, errs)
      }
    }

    po.TotalAuthentic += response.TotalAuthentic
    po.TotalCounterfeit += response.TotalCounterfeit
    po.TotalFracked += response.TotalFracked
    po.TotalLimbo += response.TotalLimbo
    po.TotalUnknown += response.TotalUnknown
    po.Coins = append(po.Coins, response.Coins...)
    po.Tickets = append(po.Tickets, response.Tickets[0])

  }

  for _, cc:= range(coins) {
    po.TotalCoins += cc.GetDenomination()
  }

  return po, nil

}

func (v *Pang) ProcessPang(ctx context.Context, coins []cloudcoin.CloudCoin) (*PangOutput, error) {
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

  pgs := make([]string, 0)
  for i := 0; i < config.TOTAL_RAIDA_NUMBER; i++ {
    pg, _ := utils.GeneratePG()
    pgs = append(pgs, pg)
  }

  // Iterate over raida servers
	for idx, _ := range params {
    if isSky {
    	params[idx] = v.GetHeaderSky(COMMAND_PANG, idx, cce)
    } else {
    	params[idx] = v.GetHeader(COMMAND_PANG, idx, cce)
    }
    encb := make([]byte, 0)
    encb = append(encb, v.GetChallenge()...)

    pg, _ := hex.DecodeString(pgs[idx])

    logger.L(ctx).Debugf("r%d pgggg %s", idx, pgs[idx])

    encb = append(encb, pg...)


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
  pownArray, tickets := v.ProcessGenericResponsesCommon(ctx, coins, results, v.MixedPangSuccessFunction, cce)

  err = nil
  if !v.IsQuorumCollected(pownArray) {
    details := v.GetErrorsFromResults(results)
    logger.L(ctx).Errorf("Not enough valid responses. Quorum is not reached")
    err = perror.NewWithDetails(perror.ERROR_RAIDA_QUORUM, "Not enough valid responses from the RAIDA servers", details)
  }
  

  logger.L(ctx).Debug("Pang Results")
  var a, f, c, u, l, total int
  coinResults := make([]CoinOutput, len(coins))
  for idx, cc := range (coins) {
    coins[idx].Grade()
//    coins[idx].SetAnsToPansIfPassed()


		for ridx := 0; ridx < config.TOTAL_RAIDA_NUMBER; ridx++ {
			if coins[idx].Statuses[ridx] == config.RAIDA_STATUS_PASS {
        h := strconv.Itoa(ridx) + strconv.Itoa(int(coins[idx].Sn)) + pgs[ridx]
        an := cloudcoin.GetANFromPG(ridx, coins[idx].Sn, pgs[ridx])

        logger.L(ctx).Debugf("Setting an%d for sn%d to %s h=%s", ridx, coins[idx].Sn, an, h)
        coins[idx].SetAn(ridx, an)
      }
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

  pr := &PangOutput{
    TotalCoins: total,
    TotalAuthentic: a,
    TotalFracked: f,
    TotalCounterfeit: c,
    TotalLimbo: l,
    TotalUnknown: u,
    Tickets: [][]string{ sTickets },
  }

  pr.Coins = coinResults

  return pr, err
}


func (v *Pang) ReadTicket(ctx context.Context, idx int, data []byte) (int, []byte) {
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

func (v *Pang) MixedPangSuccessFunction(ctx context.Context, idx int, status int, coins []cloudcoin.CloudCoin, rdata []byte) (int, interface{}) {
  if (status == RESPONSE_STATUS_ALL_PASS) {
    logger.L(ctx).Debugf("Raida%d AllPass Status %d", idx, status)
    
    status, ticket := v.ReadTicket(ctx, idx, rdata[config.RAIDA_REQUEST_HEADER_SIZE:])
    if status != config.RAIDA_STATUS_PASS {
      logger.L(ctx).Errorf("Failed to read ticket for mixed results. Raida %d", idx)
      v.SetCoinsStatus(coins, idx, status)
      return status, nil
    }

    logger.L(ctx).Debugf("Got ticket %v", ticket)

    v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_PASS)
    return config.RAIDA_STATUS_PASS, ticket
  } 
  
  if (status == RESPONSE_STATUS_MIX) {
    logger.L(ctx).Debugf("Raida%d Mix Status %d. Reading ticket firts", idx, status)

    status, ticket := v.ReadTicket(ctx, idx, rdata[config.RAIDA_REQUEST_HEADER_SIZE:])
    if status != config.RAIDA_STATUS_PASS {
      logger.L(ctx).Errorf("Failed to read ticket for mixed results. Raida %d", idx)
      v.SetCoinsStatus(coins, idx, status)
      return status, nil
    }

    logger.L(ctx).Debugf("Got ticket %v", ticket)

    result := v.ReadMixedResultsAndUpdateCoins(ctx, idx, coins, rdata[config.RAIDA_RESPONSE_HEADER_SIZE + 4:])

    return result, ticket
  }

  if (status == RESPONSE_STATUS_ALL_FAIL) {
    logger.L(ctx).Errorf("Raida%d AllFail Status %d", idx, status)

    // The coins themselves are failed
    v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_FAIL)

    // It is ok have 'PASS' there. It is not about the coins. This status is about RAIDA response itself
    return config.RAIDA_STATUS_PASS, nil
  } 
  
  logger.L(ctx).Errorf("Raida%d Invalid Status %d", idx, status)
  v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_ERROR)

  return config.RAIDA_STATUS_ERROR, nil
}
