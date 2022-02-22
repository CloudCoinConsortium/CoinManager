package api

import (
	"encoding/json"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/backup"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/skywallet"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	//"github.com/go-ozzo/ozzo-validation/v4/is"
)

//func (v *WalletsCall) Run() {
//    s := raida.NewEcho(v.ProgressChannel)
//    s.Echo()
//}

type SkyBackupRequest struct {
	Name string `json:"name"`
  Folder string `json:"folder"`
  Tag string `json:"tag"`
}

type SkyBackupResponse struct {
  FileName string
}

func (s *SkyBackupRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.By(ValidateSkyWallet)),
    validation.Field(&s.Tag, validation.By(ValidateTag)),
  )

  return err
}

func SkyBackupReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var backupRequest SkyBackupRequest

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

  sw := skywallet.New(nil)
  wallet, err := sw.GetIDOnly(ctx, backupRequest.Name)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  backup := backup.New(nil, nil)
  fpath, err := backup.SkyBackup(ctx, wallet, backupRequest.Folder, backupRequest.Tag)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }


  sbResponse := &SkyBackupResponse{}
  sbResponse.FileName = fpath

  SuccessResponse(ctx, w, sbResponse)
}


