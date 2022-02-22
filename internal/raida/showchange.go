package raida

import (
	"context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
)


type ShowChange struct {
	Servant
}

type ShowChangeOutput struct {
  SerialNumbers []uint32 `json:"sns"`
}

func NewShowChange(progressChannel chan interface{}) *ShowChange {
	return &ShowChange{
		*NewServant(progressChannel),
	}
}

func (v *ShowChange) ShowChange(ctx context.Context, denomination int) (*ShowChangeOutput, error) {
	logger.L(ctx).Debugf("ShowChangeing for denomination %d ", denomination)

  so := &ShowChangeOutput{}
	params := make([][]byte, v.Raida.TotalServers())
  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

  psn := uint32(config.PUBLIC_CHANGE_SN)
  // Iterate over raida servers
	for idx, _ := range params {
  	params[idx] = v.GetHeaderSky(COMMAND_SHOW_CHANGE, idx, cce)
    encb := make([]byte, 0)
    encb = append(encb, v.GetChallenge()...)

    cbytes := utils.ExplodeSn(psn)

    encb = append(encb, cbytes...)
    encb = append(encb, byte(denomination))
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
  pownArray, rdata := v.ProcessGenericResponsesCommon(ctx, nil, results, v.CommonSuccessFunction, cce)

  if !v.IsQuorumCollected(pownArray) {
    details := v.GetErrorsFromResults(results)
    logger.L(ctx).Errorf("Not enough valid responses. Quorum is not reached")
    return nil, perror.NewWithDetails(perror.ERROR_RAIDA_QUORUM, "Not enough valid responses from the RAIDA servers", details)
  }


  sns := make(map[int]int, 0)
  var sn int
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
      sn = int(bitem[i]) << 16 | int(bitem[i + 1]) << 8 | int(bitem[i + 2])

      logger.L(ctx).Debugf("RAIDA%d SN %d %d,%d,%d", idx, sn, bitem[i], bitem[i+1], bitem[i+2])

      sns[sn]++
    }
  }

  for sn, nums := range(sns) {
    if nums < config.MIN_QUORUM_COUNT {
      continue
    }

    logger.L(ctx).Debugf("Got SN %d", sn)
    so.SerialNumbers = append(so.SerialNumbers, uint32(sn))
  }

  logger.L(ctx).Debugf("d %v", sns)

  return so, nil
}

