package api

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/showpayment"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	//"github.com/go-ozzo/ozzo-validation/v4/is"
)

type ShowPaymentRequest struct {
	Guid string `json:"guid"`
}

func (s *ShowPaymentRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Guid, validation.Required, validation.Match(regexp.MustCompile("^[a-fA-F0-9]{32}$"))),
  )

  return err
}

type ShowPaymentCall struct {
  Parent
  Args *ShowPaymentRequest
}

func NewShowPayment(progressChannel chan interface{}, args *ShowPaymentRequest) *ShowPaymentCall {
  return &ShowPaymentCall{
    *NewParent(progressChannel),
    args,
  }
}

func (v *ShowPaymentCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("Task. ShowPayment %s", v.Args.Guid)

  guid := v.Args.Guid

  sp := showpayment.New(v.ProgressChannel)
  tr, err := sp.Show(ctx, guid)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(tr)
}

func ShowPaymentReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  logger.L(ctx).Debugf("ShowPayment request")

	var showPaymentRequest ShowPaymentRequest

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&showPaymentRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = showPaymentRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewShowPayment(task.ProgressChannel, &showPaymentRequest)
  task.Run(ctx, instance)
  
  SuccessResponse(ctx, w, task)
}
