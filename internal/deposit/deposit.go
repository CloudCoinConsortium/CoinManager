package deposit

import (
	"context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/skywallet"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/transactions"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/worker"
)

type Depositer struct {
  worker.Worker
  progressChannel chan interface{}
  Wallet *wallets.Wallet
}

type DepositResult struct {
  TotalCoins int
  TotalAuthentic int
  TotalFracked int
  TotalCounterfeit int
  TotalLimbo int
  Details [][]string
  Coins []raida.CoinOutput
}

func New(progressChannel chan interface{}) (*Depositer, error) {
  return &Depositer{
    *worker.New(progressChannel),
    progressChannel,
    nil,
  }, nil

}

func (v *Depositer) BatchFunction(ctx context.Context, coins []cloudcoin.CloudCoin) error {
  logger.L(ctx).Debugf("Calling batch function for deposit")

  var err error
  for idx, cc := range(coins) {
    isAuthentic, _, isCounterfeit := cc.IsAuthentic()
    if isAuthentic {
      logger.L(ctx).Debugf("Moving cc %d to sent", cc.Sn)
      err = storage.GetDriver().SetLocation(ctx, v.Wallet, []cloudcoin.CloudCoin{coins[idx]}, config.COIN_LOCATION_STATUS_SENT)
    } else if isCounterfeit {
      logger.L(ctx).Debugf("Moving cc %d to counterfeit", cc.Sn)
      err = storage.GetDriver().SetLocation(ctx, v.Wallet, []cloudcoin.CloudCoin{coins[idx]}, config.COIN_LOCATION_STATUS_COUNTERFEIT)
    } else {
      err = storage.GetDriver().SetLocation(ctx, v.Wallet, []cloudcoin.CloudCoin{coins[idx]}, config.COIN_LOCATION_STATUS_LIMBO)
    }

    if err != nil {
      return err
    }
  }

  return nil
}

  

func (v *Depositer) Deposit(ctx context.Context, wallet *wallets.Wallet, to uint32, toName string, amount int, tag string) (*DepositResult, error) {
  logger.L(ctx).Debugf("Depositing %d coins from %s to SN%d", amount, wallet.Name, to)

  v.Wallet = wallet
  if (amount > wallet.Balance) {
    logger.L(ctx).Debugf("Not enough coins")
    return nil, perror.New(perror.ERROR_NOT_ENOUGH_COINS, "Not enough CloudCoins")
  }

  logger.L(ctx).Debugf("Depositing %d CC to sn#%d from %s", amount, to, wallet.Name)

  coinsToDeposit, err := v.GetCoinsToDealWith(ctx, wallet, amount)
  if err != nil {
    logger.L(ctx).Errorf("Failed to pick coins to deposit: %s", err.Error())
    return nil, err
  }

  receiptID, _ := utils.GenerateReceiptID()
  d := raida.NewDeposit(v.progressChannel)
  d.SetBatchFunction(v.BatchFunction)

  response, err := d.Deposit(ctx, coinsToDeposit, to, receiptID, tag)
  if err != nil {
    logger.L(ctx).Errorf("Failed to deposit coins: %s", err.Error())
    return nil, err
  }


  depositedAmount := response.TotalAuthentic + response.TotalFracked
  if amount != depositedAmount {
    logger.L(ctx).Warnf("Not all coins were deposited: %d. Requested %d", depositedAmount, amount)
  } else {
    logger.L(ctx).Debugf("All coins deposited")
  }


  var optype string
  sw := skywallet.New(nil)
  swallet, err := sw.GetByID(ctx, to)
  if err != nil {
    logger.L(ctx).Debugf("Not our skywallet")
  } else {
    logger.L(ctx).Debugf("Our skywallet")
    logger.L(ctx).Debugf("Deposited to our wallet %s", swallet.Name)
    toName = swallet.Name
  }
  optype = "Sent"

  // Adding Transaction
  t := transactions.New(-1 * amount, "To: " + toName + "<br>" + tag, optype, receiptID)
  for _, cc := range(coinsToDeposit) {
    cc.Grade()
    t.AddDetail(cc.Sn, cc.GetPownString(), cc.GetGradeStatusString())
  }
  err = storage.GetDriver().AppendTransaction(ctx, wallet, t)
  if err != nil {
    logger.L(ctx).Warnf("Failed to save transaction for wallet %s:%s", wallet.Name, err.Error())
    return nil, perror.New(perror.ERROR_SAVE_TRANSACTION, "Failed to save transaction: " + err.Error())
  }

  if toName != "" {
    err = sw.AppendSenderHistory(ctx, toName)
    if err != nil {
      logger.L(ctx).Warnf("Failed to append sender history: %s", err.Error())
    }
  }

  r := &wallets.StatementTransaction{}
  r.ID = receiptID
  r.From = wallet.Name
  r.Type = "Deposit"
  r.Details = make([]wallets.StatementDetail, 0)
  for _, cc := range(coinsToDeposit) {
    // Graded already above
    d := wallets.StatementDetail{}
    d.PownString = cc.GetPownString()
    d.Result = cc.GetGradeStatusString()
    d.Sn = cc.Sn

    r.Details = append(r.Details, d)
  }
  sw.AddReceipt(ctx, swallet, r)

  depositResult := &DepositResult{}
  depositResult.TotalCoins = response.TotalCoins
  depositResult.TotalCounterfeit = response.TotalCounterfeit
  depositResult.TotalAuthentic = response.TotalAuthentic
  depositResult.TotalFracked = response.TotalFracked
  depositResult.TotalLimbo = response.TotalLimbo
  depositResult.Details = response.Details
  depositResult.Coins = response.Coins

  return depositResult, nil
//  return ddata, nil
}
