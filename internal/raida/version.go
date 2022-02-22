package raida

import (

	//"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"

	"context"
	"strconv"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
)

type Version struct {
	Servant
}

type VersionOutput struct {
  RawVersions []string `json:"raw_versions"`
  Versions []string `json:"versions"`
}

func NewVersion(progressChannel chan interface{}) *Version {
	return &Version{
		*NewServant(progressChannel),
	}
}

func (v *Version) Version(ctx context.Context) (*VersionOutput, error) {
	logger.L(ctx).Debug("Getting SuperRAIDA Version")

	params := make([][]byte, v.Raida.TotalServers())
	for idx, _ := range params {
  	params[idx] = v.GetHeader(COMMAND_VERSION, idx, nil)
    params[idx] = append(params[idx], 0x3e, 0x3e)
	}

  v.UpdateHeaderUdpPackets(params)
	results := v.Raida.SendRequest(ctx, params, 0)
  pownArray, rdata := v.ProcessGenericResponses(ctx, nil, results, v.VersionSuccessFunction)

  er := &VersionOutput{}
  er.RawVersions = make([]string, config.TOTAL_RAIDA_NUMBER)
  er.Versions = make([]string, config.TOTAL_RAIDA_NUMBER)

  for idx, status := range(pownArray) {
    if status == config.RAIDA_STATUS_PASS {
      bytes, ok := rdata[idx].([]byte)
      if ok {
        er.RawVersions[idx] = string(bytes)
        tu, err := strconv.ParseInt(er.RawVersions[idx], 10, 64)
		    if err != nil {
          er.Versions[idx] = "error"
		    } else {
          t := time.Unix(tu, 0)
          er.Versions[idx] = t.Format(time.RFC3339)
        }
      } else {
        er.RawVersions[idx] = "error"
        er.Versions[idx] = "error"
      }
    } else {
      er.RawVersions[idx] = ""
      er.Versions[idx] = ""
    }
  }

  // To Log it
  v.GetPownStringFromStatusArray(pownArray)
  


  return er, nil
//  v.SendResult(er)
}

func (s *Servant) VersionSuccessFunction(ctx context.Context, idx int, status int, coins []cloudcoin.CloudCoin, rdata []byte) (int, interface{}) {
  logger.L(ctx).Debugf("Raida%d Version Command Response:  %v", idx, rdata)

  if (status == RESPONSE_STATUS_FAIL) {
    logger.L(ctx).Errorf("Raida%d Version Fail Status", idx)
    return config.RAIDA_STATUS_FAIL, nil
  }

  if (status != RESPONSE_STATUS_SUCCESS) {
    logger.L(ctx).Errorf("Raida%d Invalid Status %d", idx, status)
    return config.RAIDA_STATUS_ERROR, nil
  }

  return config.RAIDA_STATUS_PASS, rdata[config.RAIDA_REQUEST_HEADER_SIZE:]
}

