package api

import (
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
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

type LeftOverRequest struct {
	Name string `json:"name"`
}

func (s *LeftOverRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.By(ValidateWallet)),
  )

  return err
}


func LeftOverReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	vars := mux.Vars(r)

  name, ok := vars["name"]
  if !ok {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_PARAMS_WALLET_NAME_MISSING, "Wallet Name Missing"))
    return
  }

  leftOverRequest := LeftOverRequest{
    Name: name,
  }
  err := leftOverRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  wallet, err := storage.GetDriver().GetWallet(ctx, name)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  err = storage.GetDriver().UpdateWalletContentsForLocation(ctx, wallet, config.COIN_LOCATION_STATUS_SUSPECT)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  SuccessResponse(ctx, w, wallet)
}


