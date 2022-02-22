package skyhealth

import (
	"context"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
)

type SyncResult struct {
  TotalCoins int
  TotalValue int
  TotalValueAdded int
  TotalValueDeleted int
  TotalValueSkipped int
  TotalValueUnknown int
  FailedRequests int
}

func (v *SkyHealth) Sync(ctx context.Context, wallet *wallets.SkyWallet, items map[int][]int) (*SyncResult, error) {
  logger.L(ctx).Debugf("Syncing %s", wallet.Name)

  syncResult := &SyncResult{}
  quorum := (config.TOTAL_RAIDA_NUMBER / 2) + 1

  var coinsPerRaida [config.TOTAL_RAIDA_NUMBER][2][]cloudcoin.CloudCoin
  for sn, results := range(items) {
    logger.L(ctx).Debugf("sn %d v %v",results, sn)

    cc := cloudcoin.NewFromData(uint32(sn))
    var n, p, np int
    for _, ival := range(results) {
      switch ival {
      case config.HEALTH_CHECK_STATUS_NETWORK:
        n++
      case config.HEALTH_CHECK_STATUS_NOT_PRESENT:
        np++
      case config.HEALTH_CHECK_STATUS_PRESENT:
        p++
      }
    }

    for idx, ival := range(results) {
      if p >= quorum && np > 0 {
        if ival == config.HEALTH_CHECK_STATUS_NOT_PRESENT {
          logger.L(ctx).Debugf("Coin %d will be added to raida %d (exists %d, notexist %d)", sn, idx, p, np)
          coinsPerRaida[idx][0] = append(coinsPerRaida[idx][0], *cc)
          syncResult.TotalValueAdded += cc.GetDenomination()
        } else {
          syncResult.TotalValueSkipped += cc.GetDenomination()
        }
      } else if np >= quorum && p > 0 {
        if ival == config.HEALTH_CHECK_STATUS_PRESENT {
          logger.L(ctx).Debugf("Coin %d will be deleted from raida %d (exists %d, notexist %d)", sn, idx, p, np)
          coinsPerRaida[idx][1] = append(coinsPerRaida[idx][1], *cc)
          syncResult.TotalValueDeleted += cc.GetDenomination()
        } else {
          syncResult.TotalValueSkipped += cc.GetDenomination()
        }
      } else {
        logger.L(ctx).Debugf("Coin %d can't be synced on raida %d. Qurum is not reached (exists %d, notexist %d)", sn, idx, p, np)
        syncResult.TotalValueUnknown += cc.GetDenomination()
      }

      syncResult.TotalValue += cc.GetDenomination()
    }

    syncResult.TotalCoins += cc.GetDenomination()
  }


  syncer := raida.NewSync(v.progressChannel)
  iterations := 0
  if v.Task != nil {
    for _, carr := range(coinsPerRaida) {
      iadd := utils.GetTotalIterations(len(carr[0]), syncer.GetStrideSize())
      idel := utils.GetTotalIterations(len(carr[1]), syncer.GetStrideSize())

      iterations += iadd
      iterations += idel
    }

    v.Task.SetTotalIterations(iterations)
    v.Task.Progress = 0
  }
  for ridx, carr := range(coinsPerRaida) {
    if (len(carr[0]) > 0) {
      if v.Task != nil {
        v.Task.Message = "Adding coins to RAIDA " + strconv.Itoa(ridx)

        _, err := syncer.Sync(ctx, ridx, *wallet.IDCoin, carr[0], true)
        if err != nil {
          logger.L(ctx).Debugf("Sync request failed on r%d: %s", ridx, err.Error())
          syncResult.FailedRequests++
        }
      }
    }

    if (len(carr[1]) > 0) {
      if v.Task != nil {
        v.Task.Message = "Deleting coins from RAIDA " + strconv.Itoa(ridx)

        _, err := syncer.Sync(ctx, ridx, *wallet.IDCoin, carr[1], false)
        if err != nil {
          logger.L(ctx).Debugf("Sync request failed on r%d: %s", ridx, err.Error())
          syncResult.FailedRequests++
        }
      }
    }

  }

  return syncResult, nil
}
