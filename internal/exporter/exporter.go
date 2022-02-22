package exporter

import (
	"archive/zip"
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/core"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/transactions"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/worker"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

type Exporter struct {
  worker.Worker
  font *truetype.Font
  rfont *truetype.Font
  progressChannel chan interface{}
}

func New(progressChannel chan interface{}) (*Exporter, error) {

  rFontPath := core.GetTemplateDirPath() + config.Ps() + "Barlow-Regular.ttf"
  fontPath := core.GetTemplateDirPath() + config.Ps() + "Barlow-Bold.ttf"

  fontBytes, err := ioutil.ReadFile(fontPath)
  if err != nil {
    return nil, err
  }

  font, err := freetype.ParseFont(fontBytes)
  if err != nil {
    return nil, err
  }

  fontBytes, err = ioutil.ReadFile(rFontPath)
  if err != nil {
    return nil, err
  }

  rfont, err := freetype.ParseFont(fontBytes)
  if err != nil {
    return nil, err
  }

  return &Exporter{
    *worker.New(progressChannel),
    font,
    rfont,
    progressChannel,
  }, nil

}


func (v *Exporter) Export(ctx context.Context, wallet *wallets.Wallet, amount int, tag string, path string, etype string) ([]byte, error) {
  logger.L(ctx).Debugf("Exporting %d coins from %s", amount, wallet.Name)

  if err := v.ValidateFolder(ctx, path); err != nil {
    return nil, err
  }

  if (amount > wallet.Balance) {
    logger.L(ctx).Debugf("Not enough coins")
    return nil, perror.New(perror.ERROR_NOT_ENOUGH_COINS, "Not enough CloudCoins")
  }

  coinsToExport, err := v.GetCoinsToDealWith(ctx, wallet, amount)
  if err != nil {
    logger.L(ctx).Errorf("Failed to pick coins to export: %s", err.Error())
    return nil, err
  }

  logger.L(ctx).Debugf("picked %d coins", len(coinsToExport))

  bcc := make([]byte, 0)
  headerAdded := false
  for _, cc := range(coinsToExport) {
    if !headerAdded {
      header, err := cc.GetHeader()
      if err != nil {
        logger.L(ctx).Debugf("Failed to get header for coin %d", cc.Sn)
        return nil, err
      }
      bcc = append(bcc, header...)
      headerAdded = true
    }

    bdata, err := cc.GetContentData()
    if err != nil {
      logger.L(ctx).Debugf("Failed to get data for coin %d", cc.Sn)
      return nil, err
    }

    bcc = append(bcc, bdata...)
  }

  var ddata []byte

  ext := "bin"
  if (etype == config.EXPORT_TYPE_BIN) {
    ddata = bcc
  } else if (etype == config.EXPORT_TYPE_PNG) {
    amountStr := v.NumberFormat(int64(amount))
    logger.L(ctx).Debugf("total %s", amountStr)

    textAmount := utils.Convert(amount)
    ddata, err = v.DrawData(ctx, amountStr, textAmount, bcc)
    if err != nil {
      logger.L(ctx).Debugf("Failed to draw stack: %s", err.Error())
      return nil, perror.New(perror.ERROR_FAILED_TO_DRAW_FILE, "Failed to draw file: " + err.Error())
    }

    ext = "png"
  } else if (etype == config.EXPORT_TYPE_ZIP) {
    logger.L(ctx).Debugf("Zipping it")

    ddata, err = v.GetZip(ctx, coinsToExport)
    if err != nil {
      logger.L(ctx).Debugf("Failed to zip stack: %s", err.Error())
      return nil, perror.New(perror.ERROR_ZIP, "Failed to zip file: " + err.Error())
    }

    ext = "zip"
  } else {
    logger.L(ctx).Errorf("Export type %s is not recognized", etype)
    return nil, perror.New(perror.ERROR_INVALID_EXPORT_TYPE, "Invalid export type")
  }


  var fname string
  if len(coinsToExport) == 1 {
    cc := coinsToExport[0]
    fname = strconv.Itoa(int(cc.Sn)) + ".CloudCoin." + strconv.Itoa(int(cc.GetCoinID())) + "." + tag + "." + ext
  } else {
    fname = strconv.Itoa(amount) + ".CloudCoin." + tag + "." + ext
  }

  fpath := path + config.Ps() + fname

  logger.L(ctx).Debugf("Writing to file %s", fpath)
  err = os.WriteFile(fpath, ddata, 0644) 
  if err != nil {
    logger.L(ctx).Errorf("Failed to save file %s: %s", fname, err.Error())
    return nil, perror.New(perror.ERROR_SAVE_COIN, "Failed to save file " + fname + ": " + err.Error())
  }


  // Moving coin to 'Sent'
  err = storage.GetDriver().SetLocation(ctx, wallet, coinsToExport, config.COIN_LOCATION_STATUS_SENT)
  if err != nil {
    logger.L(ctx).Errorf("Failed to move coin to Sent: %s", err.Error())
    return nil, perror.New(perror.ERROR_UPDATE_COIN_STATUS, "Failed to update coin status: " + err.Error())
  }

  // Adding Transaction
  receiptID, _ := utils.GenerateReceiptID()
  t := transactions.New(-1 * amount, tag, "Export", receiptID)
  for _, cc := range(coinsToExport) {
    cc.Grade()
    t.AddDetail(cc.Sn, cc.GetPownString(), cc.GetGradeStatusString())
  }

  err = storage.GetDriver().AppendTransaction(ctx, wallet, t)
  if err != nil {
    logger.L(ctx).Warnf("Failed to save transaction for wallet %s:%s", wallet.Name, err.Error())
    return nil, perror.New(perror.ERROR_SAVE_TRANSACTION, "Failed to save transaction: " + err.Error())
  }


  return ddata, nil
}

func (v *Exporter) GetZip(ctx context.Context, ccs []cloudcoin.CloudCoin) ([]byte, error) {

  buf := new(bytes.Buffer)
  zipWriter := zip.NewWriter(buf)

  for _, cc := range(ccs) {
    amount := cc.GetDenomination()
    amountStr := v.NumberFormat(int64(amount))
    logger.L(ctx).Debugf("getting png sn#%d denomination %d", cc.Sn, amount)

    textAmount := utils.Convert(amount)
    bcc, _ := cc.GetContentData()

    ddata, err := v.DrawData(ctx, amountStr, textAmount, bcc)
    if err != nil {
      logger.L(ctx).Debugf("Failed to draw stack: %s", err.Error())
      return nil, perror.New(perror.ERROR_FAILED_TO_DRAW_FILE, "Failed to draw file: " + err.Error())
    }

    zipFile, err := zipWriter.Create(cc.GetName() + ".png")
    if err != nil {
      logger.L(ctx).Debugf("Failed to create zip subfile %s %s ", cc.GetName(), err.Error())
      return nil, err
    }

    _, err = zipFile.Write(ddata)
    if err != nil {
      logger.L(ctx).Debugf("Failed to write zip subfile %s", cc.GetName(), err.Error())
      return nil, err
    }
  }

  err := zipWriter.Close()
  if err != nil {
    logger.L(ctx).Debugf("Failed to close zip: %s", err.Error())
    return nil, err
  }

  return buf.Bytes(), nil
}


func (v *Exporter) NumberFormat(n int64) string {
  in := strconv.FormatInt(n, 10)
  numOfDigits := len(in)
  if n < 0 {
    numOfDigits-- // First character is the - sign (not a digit)
  }
  numOfCommas := (numOfDigits - 1) / 3

  out := make([]byte, len(in) + numOfCommas)
  if n < 0 {
    in, out[0] = in[1:], '-'
  }

  for i, j, k := len(in)-1, len(out) - 1, 0; ; i, j = i - 1, j - 1 {
    out[j] = in[i]
    if i == 0 {
        return string(out)
    }
    if k++; k == 3 {
        j, k = j - 1, 0
        out[j] = ','
    }
  }
}
