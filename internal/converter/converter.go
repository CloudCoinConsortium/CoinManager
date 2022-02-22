package converter

import (
	"context"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/legacyraida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/transactions"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/worker"
)

type Convert struct {
  worker.Worker
  progressChannel chan interface{}
  Wallet *wallets.Wallet
  Task *tasks.Task
}

type ConvertResult struct {
  TotalCoins int
  TotalConverted int
  TotalTicketFailed int
  TotalConvertFailed int
  Details [][]string
  Coins []raida.CoinOutput
}

func New(progressChannel chan interface{}, task *tasks.Task) (*Convert) {
  return &Convert{
		*worker.New(progressChannel),
    progressChannel,
    nil,
    task,
  }

}

func (v *Convert) Convert(ctx context.Context, coins []cloudcoin.CloudCoin, wallet *wallets.Wallet) (*ConvertResult, error) {
  logger.L(ctx).Debugf("Convertring %d coins to %s", len(coins), wallet.Name)

  v.Wallet = wallet
  stride := config.MAX_NOTES_TO_SEND
  
  cr := &ConvertResult{}
  for i := 0; i < len(coins); i += stride {
    max := i + stride
    if max > len(coins) {
      max = len(coins)
    }

    logger.L(ctx).Debugf("coins slice %d:%d, Processing %d notes", i, max, (max - i))
    response, err := v.ConvertBatch(ctx, coins[i:max], wallet)
    if err != nil {
      logger.L(ctx).Errorf("Failed to process batch: %s", err.Error())
      for _, lcc := range(coins[i:max]) {
        cr.TotalTicketFailed += lcc.GetDenomination()
      }

      continue
    }

    cr.TotalCoins += response.TotalCoins
    cr.TotalConvertFailed += response.TotalConvertFailed
    cr.TotalConverted += response.TotalConverted
    cr.TotalTicketFailed += response.TotalTicketFailed
  }
  
  return cr, nil
}


func (v *Convert) ConvertBatch(ctx context.Context, ccs []cloudcoin.CloudCoin, wallet *wallets.Wallet) (*ConvertResult, error) {
  logger.L(ctx).Debugf("Convertring Batch of %d coins to %s", len(ccs), wallet.Name)

  receiptID, _ := utils.GenerateReceiptID()

  lrgt := legacyraida.NewGetTicket(v.progressChannel)
  tickets, err := lrgt.GetTicket(ctx, ccs)
  if err != nil {
    logger.L(ctx).Errorf("Failed to get tickets: %s", err.Error())
    return nil, perror.New(perror.ERROR_GET_TICKET, "Failed to get tickets from legacy raida")
  }

  convertResult := &ConvertResult{}

  validCCs := make([]cloudcoin.CloudCoin, 0)
  for idx, _ := range(ccs) {
    isAuthentic, _, _ := ccs[idx].IsAuthentic()
    convertResult.TotalCoins += ccs[idx].GetDenomination()
    if !isAuthentic {
      logger.L(ctx).Debugf("Coin %d is counterfeit %s", ccs[idx].Sn, ccs[idx].GetPownString())
      convertResult.TotalTicketFailed += ccs[idx].GetDenomination()
      continue
    }

    validCCs = append(validCCs, ccs[idx])

    logger.L(ctx).Debugf("coin %d will be converted", ccs[idx].Sn)
  }

  if len(validCCs) == 0 {
    logger.L(ctx).Errorf("Failed to get tickets for all coins. Zero tickets")
    return nil, perror.New(perror.ERROR_GET_TICKET, "Failed to get tickets from legacy raida for all coins. No valid tickets received from the Legacy Raida")
  }
  
  // Form new tmap
  tmap := make([]map[string][]cloudcoin.CloudCoin, config.TOTAL_RAIDA_NUMBER)
  for i := 0; i < config.TOTAL_RAIDA_NUMBER; i++ {
    tmap[i] = make(map[string][]cloudcoin.CloudCoin, 0)
  }

  for i := 0; i < config.TOTAL_RAIDA_NUMBER; i++ {
    rtmap := tickets.Tmap[i]
    for ticket, coins := range(rtmap) {
      // Iterate over coins and find if they are in the list
      for cidx, _ := range(coins) {
        for vcidx, _ := range(validCCs) {
          if coins[cidx].Sn == validCCs[vcidx].Sn {
            sn := validCCs[vcidx].Sn
            logger.L(ctx).Debugf("cc %d is ok with ticket %s on r%d", sn, ticket, i)
            _, ok := tmap[i][ticket]
            if !ok {
              tmap[i][ticket] = make([]cloudcoin.CloudCoin, 0)
            } else {
              logger.L(ctx).Debugf("Too many tickets for r%d. It is not supported in this version", i)
              return nil, perror.New(perror.ERROR_GET_TICKET, "Too many tickets are not supported in this version. Got more than one ticket in one batch for raida " + strconv.Itoa(i))
            }

            tmap[i][ticket] = append(tmap[i][ticket], validCCs[vcidx])
          }
        }
      }
    }
  }


  amount := 0
  for _, cc := range(ccs) {
    amount += cc.GetDenomination()
    convertResult.TotalConverted += cc.GetDenomination()
  }

  // Adding Transaction
  tag := "xxx"
  t := transactions.New(amount, "Converted: " + tag, "Convert", receiptID)
  for _, cc := range(ccs) {
    cc.Grade()
    t.AddDetail(cc.Sn, cc.GetPownString(), cc.GetGradeStatusString())
  }
  err = storage.GetDriver().AppendTransaction(ctx, wallet, t)
  if err != nil {
    logger.L(ctx).Warnf("Failed to save transaction for wallet %s:%s", wallet.Name, err.Error())
    return nil, perror.New(perror.ERROR_SAVE_TRANSACTION, "Failed to save transaction: " + err.Error())
  }

  return convertResult, nil
}
