package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/skytransfer"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/skywallet"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	//"github.com/go-ozzo/ozzo-validation/v4/is"
)

type SkyTransferRequest struct {
  SrcName string `json:"srcname"`
  DstName string `json:"dstname"`
  To int `json:"to"`
  Amount int `json:"amount"`
  Tag string `json:"tag"`
}

func (s *SkyTransferRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.SrcName, validation.Required, validation.Length(config.SKYWALLET_NAME_MIN_LENGTH, config.SKYWALLET_NAME_MAX_LENGTH), is.DNSName),
    validation.Field(&s.To, validation.Min(1), validation.Max(config.TOTAL_COINS), validation.When(s.DstName == "", validation.Required)),
    validation.Field(&s.DstName, validation.Length(config.SKYWALLET_NAME_MIN_LENGTH, config.SKYWALLET_NAME_MAX_LENGTH), is.DNSName, validation.When(s.To == 0, validation.Required)),
    validation.Field(&s.Amount, validation.Required, validation.Min(1)),
    validation.Field(&s.Tag, validation.Length(1, config.MAX_TAG_LENGTH), validation.By(ValidateTag)),
  )

  return err
}

type SkyTransferCall struct {
  Parent
  Args *SkyTransferRequest
  Task *tasks.Task
}

func NewSkyTransfer(progressChannel chan interface{}, args *SkyTransferRequest) *SkyTransferCall {
  return &SkyTransferCall{
    *NewParent(progressChannel),
    args,
    nil,
  }
}

func (v *SkyTransferCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("Sky Transferring Coins")

  sw := skywallet.New(v.ProgressChannel)
  srcSkyWallet, err := sw.GetWithContentsOnly(ctx, v.Args.SrcName)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  logger.L(ctx).Debugf("Skywallet %s %d", srcSkyWallet.IDCoin.GetSkyName(), srcSkyWallet.IDCoin.Sn)

  st := skytransfer.New(v.ProgressChannel, v.Task)
  result, err := st.SkyTransfer(ctx, srcSkyWallet, v.Args.DstName, uint32(v.Args.To), v.Args.Amount, v.Args.Tag)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(result)

}


func SkyTransferReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var skyTransferRequest SkyTransferRequest

  logger.L(ctx).Debugf("SkyTransfer Request")
  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&skyTransferRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = skyTransferRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewSkyTransfer(task.ProgressChannel, &skyTransferRequest)
  instance.Task = task
  

  //25 for ShowRegister and the others and 25*2 for 'getskywallet = (get balance and get statements). Later we will set another progress for transfer itself
  task.SetTotalIterations(config.TOTAL_RAIDA_NUMBER * 2)
  task.Run(ctx, instance)

  SuccessResponse(ctx, w, task)
}

