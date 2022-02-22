package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/transfer"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	//"github.com/go-ozzo/ozzo-validation/v4/is"
)

type TransferRequest struct {
  SrcName string `json:"srcname"`
  DstName string `json:"dstname"`
  Amount int `json:"amount"`
  Tag string `json:"tag"`
}

func (s *TransferRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.SrcName, validation.By(ValidateWallet)),
    validation.Field(&s.DstName, validation.By(ValidateWallet)),
    validation.Field(&s.Amount, validation.Required, validation.Min(1)),
    validation.Field(&s.Tag, validation.By(ValidateTag)),
  )

  return err
}

type TransferCall struct {
  Parent
  Args *TransferRequest
  SrcWallet *wallets.Wallet
}

func NewTransfer(progressChannel chan interface{}, args *TransferRequest) *TransferCall {
  return &TransferCall{
    *NewParent(progressChannel),
    args,
    nil,
  }
}

func (v *TransferCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("Transferring Coins")

  srcWallet, err := storage.GetDriver().GetWalletWithContents(ctx, v.Args.SrcName)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  dstWallet, err := storage.GetDriver().GetWalletWithContents(ctx, v.Args.DstName)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  logger.L(ctx).Debugf("From: %s To %s %d CC", srcWallet.Name, dstWallet.Name, v.Args.Amount)

  transfer := transfer.New(v.ProgressChannel)
  tr, err := transfer.Transfer(ctx, srcWallet, dstWallet, v.Args.Amount, v.Args.Tag)
  if err != nil {
    v.SendError(ctx, err)
    return
  }


  v.SendResult(tr)
}


func TransferReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var transferRequest TransferRequest

  logger.L(ctx).Debugf("Transfer Request")
  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&transferRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = transferRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewTransfer(task.ProgressChannel, &transferRequest)
  

  task.Run(ctx, instance)

  SuccessResponse(ctx, w, task)
}

