package skyhealth

import (
	"context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/worker"
)

type SkyHealth struct {
  worker.Worker
  progressChannel chan interface{}
  Wallet *wallets.Wallet
  Task *tasks.Task
}

type SkyHealthResult struct {
  SerialNumbers map[uint32][]int `json:"sns"`
  Balances []int `json:"balances"`
  QuorumCoins []uint32  `json:"quorum_sns"`
  QuorumBalance int `json:"quorum_balance"`
}

func New(progressChannel chan interface{}, task *tasks.Task) (*SkyHealth) {
  return &SkyHealth{
		*worker.New(progressChannel),
    progressChannel,
    nil,
    task,
  }

}

func (v *SkyHealth) BatchFunction(ctx context.Context, coins []cloudcoin.CloudCoin) error {
  logger.L(ctx).Debugf("Calling batch function for skyhealth")

  err := storage.GetDriver().UpdateStatusForNewCoin(ctx, v.Wallet, coins)
  if err != nil {
    return err
  }

  return nil
}

func (v *SkyHealth) GetStridesSize() int {
  sr0 := raida.NewSync(nil)

  return sr0.GetStrideSize()
}

func (v *SkyHealth) SkyHealth(ctx context.Context, wallet *wallets.SkyWallet) (*SkyHealthResult, error) {
  logger.L(ctx).Debugf("Sky Health for %s", wallet.Name)

  sr := raida.NewShowCoinsByDenomination(v.progressChannel)
  response, err := sr.ShowCoinsByDenominationRaw(ctx, wallet.IDCoin)
  if err != nil {
    logger.L(ctx).Errorf("Failed to show registry: %s", err.Error())
    return nil, err
  }

  b := raida.NewBalance(v.progressChannel)
  bresponse, err := b.BalanceRaw(ctx, wallet.IDCoin)
  if err != nil {
    logger.L(ctx).Errorf("Failed to get balance: %s", err.Error())
    return nil, perror.New(perror.ERROR_GET_BALANCE, "Failed to get balance: " + err.Error())
  }



 // v.Wallet = wallet
/*
*/
  /*
  }
*/
  skyhealthResult := &SkyHealthResult{}
  skyhealthResult.SerialNumbers = response.SerialNumbers
  skyhealthResult.QuorumBalance = wallet.Balance

  coins := make([]uint32, 0)
  for _, cdns := range(wallet.CoinsByDenomination) {
    for _, cc := range(cdns) {
      coins = append(coins, cc.Sn)
    }
  }

  skyhealthResult.QuorumCoins = coins
  skyhealthResult.Balances = bresponse.Total

  // Adding Transaction
  return skyhealthResult, nil
}

