package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/converter"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ConvertRequest struct {
  Name string `json:"name"`
  Coins []CoinRequest `json:"coins"`
}

func (s *ConvertRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.By(ValidateWallet)),
    validation.Field(&s.Coins, validation.Required, validation.Each(validation.By(ValidateCoin))),
  )

  return err
}


type ConvertCall struct {
  Parent
  Args *ConvertRequest
}

func NewConvert(progressChannel chan interface{}, args *ConvertRequest) *ConvertCall {
  return &ConvertCall{
    *NewParent(progressChannel),
    args,
  }
}


func (v *ConvertCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("Converting Coins")

  /* Lite validation */
  wallet, err := storage.GetDriver().GetWalletWithContents(ctx, v.Args.Name)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  ccs := make([]cloudcoin.CloudCoin, len(v.Args.Coins))
  for i := 0; i < len(ccs); i++ {
    ccs[i] = *cloudcoin.NewFromData(uint32(v.Args.Coins[i].Sn))
    ccs[i].Ans = v.Args.Coins[i].Ans
  }

  logger.L(ctx).Debugf("ccs %v", ccs)

  cter := converter.New(v.ProgressChannel, nil)
  response, err := cter.Convert(ctx, ccs, wallet)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(response)
}


func ConvertReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var convertRequest ConvertRequest

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&convertRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = convertRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewConvert(task.ProgressChannel, &convertRequest)

//  iterations := instance.CalcTotalIterations(len(instance.Args.Coins))
  //task.SetTotalIterations(iterations)
  task.Run(ctx, instance)
  
  SuccessResponse(ctx, w, task)
  
}
