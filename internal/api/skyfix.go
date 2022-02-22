package api

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/fix"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/skywallet"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)


type SkyFixRequest struct {
  Name string `json:"name"`
  PownString string `json:"pownstring"`
}


func (s *SkyFixRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.PownString, validation.Required, validation.Match(regexp.MustCompile("^[upfne]{" + strconv.Itoa(config.TOTAL_RAIDA_NUMBER) + "}$"))),
    validation.Field(&s.Name, validation.By(ValidateSkyWallet)),
  )

  return err
}

type SkyFixCall struct {
  Parent
  Args *SkyFixRequest
}

func NewSkyFix(progressChannel chan interface{}, args *SkyFixRequest) *SkyFixCall {
  return &SkyFixCall{
    *NewParent(progressChannel),
    args,
  }
}

func (v *SkyFixCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("Fixing ID Coin %s: %s", v.Args.Name, v.Args.PownString)

  /* Lite validation */
  sw := skywallet.New(v.ProgressChannel)
  skyWallet, err := sw.GetIDOnly(ctx, v.Args.Name)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  cc := skyWallet.IDCoin
  cc.SetPownString(v.Args.PownString)

  raidas := []int{}
  logger.L(ctx).Debugf("Got coin %d (%s) isSky %v", cc.GetSkyName(), cc.Sn, cc.IsIDCoin())
  for ridx, status := range(cc.Statuses) {
    logger.L(ctx).Debugf("r%d st=%d", cc.Sn, ridx, status)
    if status == config.COIN_STATUS_COUNTERFEIT {
      logger.L(ctx).Debugf("failed on raida %d", cc.Sn, ridx)
      raidas = append(raidas, ridx)
    }
  }

  logger.L(ctx).Debugf("raidas to fix %v", raidas)

  fixer, _ := fix.New(v.ProgressChannel)
  fr, err := fixer.SkyFix(ctx, skyWallet, raidas)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(fr)

}

func SkyFixReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var fixRequest SkyFixRequest

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&fixRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = fixRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewSkyFix(task.ProgressChannel, &fixRequest)
  task.Run(ctx, instance)
  
  SuccessResponse(ctx, w, task)
  
}

