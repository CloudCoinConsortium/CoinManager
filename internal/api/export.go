package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/exporter"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ExportRequest struct {
  Name string `json:"name"`
  Amount int `json:"amount"`
  Tag string `json:"tag"`
  Type string `json:"type"`
  Folder string `json:"folder"`
}


type ExportResponse struct {
  Coins string `json:"coins"`
}

func (s *ExportRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Amount, validation.Required, validation.Min(1)),
    validation.Field(&s.Name, validation.By(ValidateWallet)),
    validation.Field(&s.Tag, validation.By(ValidateTag)),
    validation.Field(&s.Type, validation.In(config.EXPORT_TYPE_PNG, config.EXPORT_TYPE_ZIP, config.EXPORT_TYPE_BIN)),
    validation.Field(&s.Folder, validation.Required),
  )

  return err
}

type ExportCall struct {
  Parent
  Args *ExportRequest
  Wallet *wallets.Wallet
}

func NewExport(progressChannel chan interface{}, args *ExportRequest) *ExportCall {
  return &ExportCall{
    *NewParent(progressChannel),
    args,
    nil,
  }
}

func (v *ExportCall) Run(ctx context.Context) {

  logger.L(ctx).Debugf("Exporting Coins")

  /* Lite validation */
  wallet, err := storage.GetDriver().GetWalletWithContents(ctx, v.Args.Name)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.Wallet = wallet
  v.SendProgress("Exporting Coins", 5)

  logger.L(ctx).Debugf("We have read wallet contents, balance %d", wallet.Balance)

  exporter, err := exporter.New(v.ProgressChannel)
  if err != nil {
    logger.L(ctx).Errorf("Failed to init exporter: %s", err.Error())
    v.SendError(ctx, err)
    return
  }

  etype := v.Args.Type
  if etype == "" {
    logger.L(ctx).Debugf("Defaulting export type to PNG")
    etype = config.EXPORT_TYPE_PNG
  }

  // Ignore data
  _, err = exporter.Export(ctx, wallet, v.Args.Amount, v.Args.Tag, v.Args.Folder, etype)
  if err != nil {
    v.SendError(ctx, err)
    return
  }


  //logger.L(ctx).Debugf("v=%v", data)


//  exportResponse := &ExportResponse{}
//  exportResponse.Coins = base64.StdEncoding.EncodeToString(data)

//  v.SendResult(exportResponse)
  v.SendResult(nil)
}


func ExportReq(w http.ResponseWriter, r *http.Request) {

  ctx := r.Context()
	var exportRequest ExportRequest

  logger.L(ctx).Debugf("Export Request")
  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&exportRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = exportRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewExport(task.ProgressChannel, &exportRequest)
  
  // 10 is for filesystem operations
  task.SetTotalIterations(config.TOTAL_RAIDA_NUMBER + 10)
  task.Run(ctx, instance)

  SuccessResponse(ctx, w, task)
}

