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
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type SyncAllRequest struct {
  Name string `json:"name"`
}

type SyncRequest struct {
  Name string `json:"name"`
  SyncItems map[int][]int `json:"sync_items"`
}

func (s *SyncRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.SyncItems, validation.Required, validation.By(ValidateSyncItems)),
    validation.Field(&s.Name, validation.Required, validation.Length(config.SKYWALLET_NAME_MIN_LENGTH, config.SKYWALLET_NAME_MAX_LENGTH), is.DNSName),
  )

  return err
}

func (s *SyncAllRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.Required, validation.Length(config.WALLET_NAME_MIN_LENGTH, config.WALLET_NAME_MAX_LENGTH)),
  )

  return err
}

func ValidateSyncItems(v interface{}) error {
  m := v.(map[int][]int)

  for sn, val := range(m) {
    err := validation.Validate(sn, validation.Required, validation.Min(1), validation.Max(config.TOTAL_COINS))
    if err != nil {
      return err
    }

    err = validation.Validate(val, validation.Length(config.TOTAL_RAIDA_NUMBER, config.TOTAL_RAIDA_NUMBER), validation.Each(validation.In(config.HEALTH_CHECK_STATUS_COUNTERFEIT, config.HEALTH_CHECK_STATUS_NETWORK, config.HEALTH_CHECK_STATUS_NOT_PRESENT, config.HEALTH_CHECK_STATUS_PRESENT, config.HEALTH_CHECK_STATUS_UNKNOWN)))
    if err != nil {
      return err
    }
  }

  return nil
}

type SyncCall struct {
  Parent
  Args *SyncRequest
  Task *tasks.Task
}

func NewSync(progressChannel chan interface{}, args *SyncRequest) *SyncCall {
  return &SyncCall{
    *NewParent(progressChannel),
    args,
    nil,
  }
}

type SyncAllCall struct {
  Parent
  Args *SyncAllRequest
}

func NewSyncAll(progressChannel chan interface{}, args *SyncAllRequest) *SyncAllCall {
  return &SyncAllCall{
    *NewParent(progressChannel),
    args,
  }
}


func (v *SyncCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("Syncing SkyWallet %s", v.Args.Name)

  sw := skywallet.New(v.ProgressChannel)
  skyWallet, err := sw.GetWithContentsOnly(ctx, v.Args.Name)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  logger.L(ctx).Debugf("sw %s", skyWallet.Name)

  healther := skyhealth.New(v.ProgressChannel, v.Task)

  if v.Task != nil {
    iterations := utils.GetTotalIterations(len(v.Args.SyncItems), healther.GetStridesSize())
    v.Task.SetTotalIterations(config.TOTAL_RAIDA_NUMBER * 2 + iterations)

  }

  hcr, err := healther.Sync(ctx, skyWallet, v.Args.SyncItems)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(hcr)

}

func SyncReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var syncRequest SyncRequest

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&syncRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = syncRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewSync(task.ProgressChannel, &syncRequest)
  instance.Task = task

  task.Run(ctx, instance)


  
  SuccessResponse(ctx, w, task)
  
}

func SyncAllReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var syncAllRequest SyncAllRequest

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&syncAllRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = syncAllRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  SuccessResponse(ctx, w, task)
  
}
