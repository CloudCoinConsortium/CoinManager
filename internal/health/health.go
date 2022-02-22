package health

import (
	"context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/worker"
)

type Healther struct {
  worker.Worker
  progressChannel chan interface{}
  Wallet *wallets.Wallet
  Task *tasks.Task
}

type HealthResult struct {
  TotalCoins int
  TotalCheckedCoins int
  TotalAuthentic int
  TotalAlreadyFracked int
  TotalFracked int
  TotalCounterfeit int
  TotalLimbo int
  TotalErrors int
  Details [][]string
  Coins []raida.CoinOutput
}

func New(progressChannel chan interface{}, task *tasks.Task) (*Healther, error) {
  return &Healther{
    *worker.New(progressChannel),
    progressChannel,
    nil,
    task,
  }, nil

}

func (v *Healther) BatchFunction(ctx context.Context, coins []cloudcoin.CloudCoin) error {
  logger.L(ctx).Debugf("Calling batch function for health")

  for idx, _ := range(coins) {
    err := storage.GetDriver().UpdateStatus(ctx, v.Wallet, coins[idx:idx + 1])
    // Error might be coin already exists. If it was authentic it stays in the Bank
    if err != nil {
      switch err.(type) {
      case *perror.ProgramError:
        perr := err.(*perror.ProgramError)
        if perr.Code == perror.ERROR_COIN_ALREADY_EXISTS {
          logger.L(ctx).Debugf("coin %d is not moved", coins[idx].Sn)
          continue
        }
      }

      logger.L(ctx).Errorf("Failed to move coin %d: %s", coins[idx].Sn, err.Error())
      continue
    }
  }

  return nil
}

  

func (v *Healther) HealthCheck(ctx context.Context, wallet *wallets.Wallet) (*HealthResult, error) {
  logger.L(ctx).Debugf("Health checking %s", wallet.Name)

  v.Wallet = wallet

  healthResult := &HealthResult{}
  coins := make([]cloudcoin.CloudCoin, 0)
  for d, cdns := range(wallet.CoinsByDenomination) {
    logger.L(ctx).Debugf("dn %d, total notes %d", d, len(cdns))
    for _, cc := range(cdns) {
      logger.L(ctx).Debugf("cc %d l=%d", cc.Sn, cc.GetLocationStatus())
      healthResult.TotalCoins += cc.GetDenomination()

      err := storage.GetDriver().ReadCoin(ctx, wallet, cc)
      if err != nil {
        logger.L(ctx).Warnf("Failed to read coin %d: %s", cc.Sn, err.Error())
        healthResult.TotalErrors += cc.GetDenomination()
        continue
      }

      if cc.GetLocationStatus() == config.COIN_LOCATION_STATUS_FRACKED {
        healthResult.TotalAlreadyFracked += cc.GetDenomination()
        continue
      }

      coins = append(coins, *cc)
      healthResult.TotalCheckedCoins += cc.GetDenomination()
    }
  }


  d := raida.NewDetect(v.progressChannel)
  if v.Task != nil {
    iterations := utils.GetTotalIterations(len(coins), d.GetStrideSize())
    v.Task.SetTotalIterations(iterations)
    v.Task.Progress = 0
    v.Task.Message = "Health Checking Coins"
  }

  d.SetBatchFunction(v.BatchFunction)
  response, err := d.Detect(ctx, coins)
  if err != nil {
    logger.L(ctx).Errorf("Failed to health coins: %s", err.Error())
    return nil, err
  }

  healthResult.TotalAuthentic = response.TotalAuthentic
  healthResult.TotalFracked = response.TotalFracked
  healthResult.TotalLimbo = response.TotalLimbo
  healthResult.TotalCounterfeit = response.TotalCounterfeit
  healthResult.Details = response.Details
  healthResult.Coins = response.Coins

  return healthResult, nil
//  return ddata, nil
}
