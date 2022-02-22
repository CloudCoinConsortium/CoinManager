package transactions

import (
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
)


type Transaction struct {
  Amount int `json:"amount"`
  Message string `json:"message"`
  Type string `json:"type"`
  Datetime time.Time `json:"datetime"`
  ReceiptID string `json:"receiptid"`
  Details []TransactionDetail `json:"details"`
  RunningBalance int `json:"running_balance"`
}

type TransactionDetail struct {
  Sn uint32 `json:"sn"`
  Denomination int `json:"denomination"`
  PownString string `json:"pownstring"`
  Result string `json:"result"`
}

func New(amount int, message string, ttype string, receiptID string) *Transaction {
  ts := time.Now()

  t := &Transaction{
    Amount: amount,
    Type: ttype,
    Message: message,
    ReceiptID: receiptID,
    Datetime: ts,
  }

  t.Details = make([]TransactionDetail, 0)

  return t
}

func (t *Transaction) AddDetail(sn uint32, pownstring string, result string) {
  td := TransactionDetail{
    Sn: sn,
    PownString: pownstring,
    Result: result,
  }

  td.Denomination = cloudcoin.GetDenomination(sn)

  t.Details = append(t.Details, td)
}
