package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/skywallet"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/withdraw"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	//"github.com/go-ozzo/ozzo-validation/v4/is"
)

type WithdrawRequest struct {
  SrcName string `json:"srcname"`
  DstName string `json:"dstname"`
  Amount int `json:"amount"`
  Tag string `json:"tag"`
}

func (s *WithdrawRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.SrcName, validation.By(ValidateSkyWallet)),
    validation.Field(&s.DstName, validation.By(ValidateWallet)),
    validation.Field(&s.Amount, validation.Required, validation.Min(1)),
    validation.Field(&s.Tag, validation.By(ValidateTag)),
  )

  return err
}

type WithdrawCall struct {
  Parent
  Args *WithdrawRequest
  Task *tasks.Task
}

func NewWithdraw(progressChannel chan interface{}, args *WithdrawRequest) *WithdrawCall {
  return &WithdrawCall{
    *NewParent(progressChannel),
    args,
    nil,
  }
}

func (v *WithdrawCall) Run(ctx context.Context) {
  sw := skywallet.New(v.ProgressChannel)
  srcSkyWallet, err := sw.GetWithContentsOnly(ctx, v.Args.SrcName)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  dstWallet, err := storage.GetDriver().GetWalletWithContents(ctx, v.Args.DstName)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  logger.L(ctx).Debugf("Withdraw from skywallet %s to %s", srcSkyWallet.IDCoin.GetSkyName(), dstWallet.Name)

  withdraw := withdraw.New(v.ProgressChannel, v.Task)
  wr, err := withdraw.Withdraw(ctx, srcSkyWallet, dstWallet, v.Args.Amount, v.Args.Tag)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(wr)
}


func WithdrawReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var withdrawRequest WithdrawRequest

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&withdrawRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = withdrawRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewWithdraw(task.ProgressChannel, &withdrawRequest)
  instance.Task = task
  

  //task.SetTotalIterations(iterations)
  //25 for ShowRegister and the others and 25*2 for 'getskywallet = (get balance and get statements). 
  task.SetTotalIterations(config.TOTAL_RAIDA_NUMBER * 2)

  task.Run(ctx, instance)

  SuccessResponse(ctx, w, task)
}

