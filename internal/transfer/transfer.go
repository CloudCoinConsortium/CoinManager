package transfer

import (
	"context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/transactions"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/worker"
)

type Transfer struct {
  worker.Worker
  progressChannel chan interface{}
}

type TransferResult struct {
  Amount int `json:"amount"`
}

func New(progressChannel chan interface{}) (*Transfer) {
  return &Transfer{
		*worker.New(progressChannel),
    progressChannel,
  }

}

func (v *Transfer) Transfer(ctx context.Context, srcWallet, dstWallet *wallets.Wallet, amount int, tag string) (*TransferResult, error) {
  logger.L(ctx).Debugf("Transferring %d coins from %s to %s", amount, srcWallet.Name, dstWallet.Name)

  if (amount > srcWallet.Balance) {
    logger.L(ctx).Debugf("Not enough coins. Balance is %d", srcWallet.Balance)
    return nil, perror.New(perror.ERROR_NOT_ENOUGH_COINS, "Not enough CloudCoins")
  }

  coinsToSend, err := v.GetCoinsToDealWithNoRead(ctx, srcWallet, amount)
  if err != nil {
    logger.L(ctx).Errorf("Failed to pick coins to send: %s", err.Error())
    return nil, err
  }

  sentAmount, err := storage.GetDriver().MoveCoins(ctx, srcWallet, dstWallet, coinsToSend)
  if err != nil {
    logger.L(ctx).Errorf("Failed to move coins: %s", err.Error())
    return nil, perror.New(perror.ERROR_FAILED_TO_MOVE_COINS, "Failed to move coins: " + err.Error())
  }

  logger.L(ctx).Debugf("totalSent %d", sentAmount)


  // Adding Transaction
  receiptID, _ := utils.GenerateReceiptID()

  t := transactions.New(-1 * amount, "Transfer to " + dstWallet.Name + ": " + tag, "Local Transfer", receiptID)
  for _, cc := range(coinsToSend) {
    cc.Grade()
    t.AddDetail(cc.Sn, cc.GetPownString(), cc.GetGradeStatusString())
  }
  err = storage.GetDriver().AppendTransaction(ctx, srcWallet, t)
  if err != nil {
    logger.L(ctx).Warnf("Failed to save transaction for wallet %s:%s", srcWallet.Name, err.Error())
    return nil, perror.New(perror.ERROR_SAVE_TRANSACTION, "Failed to save transaction: " + err.Error())
  }

  t = transactions.New(amount, "Transfer from " + srcWallet.Name + ": " + tag, "Local Transfer", receiptID)
  for _, cc := range(coinsToSend) {
    cc.Grade()
    t.AddDetail(cc.Sn, cc.GetPownString(), cc.GetGradeStatusString())
  }
  err = storage.GetDriver().AppendTransaction(ctx, dstWallet, t)
  if err != nil {
    logger.L(ctx).Warnf("Failed to save transaction for wallet %s:%s", dstWallet.Name, err.Error())
    return nil, perror.New(perror.ERROR_SAVE_TRANSACTION, "Failed to save transaction: " + err.Error())
  }


  tr := &TransferResult{
    Amount: sentAmount,
  }

  return tr, nil
}

