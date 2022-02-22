package api

import (
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/skywallet"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gorilla/mux"
	//"github.com/go-ozzo/ozzo-validation/v4/is"
)

//func (v *WalletsCall) Run() {
//    s := raida.NewEcho(v.ProgressChannel)
//    s.Echo()
//}

type GetSkyTransactionRequest struct {
  Name string `json:"name"`
  Guid string `json:"guid"`
}

func (s *GetSkyTransactionRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.By(ValidateSkyWallet)),
    validation.Field(&s.Guid, validation.By(ValidateGUID)),
  )

  return err
}

func GetSkyTransactionsReq(w http.ResponseWriter, r *http.Request) {
  ctx  := r.Context()
	vars := mux.Vars(r)

  name, ok := vars["name"]
  if !ok {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_PARAMS_WALLET_NAME_MISSING, "Wallet Name Missing"))
    return
  }

  guid, ok := vars["guid"]
  if !ok {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_PARAMS_WALLET_NAME_MISSING, "GUID Missing"))
    return
  }

  getSkyTransactionRequest := GetSkyTransactionRequest{
    Name: name,
    Guid: guid,
  }
  err := getSkyTransactionRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  sw := skywallet.New(nil)
  wallet, err := sw.GetIDOnly(ctx, name)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  st, err := sw.GetReceipt(ctx, wallet, guid)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  SuccessResponse(ctx, w, st)
}



