package api

import (
	"net/http"
	"runtime"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/core"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/dlgs"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)


type FilePickerRequest struct {
  Type string `json:"type"`
}

func (s *FilePickerRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Type, validation.In(config.FILEPICKER_TYPE_FILE, config.FILEPICKER_TYPE_FOLDER)),
  )

  return err
}

type FilePickerCall struct {
  Parent
}

type FilePickerResponse struct {
  ItemsPicked []string `json:"items_picked"`
}

func NewFilePicker(progressChannel chan interface{}) *FilePickerCall {
  return &FilePickerCall{
    *NewParent(progressChannel),
  }
}

func FilePickerReq(w http.ResponseWriter, r *http.Request) {

  ctx := r.Context()

  logger.L(ctx).Debugf("File picker")

  query := r.URL.Query()

  ftype := config.FILEPICKER_TYPE_FOLDER
  rftype, ok := query["type"]

  logger.L(ctx).Debugf("query rftype %s", rftype)
  if ok && len(rftype) > 0 {
    fpr := &FilePickerRequest{}
    fpr.Type = rftype[0]
    err := fpr.Validate()
    if err != nil {
      ErrorResponse(ctx, w, err)
      return
    }

    ftype = fpr.Type

  }

  logger.L(ctx).Debugf("Picking type %s", ftype)

  var files []string
  var err error

  pt := core.GetUIHandle()
  logger.L(ctx).Debugf("ui handle  %v",pt)
  if ftype == config.FILEPICKER_TYPE_FOLDER {
    var folder string

    folder, ok, err = dlgs.File("Select Folder", "", true,  pt)

    files = make([]string, 1)
    files[0] = folder
  } else if ftype == config.FILEPICKER_TYPE_FILE {
    //files, ok, err = dlgs.FileMulti("Select CloudCoins", "*.bin *.png")
    var filter string
    if runtime.GOOS == "windows" {
      filter = "CloudCoins" + string(0x0) +"*.png;*.bin;*.zip" 
    } else {
      filter = "*.bin *.png *.zip"
    }

    files, ok, err = dlgs.FileMulti("Select CloudCoins", filter, pt)
  } else {
      ErrorResponse(ctx, w, perror.New(perror.ERROR_INTERNAL, "Invalid ftype"))
      return
  }

  if err != nil {
    logger.L(ctx).Errorf("Failed to pick: %s", err.Error())
    ErrorResponse(ctx, w, perror.New(perror.ERROR_FILE_PICK, "Failed to pick: " + err.Error()))
    return
  }

  if !ok {
    logger.L(ctx).Warnf("User cancelled dialog")
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DIALOG_CANCELLED, "Dialog cancelled"))
    return
  }
  logger.L(ctx).Debugf("Chosen items %v", files)

  rs := &FilePickerResponse{
    ItemsPicked: files,
  }

  SuccessResponse(ctx, w, rs)
}
