package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/skyhealth"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/skywallet"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	//"github.com/go-ozzo/ozzo-validation/v4/is"
)

type SkyHealthRequest struct {
  Name string `json:"name"`
}

func (s *SkyHealthRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.Required, validation.Length(config.SKYWALLET_NAME_MIN_LENGTH, config.SKYWALLET_NAME_MAX_LENGTH), is.DNSName),
  )

  return err
}

type SkyHealthCall struct {
  Parent
  Args *SkyHealthRequest
  Task *tasks.Task
}

func NewSkyHealth(progressChannel chan interface{}, args *SkyHealthRequest) *SkyHealthCall {
  return &SkyHealthCall{
    *NewParent(progressChannel),
    args,
    nil,
  }
}

func (v *SkyHealthCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("SkyHealth RUN")

  sw := skywallet.New(v.ProgressChannel)
  skyWallet, err := sw.GetWithContents(ctx, v.Args.Name)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  logger.L(ctx).Debugf("SkyHealth skywallet %s to %s", skyWallet.IDCoin.GetSkyName())

  skyhealther := skyhealth.New(v.ProgressChannel, v.Task)
  wr, err := skyhealther.SkyHealth(ctx, skyWallet)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(wr)
}


func SkyHealthReq(w http.ResponseWriter, r *http.Request) {  
  ctx := r.Context()
	var skyHealthRequest SkyHealthRequest

  logger.L(ctx).Debugf("SkyHealth Request")
  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&skyHealthRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = skyHealthRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewSkyHealth(task.ProgressChannel, &skyHealthRequest)
  instance.Task = task
  

  //task.SetTotalIterations(iterations)
  //25 for ShowRegister, 25 for Balance, and the others and 25*3 for 'getskywallet = (get balance, register and get statements). 
  task.SetTotalIterations(config.TOTAL_RAIDA_NUMBER * 5)

  task.Run(ctx, instance)

  SuccessResponse(ctx, w, task)
}

