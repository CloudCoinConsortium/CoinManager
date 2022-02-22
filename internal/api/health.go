package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/health"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	//"github.com/go-ozzo/ozzo-validation/v4/is"
)

type HealthRequest struct {
  Name string `json:"name"`
}

func (s *HealthRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.By(ValidateWallet)),
  )

  return err
}

type HealthCall struct {
  Parent
  Args *HealthRequest
  Wallet *wallets.Wallet
  Task *tasks.Task
}

func NewHealth(progressChannel chan interface{}, args *HealthRequest) *HealthCall {
  return &HealthCall{
    *NewParent(progressChannel),
    args,
    nil,
    nil,
  }
}

func (v *HealthCall) BatchFunction(ctx context.Context, coins []cloudcoin.CloudCoin) error {
  logger.L(ctx).Debugf("Health. Calling batch function for %d notes", len(coins))

  err := storage.GetDriver().UpdateStatus(ctx, v.Wallet, coins)
  if err != nil {
    logger.L(ctx).Debugf("Failed to set status for coins len=%d: %s", len(coins), err.Error())
    return err
  }

  return nil
}

func (v *HealthCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("Health Cheking Coins")

  /* Lite validation */
  wallet, err := storage.GetDriver().GetWalletWithContents(ctx, v.Args.Name)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.Wallet = wallet

  healther, _ := health.New(v.ProgressChannel, v.Task)
  response, err := healther.HealthCheck(ctx, wallet)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(response)
}


func HealthReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var healthRequest HealthRequest

  logger.L(ctx).Debugf("Health Request")
  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&healthRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = healthRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewHealth(task.ProgressChannel, &healthRequest)
  instance.Task = task
  

  // 10 is for Unpack and File Operations
  task.Run(ctx, instance)

  SuccessResponse(ctx, w, task)
}

