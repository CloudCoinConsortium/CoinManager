package backup

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/worker"
)

type Backup struct {
  worker.Worker
  progressChannel chan interface{}
  Wallet *wallets.Wallet
  Task *tasks.Task
}

type BackupResult struct {
  TotalCoins int
  FileName string
}

func New(progressChannel chan interface{}, task *tasks.Task) (*Backup) {
  return &Backup{
		*worker.New(progressChannel),
    progressChannel,
    nil,
    task,
  }

}

func (v *Backup) Backup(ctx context.Context, wallet *wallets.Wallet, path string, tag string) (*BackupResult, error) {
  logger.L(ctx).Debugf("Backup coins from %s", wallet.Name)

  if err := v.ValidateFolder(ctx, path); err != nil {
    return nil, err
  }

  if tag == "" {
    tag = "CC"
  }

  buf := new(bytes.Buffer)
  zipWriter := zip.NewWriter(buf)

  amount := 0
  for _, cdns := range(wallet.CoinsByDenomination) {
    for _, cc := range(cdns) {
      x := fmt.Sprintf("Coin #%d", cc.Sn)
      v.SendProgress(x)

      err := storage.GetDriver().ReadCoin(ctx, wallet, cc)
      if err != nil {
        logger.L(ctx).Debugf("Failed to read coin %d: %s", cc.Sn, err.Error())
        continue
      }

      zipFile, err := zipWriter.Create(cc.GetName() + config.CC_FILE_BINARY_EXTENSION)
      if err != nil {
        logger.L(ctx).Debugf("Failed to create zip subfile %s %s ", cc.GetName(), err.Error())
        return nil, err
      }

      ddata, _ := cc.GetData()
      _, err = zipFile.Write(ddata)
      if err != nil {
        logger.L(ctx).Debugf("Failed to write zip subfile %s", cc.GetName(), err.Error())
        return nil, err
      }

      amount += cc.GetDenomination()
    }
  }

  err := zipWriter.Close()
  if err != nil {
    logger.L(ctx).Debugf("Failed to close zip: %s", err.Error())
    return nil, err
  }

  data := buf.Bytes()

  ts := time.Now()
  sts := ts.Format("2006-01-02-15-04-05")
  fname := strconv.Itoa(amount) + ".Backup." + tag + "." + sts + ".zip"
  fpath := path + config.Ps() + fname

  logger.L(ctx).Debugf("Backupping to %s", fpath)

  err = os.WriteFile(fpath, data, 0644)
  if err != nil {
    logger.L(ctx).Errorf("Failed to write zip file: %s", err.Error())
    return nil, perror.New(perror.ERROR_WRITE_FILE, "Failed to write zip file: %s" + err.Error())
  }

  backupResult := &BackupResult{}
  backupResult.TotalCoins = amount
  backupResult.FileName = fpath

  return backupResult, nil
}

func (v *Backup) SkyBackup(ctx context.Context, wallet *wallets.SkyWallet, path string, tag string) (string, error) {
  logger.L(ctx).Debugf("Backup ID coin %s", wallet.Name)
  if err := v.ValidateFolder(ctx, path); err != nil {
    return "", err
  }

  if tag == "" {
    tag = "CC"
  }

  ts := time.Now()
  sts := ts.Format("2006-01-02-15-04-05")
  fname := wallet.Name + ".Backup." + tag + "." + sts + config.CC_FILE_BINARY_EXTENSION
  fpath := path + config.Ps() + fname

  logger.L(ctx).Debugf("Backup to %s", fpath)

  data, _ := wallet.IDCoin.GetData()

  err := os.WriteFile(fpath, data, 0644)
  if err != nil {
    logger.L(ctx).Errorf("Failed to write file: %s", err.Error())
    return "", err
  }


  return fpath, nil
}

