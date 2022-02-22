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


type Sync struct {
	Servant
}

type SyncOutput struct {
  // Array of batches of 25 tickets
  Details [][]string `json:"details,omitempty"`
}

type SyncResult struct {
}

func NewSync(progressChannel chan interface{}) *Sync {
	return &Sync{
		*NewServant(progressChannel),
	}
}

func (v *Sync) GetStrideSize() int {
  // packetSize = 1024 - header(22) - challenge(16) - signature(2)
  coinsLen := 1024 - 22 - 16 - 2


  // owCbyte
  coinsLen -= 3

  // owAn
  coinsLen -= 16

  // Sn + An
  coinItemSize := 3 
  
  return coinsLen / coinItemSize
}

func (v *Sync) Sync(ctx context.Context, ridx int, ownerCC cloudcoin.CloudCoin, coins []cloudcoin.CloudCoin, add bool) (*SyncOutput, error) {
  stride := v.GetStrideSize()
  if add {
    logger.L(ctx).Debugf("Adding %d coins to Raida %d (%d notes per packet)", len(coins), ridx, stride)
  } else {
    logger.L(ctx).Debugf("Deleting %d coins from Raida %d (%d notes per packet)", len(coins), ridx, stride)
  }

  var do = &SyncOutput{}
  for i := 0; i < len(coins); i += stride {
    max := i + stride
    if max > len(coins) {
      max = len(coins)
    }

    logger.L(ctx).Debugf("coins slice %d:%d, Processing %d notes", i, max, (max - i))
    _, err := v.ProcessSync(ctx, ridx, ownerCC, coins[i:max], add)
    if err != nil {
      logger.L(ctx).Errorf("Failed to process batch: %s", err.Error())
      perr := err.(*perror.ProgramError)
      do.Details = append(do.Details, perr.Details)
      return nil, err
    }
  }

  return do, nil
}

func (v *Sync) ProcessSync(ctx context.Context, ridx int, ownerCC cloudcoin.CloudCoin,  ccs []cloudcoin.CloudCoin, add bool) (*SyncResult, error) {
  logger.L(ctx).Debugf("Syncing %d coins on raida %d", len(ccs), ridx)

  var command int
  if add {
    command = COMMAND_SYNC_OWNER_ADD
  } else {
    command = COMMAND_SYNC_OWNER_DELETE
  }

  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

	params := make([][]byte, v.Raida.TotalServers())
  params[ridx] = v.GetHeaderSky(uint16(command), ridx, cce)
  
  encb := make([]byte, 0)
  params[ridx] = append(params[ridx], v.GetChallenge()...)

  owCbyte := utils.ExplodeSn(ownerCC.Sn)
  owAn, _ := hex.DecodeString(ownerCC.Ans[ridx])
  encb = append(encb, owCbyte...)
  encb = append(encb, owAn...)

  for _, cc := range(ccs) {
    sn := cc.Sn
    cbyte := utils.ExplodeSn(sn)

    encb = append(encb, cbyte...)
  }

  encb = append(encb, v.GetSignature()...)
  encb, err = v.EncryptIfRequired(ctx, cce, ridx, encb)
  if err != nil {
    logger.L(ctx).Debugf("Failed to encrypt body for R%d: %s", ridx, err.Error())
    return nil, err
  }
  params[ridx] = append(params[ridx], encb...)

  

  v.UpdateHeaderUdpPackets(params)
	results := v.Raida.SendRequestToSpecificRS(ctx, params, 0, []int{ridx})
  pownArray, _ := v.ProcessGenericResponsesCommon(ctx, nil, results, v.SyncSuccessFunction, cce)

  status := pownArray[ridx]
  logger.L(ctx).Debugf("Status %d", status)

  if status != config.RAIDA_STATUS_PASS {
    logger.L(ctx).Errorf("Failed to sync coins on raida %d. Status %d", ridx, status)
    return nil, perror.New(perror.ERROR_FAILED_TO_FIX, "Invalid status from RAIDA")
  }

  fr := &SyncResult{}

  return fr, nil
}

func (v *Sync) SyncSuccessFunction(ctx context.Context, idx int, status int, coins []cloudcoin.CloudCoin, rdata []byte) (int, interface{}) {
  if (status == RESPONSE_STATUS_WORKING) {
    logger.L(ctx).Debugf("Raida%d Got request %d", idx, status)
    
    return config.RAIDA_STATUS_PASS, nil
  } 

  logger.L(ctx).Debugf("Raida%d Got unknown response for sync status=%d", idx, status)
  
  return config.RAIDA_STATUS_ERROR, nil
}
