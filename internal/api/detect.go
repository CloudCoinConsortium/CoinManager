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
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type DetectRequest struct {
  Coins []CoinRequest
}

func (s *DetectRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Coins, validation.Required, validation.Each(validation.By(ValidateCoin))),
  )

  return err
}


type DetectCall struct {
  Parent
  Task *tasks.Task
  Args *DetectRequest

}

func NewDetect(progressChannel chan interface{}, task *tasks.Task, args *DetectRequest) *DetectCall {
  return &DetectCall{
    *NewParent(progressChannel),
    task,
    args,
  }
}


func (v *DetectCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("Detecting Coins")

  ccs := make([]cloudcoin.CloudCoin, len(v.Args.Coins))
  for i := 0; i < len(ccs); i++ {
    ccs[i] = *cloudcoin.NewFromData(uint32(v.Args.Coins[i].Sn))
    ccs[i].Ans = v.Args.Coins[i].Ans

    ccs[i].SetCoinID(uint16(v.Args.Coins[i].CoinType))
  }

  s := raida.NewDetect(v.ProgressChannel)
  
  iterations := utils.GetTotalIterations(len(v.Args.Coins), s.GetStrideSize())
  v.Task.SetTotalIterations(iterations)

  response, err := s.Detect(ctx, ccs)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(response)
}


func DetectReq(w http.ResponseWriter, r *http.Request) {

  ctx := r.Context()
	var detectRequest DetectRequest

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&detectRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = detectRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewDetect(task.ProgressChannel, task, &detectRequest)

  task.Run(ctx, instance)
  
  SuccessResponse(ctx, w, task)
  
}
