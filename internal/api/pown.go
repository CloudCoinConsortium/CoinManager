package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type PownRequest struct {
  Coins []CoinRequest
}

func (s *PownRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Coins, validation.Required, validation.Each(validation.By(ValidateCoinWithPans))),
  )

  return err
}


type PownCall struct {
  Parent
  Args *PownRequest
}

func NewPown(progressChannel chan interface{}, args *PownRequest) *PownCall {
  return &PownCall{
    *NewParent(progressChannel),
    args,
  }
}


func (v *PownCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("Powning Coins")

  ccs := make([]cloudcoin.CloudCoin, len(v.Args.Coins))
  for i := 0; i < len(ccs); i++ {
    ccs[i] = *cloudcoin.NewFromData(uint32(v.Args.Coins[i].Sn))
    ccs[i].Ans = v.Args.Coins[i].Ans
    ccs[i].Pans = v.Args.Coins[i].Pans
  }

  s := raida.NewPown(v.ProgressChannel)
  response, err := s.Pown(ctx, ccs)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(response)
}


func PownReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var pownRequest PownRequest

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&pownRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = pownRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewPown(task.ProgressChannel, &pownRequest)

  task.Run(ctx, instance)
  
  SuccessResponse(ctx, w, task)
  
}
