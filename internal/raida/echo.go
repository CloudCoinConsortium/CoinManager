package raida

import (

	//"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
  "context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
)

type Echo struct {
	Servant
}

type EchoOutput struct {
  Online int `json:"online"`
  PownString string `json:"pownstring"`
  PownArray []int `json:"pownarray"`
  Latencies []int64 `json:"latencies"`
}

func NewEcho(progressChannel chan interface{}) *Echo {
	return &Echo{
		*NewServant(progressChannel),
	}
}

func (v *Echo) Echo(ctx context.Context) (*EchoOutput, error) {
	logger.L(ctx).Debug("Echoing")

	params := make([][]byte, v.Raida.TotalServers())
/*
  cce, err := v.GetEncryptionCoin()
  if err != nil {
    logger.L(ctx).Errorf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
    return nil, err
  }
*/
  
	for idx, _ := range params {
  	params[idx] = v.GetHeader(COMMAND_ECHO, idx, nil)
  	//params[idx] = v.GetHeaderEnc(COMMAND_ECHO, config.COINID_TYPE_CLOUD, cce, idx)
    params[idx] = append(params[idx], 0x3e, 0x3e)
	}

  v.UpdateHeaderUdpPackets(params)
	results := v.Raida.SendRequest(ctx, params, 0)
  pownArray, _ := v.ProcessGenericResponses(ctx, nil, results, v.EchoSuccessFunction)

  times := make([]int64, config.TOTAL_RAIDA_NUMBER)
  for idx, r := range(results) {
    times[idx] = r.RequestTime
  }

  online := 0
  for _, status := range(pownArray) {
    if status == config.RAIDA_STATUS_PASS {
      online++
    }
  }

  pownString := v.GetPownStringFromStatusArray(pownArray)

  er := &EchoOutput{}
  er.Online = online
  er.PownString = pownString
  er.PownArray = pownArray
  er.Latencies = times

  return er, nil
//  v.SendResult(er)
}

func (s *Servant) EchoSuccessFunction(ctx context.Context, idx int, status int, coins []cloudcoin.CloudCoin, rdata []byte) (int, interface{}) {
  if (status != RESPONSE_STATUS_SUCCESS) {
    logger.L(ctx).Errorf("Raida%d Invalid Status %d", idx, status)
    return config.RAIDA_STATUS_ERROR, nil
  }

  return config.RAIDA_STATUS_PASS, nil
}

