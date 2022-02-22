package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/deposit"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/dnsservice"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type DepositRequest struct {
  Name string `json:"name"`
  To int `json:"to"`
  ToName string `json:"toname"`
  Amount int `json:"amount"`
  Tag string `json:"tag"`
}


type DepositResponse struct {
  Amount int `json:"amount"`
}

func (s *DepositRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Amount, validation.Required, validation.Min(1)),
    validation.Field(&s.Name, validation.Required, validation.By(ValidateWallet)),
    validation.Field(&s.To, validation.Min(1), validation.Max(config.TOTAL_COINS), validation.When(s.ToName == "", validation.Required)),
    validation.Field(&s.ToName, validation.Length(config.SKYWALLET_NAME_MIN_LENGTH, config.SKYWALLET_NAME_MAX_LENGTH), is.DNSName, validation.When(s.To == 0, validation.Required)),
    validation.Field(&s.Tag, validation.By(ValidateTag)),
  )

  return err
}

type DepositCall struct {
  Parent
  Args *DepositRequest
  Wallet *wallets.Wallet
}

func NewDeposit(progressChannel chan interface{}, args *DepositRequest) *DepositCall {
  return &DepositCall{
    *NewParent(progressChannel),
    args,
    nil,
  }
}

func (v *DepositCall) Run(ctx context.Context) {

  logger.L(ctx).Debugf("Depositing Coins")

  /* Lite validation */
  wallet, err := storage.GetDriver().GetWalletWithContents(ctx, v.Args.Name)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.Wallet = wallet
  v.SendProgress("Depositing Coins", 5)

  logger.L(ctx).Debugf("We have read wallet contents, balance %d", wallet.Balance)

  depositer, err := deposit.New(v.ProgressChannel)
  if err != nil {
    logger.L(ctx).Errorf("Failed to init depositer: %s", err.Error())
    v.SendError(ctx, err)
    return
  }

  to := uint32(v.Args.To)
  if v.Args.To == 0 {
    logger.L(ctx).Debugf("To is zero. Resolving ToName %s", v.Args.ToName)

    dns := dnsservice.New(nil)
    sn, err := dns.GetSN(ctx, v.Args.ToName)
    if err != nil || sn == 0 {
      v.SendError(ctx, perror.New(perror.ERROR_DNS_RESOLVE, "Failed to find SkyVault address"))
      return
    }

    to = sn
  }

  data, err := depositer.Deposit(ctx, wallet, to, v.Args.ToName, v.Args.Amount, v.Args.Tag)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(data)
}


func DepositReq(w http.ResponseWriter, r *http.Request) {
	var depositRequest DepositRequest

  ctx := r.Context()

  logger.L(ctx).Debugf("Deposit Request")
  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&depositRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = depositRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewDeposit(task.ProgressChannel, &depositRequest)
  
  task.SetTotalIterations(config.TOTAL_RAIDA_NUMBER)
  task.Run(ctx, instance)

  SuccessResponse(ctx, w, task)
}

