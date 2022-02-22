package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/dnsservice"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/skywallet"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gorilla/mux"
	//"github.com/go-ozzo/ozzo-validation/v4/is"
)

//func (v *WalletsCall) Run() {
//    s := raida.NewEcho(v.ProgressChannel)
//    s.Echo()
//}

type CreateSkyWalletRequest struct {
	Name string `json:"name"`
  Coin CoinRequest `json:"coin"`
  Type string `json:"type"`
}

type DeleteSkyWalletRequest struct {
	Name string `json:"name"`
  FileOnly bool `json:"file_only"`
}

type GetSkyWalletRequest struct {
  Name string `json:"name"`
}

type UpdateSkyWalletRequest struct {
  Name string `json:"name"`
  Private bool `json:"private"`
}

type UpdateSkyWalletRequestBody struct {
  Private bool `json:"private"`
}

type GetSkyWalletsRequest struct {
}

func (s *CreateSkyWalletRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.Required, validation.Length(config.SKYWALLET_NAME_MIN_LENGTH, config.SKYWALLET_NAME_MAX_LENGTH), is.DNSName),
    validation.Field(&s.Coin, validation.Required, validation.By(ValidateCoin)),
    validation.Field(&s.Coin, validation.Required, validation.By(ValidateCoin)),
    validation.Field(&s.Type, validation.In(config.SKYVAULT_TYPE_BIN, config.SKYVAULT_TYPE_CARD)),
  )

  return err
}

func (s *GetSkyWalletRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.Required, validation.Length(config.SKYWALLET_NAME_MIN_LENGTH, config.SKYWALLET_NAME_MAX_LENGTH), is.DNSName),
  )

  return err
}

func (s *UpdateSkyWalletRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.Required, validation.Length(config.SKYWALLET_NAME_MIN_LENGTH, config.SKYWALLET_NAME_MAX_LENGTH), is.DNSName),
    validation.Field(&s.Private),
  )

  return err
}


type CreateSkyWalletCall struct {
  Parent
  Args *CreateSkyWalletRequest
}

type GetSkyWalletCall struct {
  Parent
  Args *GetSkyWalletRequest
}

type GetSkyWalletsCall struct {
  Parent
  Args *GetSkyWalletsRequest
}

type DeleteSkyWalletCall struct {
  Parent
  Args *DeleteSkyWalletRequest
}



func NewCreateSkyWallet(progressChannel chan interface{}, args *CreateSkyWalletRequest) *CreateSkyWalletCall {
  return &CreateSkyWalletCall{
    *NewParent(progressChannel),
    args,
  }
}

func NewGetSkyWallet(progressChannel chan interface{}, args *GetSkyWalletRequest) *GetSkyWalletCall {
  return &GetSkyWalletCall{
    *NewParent(progressChannel),
    args,
  }
}

func NewGetSkyWallets(progressChannel chan interface{}, args *GetSkyWalletsRequest) *GetSkyWalletsCall {
  return &GetSkyWalletsCall{
    *NewParent(progressChannel),
    args,
  }
}

func NewDeleteSkyWallet(progressChannel chan interface{}, args *DeleteSkyWalletRequest) *DeleteSkyWalletCall {
  return &DeleteSkyWalletCall{
    *NewParent(progressChannel),
    args,
  }
}

func (s *DeleteSkyWalletRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.Required, validation.Length(config.SKYWALLET_NAME_MIN_LENGTH, config.SKYWALLET_NAME_MAX_LENGTH), is.DNSName),
  )

  return err
}

func (v *GetSkyWalletsCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("Task. Get Skywallets")

  wallets, err := storage.GetDriver().GetSkyWallets(ctx)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  sw := skywallet.New(nil)
  for idx, _ := range(wallets) {
    err := sw.UpdateBalance(ctx, &wallets[idx])
    if err != nil {
      logger.L(ctx).Errorf("Failed to update wallet %s balance: %s", wallets[idx].Name, err.Error())
    }
  }

  v.SendResult(wallets)
}

func (v *GetSkyWalletCall) Run(ctx context.Context) {
  name := v.Args.Name

  sw := skywallet.New(v.ProgressChannel)
  wallet, err := sw.Get(ctx, name)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(wallet)
}

func (v *CreateSkyWalletCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("Running task to register Skywallet %s", v.Args.Name)
  // Get IP
  sn := uint32(v.Args.Coin.Sn)

  cc := cloudcoin.NewFromData(sn)
  cc.SetAns(v.Args.Coin.Ans)
  cc.SetSkyName(v.Args.Name)


  var cr *skywallet.Card
  var err error
  sw := skywallet.New(v.ProgressChannel)
  if v.Args.Type == config.SKYVAULT_TYPE_CARD {
    logger.L(ctx).Debugf("Will register a card")
    cr, err = sw.RegisterNewCard(ctx, cc)
  } else {
    logger.L(ctx).Debugf("Will register a bin ID")
    err = sw.RegisterNew(ctx, cc)
  }

  if err != nil {
    logger.L(ctx).Errorf("Failed to register skywallet %s: %s", cc.GetSkyName(), err.Error())
    v.SendError(ctx, err)
    return 
  }

  v.SendResult(cr)
}


func (v *DeleteSkyWalletCall) Run(ctx context.Context) {
  name := v.Args.Name
  logger.L(ctx).Debugf("Running task to delete SkyWallet %s", name)

  if v.Args.FileOnly != true {
    logger.L(ctx).Debugf("Deleting filename only")
    sw := skywallet.New(nil)
    skywallet, err := sw.GetWithBalance(ctx, name)
    if err != nil {
      v.SendError(ctx, err)
      return 
    }

    if skywallet.Balance != 0 {
      logger.L(ctx).Errorf("SkyWallet %s is not empty. The balance is %d", skywallet.Name, skywallet.Balance)
      v.SendError(ctx, perror.New(perror.ERROR_WALLET_NOT_EMPTY, "Wallet is not empty"))
      return
    }

    logger.L(ctx).Debugf("Deleting DNS record")
    ds := dnsservice.New(v.ProgressChannel)
    err = ds.DeleteName(ctx, skywallet.IDCoin)
    if err != nil {
      logger.L(ctx).Errorf("Failed to register name %s: %s", name, err.Error())
      v.SendError(ctx, err)
      return 
    }
  }

  err := storage.GetDriver().DeleteSkyWallet(ctx, name)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(nil)
}


func SkyWalletsReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  logger.L(ctx).Debugf("Skywallets request")

  query := r.URL.Query()

  listOnlyS, ok := query["list_only"]
  if ok && len(listOnlyS) > 0 {
  listOnly, _ := strconv.ParseBool(listOnlyS[0])
  if listOnly {
      logger.L(ctx).Debugf("List only")
      wallets, err := storage.GetDriver().GetSkyWallets(ctx)
      if err != nil {
        ErrorResponse(ctx, w, err)
        return
      }

      SuccessResponse(ctx, w, wallets)
      return
    }
  }

  getSkyWalletsRequest := GetSkyWalletsRequest{}

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewGetSkyWallets(task.ProgressChannel, &getSkyWalletsRequest)

  // + 1 to query DNS Server
  task.SetTotalIterations(config.TOTAL_RAIDA_NUMBER + 1)
  task.Run(ctx, instance)
  
  SuccessResponse(ctx, w, task)
  /*
*/
//  SuccessResponse(ctx, w, wallets)
}

func SkyWalletReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  logger.L(ctx).Debugf("Skywallet request")
	vars := mux.Vars(r)

  name, ok := vars["name"]
  if !ok {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_PARAMS_WALLET_NAME_MISSING, "Wallet Name Missing"))
    return
  }

  getSkyWalletRequest := GetSkyWalletRequest{
    Name: name,
  }
  err := getSkyWalletRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewGetSkyWallet(task.ProgressChannel, &getSkyWalletRequest)

  // * 2 for get balance and get statments + 1 to query DNS Server
  task.SetTotalIterations(config.TOTAL_RAIDA_NUMBER * 2 + 1)
  task.Run(ctx, instance)
  
  SuccessResponse(ctx, w, task)
}


func CreateSkyWalletReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var createSkyWalletRequest CreateSkyWalletRequest

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&createSkyWalletRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = createSkyWalletRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  if createSkyWalletRequest.Type == "" {
    createSkyWalletRequest.Type = config.SKYVAULT_TYPE_BIN
  }


  logger.L(ctx).Debugf("skywallet type %s", createSkyWalletRequest.Type)

  instance := NewCreateSkyWallet(task.ProgressChannel, &createSkyWalletRequest)

  if createSkyWalletRequest.Type == config.SKYVAULT_TYPE_CARD {
    // If card we need to pown first
    task.SetTotalIterations(config.TOTAL_RAIDA_NUMBER * 2 + 1)
  } else {
    // + 1 to query DNS Server
    task.SetTotalIterations(config.TOTAL_RAIDA_NUMBER + 1)
  }
  task.Run(ctx, instance)
  
  SuccessResponse(ctx, w, task)


 // SuccessResponse(ctx, w, nil)
}


func DeleteSkyWalletReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	vars := mux.Vars(r)

  name, ok := vars["name"]
  if !ok {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_PARAMS_WALLET_NAME_MISSING, "Wallet Name Missing"))
    return
  }

  var fonly bool = false
  query := r.URL.Query()
  _, ok = query["file_only"]
  if ok {
    fonly = true
  }

  logger.L(ctx).Debugf("File only %v", fonly)

  deleteSkyWalletRequest := DeleteSkyWalletRequest{
    Name: name,
    FileOnly: fonly,
  }
  err := deleteSkyWalletRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewDeleteSkyWallet(task.ProgressChannel, &deleteSkyWalletRequest)

  // + 1 to query DNS Server
  task.SetTotalIterations(config.TOTAL_RAIDA_NUMBER + 1)
  task.Run(ctx, instance)
  
  SuccessResponse(ctx, w, task)

}

func UpdateSkyWalletReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var updateSkyWalletRequestBody UpdateSkyWalletRequestBody

	vars := mux.Vars(r)

  name, ok := vars["name"]
  if !ok {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_PARAMS_WALLET_NAME_MISSING, "Wallet Name Missing"))
    return
  }

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&updateSkyWalletRequestBody)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  updateSkyWalletRequest := UpdateSkyWalletRequest{
    Name: name,
    Private: updateSkyWalletRequestBody.Private,
  }

  err = updateSkyWalletRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  SuccessResponse(ctx, w, nil)
}
