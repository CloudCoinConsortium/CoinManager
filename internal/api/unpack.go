package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/unpacker"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	//"github.com/go-ozzo/ozzo-validation/v4/is"
)

type UnpackRequest struct {
	Data string `json:"data"`
}

func (s *UnpackRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Data, validation.Required, is.Base64),
  )

  return err
}

func UnpackReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var unpackRequest UnpackRequest

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&unpackRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = unpackRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }


  decoded, err := base64.StdEncoding.DecodeString(unpackRequest.Data)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_BASE64, "Failed to decode base64: " + err.Error()))
    return
  }

  unpacker := unpacker.New()
  coins, err := unpacker.Unpack(ctx, decoded)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_UNPACK, "Failed to unpack: " + err.Error()))
    return
  }

  SuccessResponse(ctx, w, coins)
}

