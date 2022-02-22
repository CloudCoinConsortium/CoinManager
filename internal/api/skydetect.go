package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/skywallet"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type SkyDetectRequest struct {
  Name string `json:"name"`
}

func (s *SkyDetectRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.Required, validation.Length(config.SKYWALLET_NAME_MIN_LENGTH, config.SKYWALLET_NAME_MAX_LENGTH), is.DNSName),
  )

  return err
}


type SkyDetectCall struct {
  Parent
  Args *SkyDetectRequest
}

func NewSkyDetect(progressChannel chan interface{}, args *SkyDetectRequest) *SkyDetectCall {
  return &SkyDetectCall{
    *NewParent(progressChannel),
    args,
  }
}


func (v *SkyDetectCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("Detecting ID coin %s", v.Args.Name)

  sw := skywallet.New(v.ProgressChannel)
  skyWallet, err := sw.GetIDOnly(ctx, v.Args.Name)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  logger.L(ctx).Debugf("Got skywallet %s is sky %v", skyWallet.Name, skyWallet.IDCoin.IsIDCoin())

  d := raida.NewDetect(v.ProgressChannel)
  response, err := d.Detect(ctx, []cloudcoin.CloudCoin{*skyWallet.IDCoin})
  if err != nil {
    v.SendError(ctx, err)
  }

  v.SendResult(response)
}


func SkyDetectReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var skyDetectRequest SkyDetectRequest

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&skyDetectRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = skyDetectRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewSkyDetect(task.ProgressChannel, &skyDetectRequest)
  task.Run(ctx, instance)
  
  SuccessResponse(ctx, w, task)
  
}
