package skytransfer

import (
	"context"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/dnsservice"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/skywallet"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/worker"
)

type SkyTransferer struct {
  worker.Worker
  progressChannel chan interface{}
  Task *tasks.Task
  SkyWallet *wallets.SkyWallet
}

type SkyTransferResult struct {
  TotalCoins int
  TotalAuthentic int
  TotalFracked int
  TotalCounterfeit int
  TotalLimbo int
  Details [][]string
  Coins []raida.CoinOutput
}

func New(progressChannel chan interface{}, task *tasks.Task) (*SkyTransferer) {
  return &SkyTransferer{
    *worker.New(progressChannel),
    progressChannel,
    task,
    nil,
  }

}
func (v *SkyTransferer) SkyTransfer(ctx context.Context, skyWallet *wallets.SkyWallet, dstName string, tosn uint32, amount int, tag string) (*SkyTransferResult, error) {
  logger.L(ctx).Debugf("SkyTransfering %d coins from %s to %s (%d)", amount, skyWallet.Name, dstName, tosn)

  if (amount > skyWallet.Balance) {
    logger.L(ctx).Debugf("Not enough coins")
    return nil, perror.New(perror.ERROR_NOT_ENOUGH_COINS, "Not enough CloudCoins")
  }

  var to uint32
  var toStr string 
  if tosn != 0 {
    to = tosn
    toStr = "SN " + strconv.Itoa(int(to))
  } else {
    var err error
    logger.L(ctx).Debugf("Resolving %s", dstName)
    dns := dnsservice.New(nil)
    to, err = dns.GetSN(ctx, dstName)
    if err != nil || to == 0 {
      logger.L(ctx).Errorf("Failed to resolve dst wallet %s", dstName)
      return nil, perror.New(perror.ERROR_DNS_RESOLVE, "Failed to find SkyVault address")
    }

    toStr = dstName
    logger.L(ctx).Debugf("%s resolved to %d", dstName, to)
  }

  v.SetBreakInBankCallback(v.BreakInBankCallback)

  coins, err := v.GetSkyCoinsToDealWith(ctx, skyWallet, amount)
  if err != nil {
    logger.L(ctx).Errorf("Failed to pick coins: %s", err.Error())
    return nil, err
  }

  logger.L(ctx).Debugf("coins %v", coins)





  receiptID, _ := utils.GenerateReceiptID()

  t := raida.NewTransfer(v.progressChannel)
  if v.Task != nil {
    iterations := utils.GetTotalIterations(len(coins), t.GetStrideSize())
    v.Task.SetTotalIterations(iterations)
    v.Task.Progress = 0
    v.Task.Message = "Transferring Coins"
  }
  result, err := t.Transfer(ctx, skyWallet.IDCoin, coins, to, receiptID, tag)
  if err != nil {
    logger.L(ctx).Errorf("Transfer failed")
    return nil, err
  }

  sw := skywallet.New(nil)
  r := &wallets.StatementTransaction{}
  r.ID = receiptID
  r.To = toStr
  r.Type = "Transfer"
  r.Details = make([]wallets.StatementDetail, 0)
  for _, cc := range(coins) {
    cc.Grade()
    d := wallets.StatementDetail{}
    d.PownString = cc.GetPownString()
    d.Result = cc.GetGradeStatusString()
    d.Sn = cc.Sn

    r.Details = append(r.Details, d)
  }
  sw.AddReceipt(ctx, skyWallet, r)

  err = sw.AppendSenderHistory(ctx, dstName)
  if err != nil {
    logger.L(ctx).Warnf("Failed to append sender history: %s", err.Error())
  }

  skyTransferResult := &SkyTransferResult{}
  skyTransferResult.TotalCoins = result.TotalCoins
  skyTransferResult.TotalCounterfeit = result.TotalCounterfeit
  skyTransferResult.TotalAuthentic = result.TotalAuthentic
  skyTransferResult.TotalFracked = result.TotalFracked
  skyTransferResult.TotalLimbo = result.TotalLimbo
  skyTransferResult.Details = result.Details
  skyTransferResult.Coins = result.Coins

  return skyTransferResult, nil
}

// Update SkyWallet Balance
func (v *SkyTransferer) BreakInBankCallback(ctx context.Context, wallet *wallets.SkyWallet) error {
  logger.L(ctx).Debugf("Updating skywallet %s balance", wallet.Name)

  skywalleter := skywallet.New(nil)
  err := skywalleter.Update(ctx, wallet)
  if err != nil {
    return err
  }

  return nil
}


/*
*/
  
