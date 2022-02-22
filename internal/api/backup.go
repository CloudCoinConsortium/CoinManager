package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/backup"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	//"github.com/go-ozzo/ozzo-validation/v4/is"
)

//func (v *WalletsCall) Run() {
//    s := raida.NewEcho(v.ProgressChannel)
//    s.Echo()
//}

type BackupRequest struct {
	Name string `json:"name"`
  Folder string `json:"folder"`
  Tag string `json:"tag"`
}

func (s *BackupRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.By(ValidateWallet)),
    validation.Field(&s.Tag, validation.By(ValidateTag)),
    validation.Field(&s.Folder, validation.Required),
  )

  return err
}

type BackupCall struct {
  Parent
  Args *BackupRequest
  Task *tasks.Task
}

func NewBackup(progressChannel chan interface{}, args *BackupRequest) *BackupCall {
  return &BackupCall{
    *NewParent(progressChannel),
    args,
    nil,
  }
}

func (v *BackupCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("Backuping Coins for %s", v.Args.Name)

  wallet, err := storage.GetDriver().GetWalletWithContents(ctx, v.Args.Name)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  logger.L(ctx).Debugf("c %v", wallet.Contents)
  v.Task.SetTotalIterations(len(wallet.Contents))

  backup := backup.New(v.ProgressChannel, v.Task)
  br, err := backup.Backup(ctx, wallet, v.Args.Folder, v.Args.Tag)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(br)
}

func BackupReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var backupRequest BackupRequest

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&backupRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = backupRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }


  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewBackup(task.ProgressChannel, &backupRequest)
  instance.Task = task
  

  //task.SetTotalIterations(iterations)
  //25 for ShowRegister and the others and 25*2 for 'getskywallet = (get balance and get statements). 
  task.SetTotalIterations(config.TOTAL_RAIDA_NUMBER * 3)

  task.Run(ctx, instance)

  SuccessResponse(ctx, w, task)
}


