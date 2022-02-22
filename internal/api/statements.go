package api

import (
	"context"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/skywallet"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gorilla/mux"
	//"github.com/go-ozzo/ozzo-validation/v4/is"
)

type StatementsRequest struct {
  Name string `json:"name"`
}

func (s *StatementsRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.Required, validation.Length(config.WALLET_NAME_MIN_LENGTH, config.WALLET_NAME_MAX_LENGTH)),
  )

  return err
}

type StatementsCall struct {
  Parent
  Args *StatementsRequest
  Wallet *wallets.Wallet
  Task *tasks.Task
}

func NewStatements(progressChannel chan interface{}, args *StatementsRequest) *StatementsCall {
  return &StatementsCall{
    *NewParent(progressChannel),
    args,
    nil,
    nil,
  }
}

func (v *StatementsCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("Statements for %s", v.Args.Name)

  /* Lite validation */
  /*
  skywallet, err := storage.GetDriver().GetSkyWallet(v.Args.Name)
  if err != nil {
    v.SendError(ctx, err)
    return
  }
*/
 // v.Wallet = wallet

  sw := skywallet.New(v.ProgressChannel)
  err := sw.DeleteStatements(ctx, v.Args.Name)
  if err != nil {
    v.SendError(ctx, err)
    return
  }
  



  /*
  statementser, _ := statements.New(v.ProgressChannel, v.Task)
  response, err := statementser.StatementsCheck(wallet)
  if err != nil {
    v.SendError(ctx, err)
    return
  }
*/
  v.SendResult(nil)
}


func DeleteStatementsReq(w http.ResponseWriter, r *http.Request) { 
  ctx := r.Context()
  logger.L(ctx).Debugf("Delete Statements Request")
	vars := mux.Vars(r)

  name, ok := vars["name"]
  if !ok {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_PARAMS_WALLET_NAME_MISSING, "Wallet Name Missing"))
    return
  }

  statementsRequest := &StatementsRequest{
    Name: name,
  }

  err := statementsRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewStatements(task.ProgressChannel, statementsRequest)
  instance.Task = task
  

  // 10 is for Unpack and File Operations
  task.SetTotalIterations(config.TOTAL_RAIDA_NUMBER * 3)

  task.Run(ctx, instance)

  SuccessResponse(ctx, w, task)
}

