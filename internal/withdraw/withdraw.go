package withdraw

import (
	"context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/skywallet"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/transactions"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/worker"
)

type Withdraw struct {
  worker.Worker
  progressChannel chan interface{}
  Wallet *wallets.Wallet
  Task *tasks.Task
}

type WithdrawResult struct {
  TotalCoins int
  TotalAuthentic int
  TotalFracked int
  TotalCounterfeit int
  TotalLimbo int
  Details [][]string
  Coins []raida.CoinOutput
}

func New(progressChannel chan interface{}, task *tasks.Task) (*Withdraw) {
  return &Withdraw{
		*worker.New(progressChannel),
    progressChannel,
    nil,
    task,
  }

}

func (v *Withdraw) BatchFunction(ctx context.Context, coins []cloudcoin.CloudCoin) error {
  logger.L(ctx).Debugf("Calling batch function for withdraw")

  err := storage.GetDriver().UpdateStatusForNewCoin(ctx, v.Wallet, coins)
  if err != nil {
    return err
  }

  return nil
}

func (v *Withdraw) Withdraw(ctx context.Context, srcWallet *wallets.SkyWallet, dstWallet *wallets.Wallet, amount int, tag string) (*WithdrawResult, error) {
  logger.L(ctx).Debugf("Withdrawring %d coins from %s to %s", amount, srcWallet.Name, dstWallet.Name)

  v.Wallet = dstWallet
  if (amount > srcWallet.Balance) {
    logger.L(ctx).Debugf("Not enough coins. Balance is %d", srcWallet.Balance)
    return nil, perror.New(perror.ERROR_NOT_ENOUGH_COINS, "Not enough CloudCoins")
  }

  coins, err := v.GetSkyCoinsToDealWith(ctx, srcWallet, amount)
  if err != nil {
    logger.L(ctx).Errorf("Failed to pick coins: %s", err.Error())
    return nil, err
  }

  logger.L(ctx).Debugf("Got %d coins", len(coins))

  receiptID, _ := utils.GenerateReceiptID()
  w := raida.NewWithdraw(v.progressChannel)
  if v.Task != nil {
    iterations := utils.GetTotalIterations(len(coins), w.GetStrideSize())
    v.Task.SetTotalIterations(iterations)
    v.Task.Progress = 0
    v.Task.Message = "Withdrawing Coins"
  }
  w.SetBatchFunction(v.BatchFunction)
  result, err := w.Withdraw(ctx, srcWallet.IDCoin, coins, receiptID, tag)
  if err != nil {
    logger.L(ctx).Errorf("Withdraw failed")
    return nil, err
  }

  sw := skywallet.New(nil)
  r := &wallets.StatementTransaction{}
  r.ID = receiptID
  r.To = dstWallet.Name
  r.Type = "Withdraw"
  r.Details = make([]wallets.StatementDetail, 0)
  for _, cc := range(coins) {
    cc.Grade()
    d := wallets.StatementDetail{}
    d.PownString = cc.GetPownString()
    d.Result = cc.GetGradeStatusString()
    d.Sn = cc.Sn

    r.Details = append(r.Details, d)
  }
  sw.AddReceipt(ctx, srcWallet, r)

  withdrawResult := &WithdrawResult{}
  withdrawResult.TotalCoins = result.TotalCoins
  withdrawResult.TotalCounterfeit = result.TotalCounterfeit
  withdrawResult.TotalAuthentic = result.TotalAuthentic
  withdrawResult.TotalFracked = result.TotalFracked
  withdrawResult.TotalLimbo = result.TotalLimbo
  withdrawResult.Details = result.Details
  withdrawResult.Coins = result.Coins


  // Adding Transaction

  t := transactions.New(amount, "Withdrawn from " + srcWallet.Name + ": " + tag, "Withdraw", receiptID)
  for _, cc := range(coins) {
    cc.Grade()
    t.AddDetail(cc.Sn, cc.GetPownString(), cc.GetGradeStatusString())
  }
  err = storage.GetDriver().AppendTransaction(ctx, dstWallet, t)
  if err != nil {
    logger.L(ctx).Warnf("Failed to save transaction for wallet %s:%s", srcWallet.Name, err.Error())
    return nil, perror.New(perror.ERROR_SAVE_TRANSACTION, "Failed to save transaction: " + err.Error())
  }

  return withdrawResult, nil
}

