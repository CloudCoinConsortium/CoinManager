package api

import (
	"encoding/json"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gorilla/mux"
	//"github.com/go-ozzo/ozzo-validation/v4/is"
)

//func (v *WalletsCall) Run() {
//    s := raida.NewEcho(v.ProgressChannel)
//    s.Echo()
//}

type CreateWalletRequest struct {
	Name string `json:"name"`
  Email string `json:"email"`
  Password string `json:"password"`
}

type DeleteWalletRequest struct {
	Name string `json:"name"`
}

type GetWalletRequest struct {
	Name string 
}

type UpdateWalletRequest struct {
	Name string 
  NewName string
}

type UpdateWalletRequestBody struct {
  NewName string `json:"new_name"`
}

func (s *CreateWalletRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.By(ValidateWallet)),
    validation.Field(&s.Password, validation.Length(config.PASSWORD_MIN_LENGTH, config.PASSWORD_MAX_LENGTH)),
    validation.Field(&s.Email, is.Email),
  )

  return err
}

func (s *GetWalletRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.By(ValidateWallet)),
  )

  return err
}

func (s *UpdateWalletRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.By(ValidateWallet)),
    validation.Field(&s.NewName, validation.By(ValidateWallet)),
  )

  return err
}

func (s *DeleteWalletRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.By(ValidateWallet)),
  )

  return err
}

func WalletsReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  wallets, err := storage.GetDriver().GetWallets(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  for idx, _ := range(wallets) {
    err := storage.GetDriver().UpdateWalletBalance(ctx, &wallets[idx])
    if err != nil {
      logger.L(ctx).Errorf("Failed to update wallet %s balance: %s", wallets[idx].Name, err.Error())
    }
  }

  SuccessResponse(ctx, w, wallets)
}

func WalletReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	vars := mux.Vars(r)

  name, ok := vars["name"]
  if !ok {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_PARAMS_WALLET_NAME_MISSING, "Wallet Name Missing"))
    return
  }

  getWalletRequest := GetWalletRequest{
    Name: name,
  }
  err := getWalletRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  var wallet *wallets.Wallet

  query := r.URL.Query()
  v, ok := query["contents"]
  if ok && v[0] == "true" {
    logger.L(ctx).Debugf("Get Wallet with Contents")
    wallet, err = storage.GetDriver().GetWalletWithContents(ctx, name)
  } else {
    wallet, err = storage.GetDriver().GetWalletWithBalance(ctx, name)
  }

  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  SuccessResponse(ctx, w, wallet)
}


func CreateWalletReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var createWalletRequest CreateWalletRequest

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&createWalletRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = createWalletRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  err = storage.GetDriver().CreateWallet(ctx, createWalletRequest.Name, createWalletRequest.Email, createWalletRequest.Password)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  SuccessResponse(ctx, w, nil)
}


func DeleteWalletReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	vars := mux.Vars(r)

  name, ok := vars["name"]
  if !ok {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_PARAMS_WALLET_NAME_MISSING, "Wallet Name Missing"))
    return
  }

  getWalletRequest := DeleteWalletRequest{
    Name: name,
  }
  err := getWalletRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  err = storage.GetDriver().DeleteWallet(ctx, name)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  SuccessResponse(ctx, w, nil)
}

func UpdateWalletReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var updateWalletRequestBody UpdateWalletRequestBody
	vars := mux.Vars(r)

  name, ok := vars["name"]
  if !ok {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_PARAMS_WALLET_NAME_MISSING, "Wallet Name Missing"))
    return
  }

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&updateWalletRequestBody)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  newName := updateWalletRequestBody.NewName
  updateWalletRequest := UpdateWalletRequest{
    Name: name,
    NewName: newName,
  }

  err = updateWalletRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  logger.L(ctx).Debugf("Renaming wallet %s to %s", name, newName)
  wallet, err := storage.GetDriver().GetWallet(ctx, name)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  err = storage.GetDriver().RenameWallet(ctx, wallet, newName)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  SuccessResponse(ctx, w, nil)
}
