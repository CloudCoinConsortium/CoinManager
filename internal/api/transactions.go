package api

import (
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gorilla/mux"
	//"github.com/go-ozzo/ozzo-validation/v4/is"
)

//func (v *WalletsCall) Run() {
//    s := raida.NewEcho(v.ProgressChannel)
//    s.Echo()
//}

type DeleteTransactionRequest struct {
	Name string `json:"name"`
}

type GetTransactionRequest struct {
  Name string `json:"name"`
  Guid string `json:"guid"`
}

func (s *GetTransactionRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.By(ValidateWallet)),
    validation.Field(&s.Guid, validation.By(ValidateGUID)),
  )

  return err
}

func (s *DeleteTransactionRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.By(ValidateWallet)),
  )

  return err
}

func GetTransactionsReq(w http.ResponseWriter, r *http.Request) {
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

  getTransactionRequest := GetTransactionRequest{
    Name: name,
    Guid: guid,
  }
  err := getTransactionRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  wallet, err := storage.GetDriver().GetWallet(ctx, name)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  logger.L(ctx).Debugf("Getting transaction %s for wallet %s", guid, wallet.Name)


  trs, err := storage.GetDriver().GetReceipt(ctx, wallet, guid)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  SuccessResponse(ctx, w, trs)
}



func DeleteTransactionsReq(w http.ResponseWriter, r *http.Request) {
  ctx  := r.Context()
	vars := mux.Vars(r)

  name, ok := vars["name"]
  if !ok {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_PARAMS_WALLET_NAME_MISSING, "Wallet Name Missing"))
    return
  }

  deleteTransactionRequest := DeleteWalletRequest{
    Name: name,
  }
  err := deleteTransactionRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  wallet, err := storage.GetDriver().GetWallet(ctx, name)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  err = storage.GetDriver().DeleteTransactionsAndReceipts(ctx, wallet)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  SuccessResponse(ctx, w, nil)
}
