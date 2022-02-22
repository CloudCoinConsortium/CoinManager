package showpayment

import (
	"context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/worker"
)


type ShowPaymenter struct {
  worker.Worker
  progressChannel chan interface{}
}

func New(progressChannel chan interface{}) (*ShowPaymenter) {
  return &ShowPaymenter{
    *worker.New(progressChannel),
    progressChannel,
  }
}


func (v *ShowPaymenter) Show(ctx context.Context, guid string) (*wallets.Statement, error) {
  logger.L(ctx).Debugf("ShowPayment for GUID %s", guid)

  sp := raida.NewShowPayment(v.progressChannel)
  response, err := sp.ShowPayment(ctx, guid)
  if err != nil {
    logger.L(ctx).Errorf("Failed to show payment: %s", err.Error())
    return nil, err
  }

  logger.L(ctx).Debugf("got %v", response)

  item := response.Item
  ttype := v.GetStatementTransactionType(item.TransactionType)

  st := wallets.Statement{
    Guid: guid,
    Type: ttype,
    Amount: item.Amount,
    Balance: item.Balance,
    Time: item.TimeStamp,
    Memo: item.Memo,
    Owner: item.Owner,
  }

  logger.L(ctx).Debugf("r %v", st)

  return &st, nil
}
