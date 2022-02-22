package api

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/fix"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type FixAllRequest struct {
  Name string `json:"name"`
}

type FixRequest struct {
  Name string `json:"name"`
  PownItems []PownItem `json:"pown_items"`
  Tickets [][]string `json:"tickets"`
}

type PownItem struct {
  Sn int `json:"sn"`
  PownString string `json:"pownstring"`
}

func (s *FixRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.PownItems, validation.Required, validation.Each(validation.By(ValidatePownItem))),
    validation.Field(&s.Tickets, validation.Each(validation.By(ValidateTickets))),
    validation.Field(&s.Name, validation.Required, validation.Length(config.WALLET_NAME_MIN_LENGTH, config.WALLET_NAME_MAX_LENGTH)),
  )

  return err
}

func (s *FixAllRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Name, validation.Required, validation.Length(config.WALLET_NAME_MIN_LENGTH, config.WALLET_NAME_MAX_LENGTH)),
  )

  return err
}

func ValidatePownItem(v interface{}) error {
  pi := v.(PownItem)

  err := validation.ValidateStruct(&pi,
    validation.Field(&pi.Sn, validation.Required, validation.Min(1), validation.Max(config.TOTAL_COINS)),
    validation.Field(&pi.PownString, validation.Required, validation.Match(regexp.MustCompile("^[upfne]{" + strconv.Itoa(config.TOTAL_RAIDA_NUMBER) + "}$"))),
  )

  return err
}

func ValidateTickets(v interface{}) error {
  rs := v.([]string)

  err := validation.Validate(rs, validation.Length(config.TOTAL_RAIDA_NUMBER, config.TOTAL_RAIDA_NUMBER), validation.Each(validation.Match(regexp.MustCompile("^[a-fA-F0-9]{8}$"))))

  return err
}



type FixCall struct {
  Parent
  Args *FixRequest
  Wallet *wallets.Wallet
}

func NewFix(progressChannel chan interface{}, args *FixRequest) *FixCall {
  return &FixCall{
    *NewParent(progressChannel),
    args,
    nil,
  }
}

type FixAllCall struct {
  Parent
  Args *FixAllRequest
  Wallet *wallets.Wallet
}

func NewFixAll(progressChannel chan interface{}, args *FixAllRequest) *FixAllCall {
  return &FixAllCall{
    *NewParent(progressChannel),
    args,
    nil,
  }
}

func (v *FixCall) GetStrideSize() int {
  // packetSize = 1024 - header(22) - challenge(16) - signature(2)
  coinsLen := 1024 - 22 - 16 - 2

  // owner
  coinsLen -= 3

  // receiptID
  coinsLen -= 16

  // ts
  coinsLen -= 6

  // coinType
  coinsLen -= 1


  // TY
  coinsLen -= 1

  // RAID TYPE is a part of memo
  coinsLen -= config.MAX_MEMO_LENGTH
  
  // Sn + An
  coinItemSize := 3 + 16

  return coinsLen / coinItemSize
}

func (v *FixCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("Fixing Coins")

  /* Lite validation */
  wallet, err := storage.GetDriver().GetWalletWithContents(ctx, v.Args.Name)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.Wallet = wallet

  plen := len(v.Args.PownItems)

  logger.L(ctx).Debugf("Fixing %d coins (%d coins per packet)", plen, v.GetStrideSize())

  fbatches := make([]fix.FixBatch, plen)
  for bn, _ := range(fbatches) {
    fbatches[bn].CoinsPerRaida = make(map[int][]cloudcoin.CloudCoin, 0)
  }

  var batchNumber int
  for idx, pi := range(v.Args.PownItems) {

    cc := cloudcoin.NewFromData(uint32(pi.Sn))
    cc.SetPownString(pi.PownString)

    batchNumber = idx / v.GetStrideSize()

    logger.L(ctx).Debugf("Will try to fix coin %d: %s. Batch %d", cc.Sn, cc.PownString, batchNumber)
    for ridx, status := range(cc.Statuses) {
      logger.L(ctx).Debugf("coin %d r%d st=%d", cc.Sn, ridx, status)

      if status == config.COIN_STATUS_COUNTERFEIT {
        logger.L(ctx).Debugf("coin %d failed on raida %d", cc.Sn, ridx)
        _, ok := fbatches[batchNumber].CoinsPerRaida[ridx]
        if !ok {
          fbatches[batchNumber].CoinsPerRaida[ridx] = make([]cloudcoin.CloudCoin, 0)
        }

        fbatches[batchNumber].CoinsPerRaida[ridx] = append(fbatches[batchNumber].CoinsPerRaida[ridx], *cc)
      }

    }

    if (batchNumber < len(v.Args.Tickets)) {
      fbatches[batchNumber].Tickets = v.Args.Tickets[batchNumber]
    }

    logger.L(ctx).Debugf("ccc %v", fbatches)
  }


  for bn, _ := range(fbatches) {
    logger.L(ctx).Debugf("batch #%d", bn)
    for ridx, _ := range(fbatches[bn].CoinsPerRaida) {
      logger.L(ctx).Debugf("raida%d", ridx)
      for _, cc := range(fbatches[bn].CoinsPerRaida[ridx]) {
        logger.L(ctx).Debugf("sn %d", cc.Sn) 
      }
    }
  }

  fixer, _ := fix.New(v.ProgressChannel)
  fr, err := fixer.Fix(ctx, wallet, fbatches, false)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(fr)
}

func (v *FixAllCall) Run(ctx context.Context) {
  logger.L(ctx).Debugf("Fixing All Coins in the Fracked")

  /* Lite validation */
  wallet, err := storage.GetDriver().GetWalletWithContents(ctx, v.Args.Name)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.Wallet = wallet
  fixer, _ := fix.New(v.ProgressChannel)

  fbatches := make([]fix.FixBatch, 0)
  idx := 0
  var batchNumber int
  for _, cdns := range(wallet.CoinsByDenomination) {
    for _, cc := range(cdns) {
      logger.L(ctx).Debugf("Checking if we need to fix coin %d: %s. Batch %d %s", cc.Sn, cc.PownString, batchNumber, cc.GetPownString())
      err := storage.GetDriver().ReadCoin(ctx, wallet, cc)
      if err != nil {
        logger.L(ctx).Warnf("Failed to read coin %d, won't fix it: %s", cc.Sn, err.Error())
        continue
      }

      if cc.GetGradeStatus() != config.COIN_STATUS_FRACKED {
        continue
      }

      if cc.GetLocationStatus() != config.COIN_LOCATION_STATUS_FRACKED {
        continue
      }

      batchNumber = idx / fixer.GetStrideSize()
      idx++

      logger.L(ctx).Debugf("%d/%d Will try to fix coin %d: %s. Batch %d %s", idx, fixer.GetStrideSize(), cc.Sn, cc.PownString, batchNumber, cc.GetPownString())
      for ridx, status := range(cc.Statuses) {
        logger.L(ctx).Debugf("coin %d r%d st=%d", cc.Sn, ridx, status)

        if status == config.COIN_STATUS_COUNTERFEIT {
          logger.L(ctx).Debugf("coin %d failed on raida %d", cc.Sn, ridx)

          if len(fbatches) <= batchNumber {
            fbatches = append(fbatches, fix.FixBatch{})
            fbatches[batchNumber].CoinsPerRaida = make(map[int][]cloudcoin.CloudCoin, 0)
          }

          _, ok := fbatches[batchNumber].CoinsPerRaida[ridx]
          if !ok {
            fbatches[batchNumber].CoinsPerRaida[ridx] = make([]cloudcoin.CloudCoin, 0)
          }

          fbatches[batchNumber].CoinsPerRaida[ridx] = append(fbatches[batchNumber].CoinsPerRaida[ridx], *cc)
        }
      }
    }
  }

  for bn, _ := range(fbatches) {
    logger.L(ctx).Debugf("batch #%d", bn)
    for ridx, _ := range(fbatches[bn].CoinsPerRaida) {
      logger.L(ctx).Debugf("raida%d", ridx)
      for _, cc := range(fbatches[bn].CoinsPerRaida[ridx]) {
        logger.L(ctx).Debugf("sn %d", cc.Sn) 
      }
    }
  }

  fr, err := fixer.Fix(ctx, wallet, fbatches, false)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(fr)
}


func FixReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()

	var fixRequest FixRequest

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&fixRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = fixRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewFix(task.ProgressChannel, &fixRequest)

  task.Run(ctx, instance)
  
  SuccessResponse(ctx, w, task)
  
}

func FixAllReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	var fixAllRequest FixAllRequest

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&fixAllRequest)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  err = fixAllRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewFixAll(task.ProgressChannel, &fixAllRequest)
  task.Run(ctx, instance)
  
  SuccessResponse(ctx, w, task)
  
}
