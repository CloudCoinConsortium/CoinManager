package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/transactions"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/unpacker"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	//"github.com/go-ozzo/ozzo-validation/v4/is"
)

type ImportRequest struct {
  Name string `json:"name"`
	Items []ImportItem `json:"items"`
  Tag string `json:"tag"`
}

type ImportResponse struct {
  //PownResult *raida.PownOutput `json:"pown_results"`
  PownResult *raida.PangOutput `json:"pown_results"`
  ReceiptID string `json:"receipt_id"`
}

type ImportItem struct {
  Type string `json:"type"`
  Data string `json:"data,omitempty"`
}

func (s *ImportRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.By(ValidateWallet)),
    validation.Field(&s.Items, validation.Required, validation.Each(validation.By(ValidateImportItem))),
    validation.Field(&s.Tag, validation.By(ValidateTag)),
  )

  return err
}

func ValidateImportItem(v interface{}) error {
  importItem := v.(ImportItem)

  err := validation.ValidateStruct(&importItem,
    validation.Field(&importItem.Type, validation.Required, validation.In(config.IMPORT_TYPE_FILE, config.IMPORT_TYPE_INLINE, config.IMPORT_TYPE_SUSPECT)),
    validation.Field(&importItem.Data, validation.When(importItem.Type == config.IMPORT_TYPE_INLINE, is.Base64, validation.Required)),
    validation.Field(&importItem.Data, validation.When(importItem.Type == config.IMPORT_TYPE_FILE, validation.Length(config.MIN_FILE_LENGTH, config.MAX_FILE_LENGTH), validation.Required)),
    validation.Field(&importItem.Data, validation.When(importItem.Type == config.IMPORT_TYPE_SUSPECT, validation.Empty)),
  )

  return err
}

type ImportCall struct {
  Parent
  Args *ImportRequest
  Wallet *wallets.Wallet
  Task *tasks.Task
}

func NewImport(progressChannel chan interface{}, args *ImportRequest) *ImportCall {
  return &ImportCall{
    *NewParent(progressChannel),
    args,
    nil,
    nil,
  }
}

func (v *ImportCall) BatchFunction(ctx context.Context, coins []cloudcoin.CloudCoin) error {
  logger.L(ctx).Debugf("Import. Calling batch function for %d notes", len(coins))

  err := storage.GetDriver().UpdateStatus(ctx, v.Wallet, coins)
  if err != nil {
    logger.L(ctx).Debugf("Failed to set status for coins len=%d: %s", len(coins), err.Error())
    return err
  }

  return nil
}

func (v *ImportCall) ImportLeftOvers(ctx context.Context, wallet *wallets.Wallet) {
  logger.L(ctx).Debugf("import leftovers")


  coins, ignoredCoins, err := storage.GetDriver().ReadCoinsInLocation(ctx, wallet, config.COIN_LOCATION_STATUS_SUSPECT)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  if len(coins) == 0 {
    storage.GetDriver().EmptyLocation(ctx, wallet, config.COIN_LOCATION_STATUS_SUSPECT)
    if ignoredCoins > 0 {
      v.SendError(ctx, perror.New(perror.ERROR_NO_COINS, "No valid coins to import. Total skipped corrupted coins: " + strconv.Itoa(ignoredCoins) + ". They were moved to the Trash"))
    } else {
      v.SendError(ctx, perror.New(perror.ERROR_NO_COINS, "No coins to import"))
    }
    return
  }

  logger.L(ctx).Debugf("Ready to pown %d coins", len(coins))

  ncoins := make([]cloudcoin.CloudCoin, 0)
  duplicates := 0
  ids := 0
  for _, coin := range(coins) {
    if storage.GetDriver().CoinExistsInTheWalletAndAuthentic(ctx, wallet, &coin) {
      duplicates += coin.GetDenomination()
      continue
    }

    if coin.IsIDCoin() {
      ids += coin.GetDenomination()

      logger.L(ctx).Debugf("Coin %d is an ID coin and will not be imported", coin.Sn)
      continue
    }

    ncoins = append(ncoins, coin)
  }

  for idx, _ := range(ncoins) {
    logger.L(ctx).Debugf("Ready to pown coin %d ", ncoins[idx].Sn)
    for i := 0; i < config.TOTAL_RAIDA_NUMBER; i++ {
      if ncoins[idx].Pans[i] == "" {
        ncoins[idx].Pans[i], _ = cloudcoin.GeneratePan()
      }
    }
  }

  v.SendProgress("Powning", 0)
  s := raida.NewPown(v.ProgressChannel)
  s.SetBatchFunction(v.BatchFunction)
  
  if (v.Task != nil) {
    iterations := v.GetTotalIterations(len(ncoins), s.GetStrideSize())

    v.Task.SetTotalIterations(iterations)
    logger.L(ctx).Debugf("Set iterfations for %d coins to %d", len(ncoins), iterations)
  }

  response, err := s.Pown(ctx, ncoins)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  receiptID, _ := utils.GenerateReceiptID()

  // Amount is zero initally. Will be updated later
  t := transactions.New(0, "Leftovers in Suspect", "Import", receiptID)
  amount := 0
  for _, cc := range(ncoins) {
    ok, _, _ := cc.IsAuthentic()
    if ok {
      amount += cc.GetDenomination()
    }

    t.AddDetail(cc.Sn, cc.GetPownString(), cc.GetGradeStatusString())
    logger.L(ctx).Debugf("cc %s %d", cc.GetGradeStatusString(), cc.GetGradeStatus())
  }

  t.Amount = amount
  logger.L(ctx).Debugf("amount %d", amount)
  err = storage.GetDriver().AppendTransaction(ctx, wallet, t)
  if err != nil {
    logger.L(ctx).Warnf("Failed to save transaction for wallet %s:%s", wallet.Name, err.Error())
    // Ignoring the error above
  }

  response.TotalAlreadyExists = duplicates
  response.TotalUnknown += ignoredCoins


  ir := &ImportResponse{}
  //ir.PownResult = response
  ir.ReceiptID = receiptID
  v.SendResult(ir)
  

}


func (v *ImportCall) Run(ctx context.Context) {
  var coins []cloudcoin.CloudCoin

  logger.L(ctx).Debugf("Importing Coins")

  /* Lite validation */
  wallet, err := storage.GetDriver().GetWalletWithContents(ctx, v.Args.Name)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.Wallet = wallet
  v.SendProgress("Unpacking Coins", 0)

  coins = make([]cloudcoin.CloudCoin, 0)
  unpacker := unpacker.New()
  items := v.Args.Items

  var needSuspect = false
  var others = false
  for _, item := range (items) {
    if item.Type == config.IMPORT_TYPE_SUSPECT {
      needSuspect = true
    } else {
      others = true
    }
  }

  if needSuspect {
    if others {
      v.SendError(ctx, perror.New(perror.ERROR_MIX_SUSPECT_WITH_OTHERS, "Suspect item can't be mixed with others"))
      return
    }

    logger.L(ctx).Debugf("Doing leftovers")
    v.ImportLeftOvers(ctx, wallet)

    return
  }



  files := make([]string, 0)
  for _, item := range (items) {
    var decoded []byte
    var err error

    if item.Type == config.IMPORT_TYPE_INLINE {
      decoded, err = base64.StdEncoding.DecodeString(item.Data)
      if err != nil {
        logger.L(ctx).Errorf("Failed to decode base64: %s", err.Error())
        v.SendError(ctx, perror.New(perror.ERROR_DECODE_BASE64, "Failed to decode base64: " + err.Error()))
        return
      }


    } else if item.Type == config.IMPORT_TYPE_FILE {
      cfile, err := ioutil.ReadFile(item.Data)
      if err != nil {
        logger.L(ctx).Errorf("Failed to read file %s: %s", item.Data, err.Error())
        v.SendError(ctx, perror.New(perror.ERROR_FILESYSTEM, "Failed to read file: " + err.Error()))
        return
      }

      decoded = []byte(cfile)
      files = append(files, item.Data)
      //err = storage.GetDriver().PutInImport(wallet, coins)
      /*
      err = storage.GetDriver().PutInImport(wallet, item.Data)
      if err != nil {
          v.SendError(ctx, err)
          return
      }
      */

    } else {
      logger.L(ctx).Debugf("Invalid Coin Type %s. Skipping it", item.Type)
      continue
    }

    logger.L(ctx).Debugf("Decodes len %d", len(decoded))
    v.SendProgress("Unpacking Coins", 0)

    lcoins, err := unpacker.Unpack(ctx, decoded)
    if err != nil {
        v.SendError(ctx, perror.New(perror.ERROR_UNPACK, "Failed to unpack a coin: " + err.Error()))
        return
    }

    coins = append(coins, lcoins...)
  }

  logger.L(ctx).Debugf("Unpacked %d coins", len(coins))
  v.SendProgress("Placing Coins", 0)
  
 // v.SendResult(nil)
 // return
  

  logger.L(ctx).Debugf("Total coins before checking for duplicates and ID coins: %d", len(coins))



  ncoins := make([]cloudcoin.CloudCoin, 0)
  duplicates := 0
  ids := 0
  for _, coin := range(coins) {
    if storage.GetDriver().CoinExistsInTheWallet(ctx, wallet, &coin) {
      duplicates += coin.GetDenomination()
      continue
    }

    if coin.IsIDCoin() {
      ids += coin.GetDenomination()

      logger.L(ctx).Debugf("Coin %d is an ID coin and will not be imported", coin.Sn)
      continue
    }

    ncoins = append(ncoins, coin)
  }

  logger.L(ctx).Debugf("Total coins after checking for duplicates: %d", len(ncoins))
  v.SendProgress("Generating Pans", 0)
  for idx, cc := range(ncoins) {
    logger.L(ctx).Debugf("Ready to pown coin %d %p %p", cc.Sn, &cc, &ncoins[idx])
    ncoins[idx].GenerateMyPans()
  }

  err = storage.GetDriver().PutInSuspect(ctx, wallet, ncoins)
  if err != nil {
      v.SendError(ctx, err)
      return
  }

  for _, file := range(files) {
    logger.L(ctx).Debugf("Moving original file to the Imported: %s", file)
    err = storage.GetDriver().PutInImported(ctx, wallet, file)
    if err != nil {
        logger.L(ctx).Warnf("Failed to move file to imported: %s", err.Error())
    }
  }

  v.SendProgress("Powning", 0)
  //s := raida.NewPown(v.ProgressChannel)
  s := raida.NewPang(v.ProgressChannel)
  s.SetBatchFunction(v.BatchFunction)
  
  if (v.Task != nil) {
    iterations := v.GetTotalIterations(len(ncoins), s.GetStrideSize())

    v.Task.SetTotalIterations(iterations)
    logger.L(ctx).Debugf("Set iterfations for %d coins to %d", len(ncoins), iterations)
  }

  //response, err := s.Pown(ctx, ncoins)
  response, err := s.Pang(ctx, ncoins)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  response.TotalAlreadyExists = duplicates
  response.TotalUnknown += ids


  // Adding Transaction
  receiptID, _ := utils.GenerateReceiptID()

  // Amount is zero initally. Will be updated later
  t := transactions.New(0, "" + v.Args.Tag, "Import", receiptID)
  amount := 0
  for _, cc := range(ncoins) {
    ok, _, _ := cc.IsAuthentic()
    if ok {
      amount += cc.GetDenomination()
    }

    t.AddDetail(cc.Sn, cc.GetPownString(), cc.GetGradeStatusString())
    logger.L(ctx).Debugf("cc status=%s status(num)=%d", cc.GetGradeStatusString(), cc.GetGradeStatus())
  }

  t.Amount = amount

  logger.L(ctx).Debugf("auth %d", amount)
  err = storage.GetDriver().AppendTransaction(ctx, wallet, t)
  if err != nil {
    logger.L(ctx).Warnf("Failed to save transaction for wallet %s:%s", wallet.Name, err.Error())
    // Ignoring the error above
  }


  ir := &ImportResponse{}
  ir.PownResult = response
  ir.ReceiptID = receiptID

  v.SendResult(ir)
}


func ImportReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var importRequest ImportRequest

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&importRequest)
  if err != nil {
    logger.L(ctx).Debugf("err=%s",err)
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = importRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewImport(task.ProgressChannel, &importRequest)
  instance.Task = task
  

  task.Run(ctx, instance)

  SuccessResponse(ctx, w, task)
}

