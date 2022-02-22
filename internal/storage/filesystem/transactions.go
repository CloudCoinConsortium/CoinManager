package filesystem

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
  "context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/transactions"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
)

const TRANSACTIONS_FILENAME = "transactions.csv"

/* Interface Functions */

func (v *FileSystem) AppendTransaction(ctx context.Context, wallet *wallets.Wallet, t *transactions.Transaction) error {
  logger.L(ctx).Debugf("Appending Transaction %d: %s receiptID %s", t.Amount, t.Message, t.ReceiptID)
  //path := v.GetTransactionFilePath(wallet)
  err := v.InitTransactionsFile(ctx, wallet) 
  if err != nil {
    return err
  }
  
  logger.L(ctx).Debugf("Wallet Balance %d Amount %d", wallet.Balance, t.Amount)

  // Call it to possibly adjust incorrect balance
  _, err = v.GetTransactions(ctx, wallet)
  if err != nil {
    logger.L(ctx).Debugf("Failed to get transactions for %s", wallet.Name)
    return err
  }

  err = v.DoAppendTransaction(ctx, wallet, t)
  if err != nil {
    logger.L(ctx).Debugf("Failed to append transaction for %s", wallet.Name)
    return err
  }

  return nil
}


func (v *FileSystem) DoAppendTransaction(ctx context.Context, wallet *wallets.Wallet, t *transactions.Transaction) error {
  path := v.GetTransactionFilePath(wallet)

  timeStr := strconv.FormatInt(t.Datetime.Unix(), 10)

  message := strings.ReplaceAll(t.Message, ",", "\\,")
  s := "" + message + "," + timeStr + "," + strconv.Itoa(t.Amount) + "," + t.ReceiptID + "," + t.Type + v.Ls

  f, err := os.OpenFile(path, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
  if err != nil {
    logger.L(ctx).Errorf("Failed to open transactions file for %s:%s", wallet.Name, err.Error())
    return err
  }

  defer f.Close()

  _, err = f.WriteString(s)
  if err != nil {
    logger.L(ctx).Errorf("Failed to write string to file %s:%s", path, err.Error())
    return err
  }

  receiptsDir := v.GetReceiptsDir(wallet)
  trFile := receiptsDir + v.Ps + t.ReceiptID + ".txt"

  logger.L(ctx).Debugf("logging transaction details to file %s", trFile)

  ft, err := os.OpenFile(trFile, os.O_CREATE | os.O_WRONLY, 0644)
  if err != nil {
    logger.L(ctx).Errorf("Failed to open GUID file %s for %s:%s", trFile, wallet.Name, err.Error())
    return err
  }

  defer ft.Close()

  msg := ""
  for _, detail := range(t.Details) {
    msg += strconv.Itoa(int(detail.Sn)) + "," + strconv.Itoa(detail.Denomination) + "," + detail.PownString + "," + detail.Result + v.Ls
  }

  _, err = ft.WriteString(msg)
  if err != nil {
    logger.L(ctx).Errorf("Failed to write string to file %s:%s", trFile, err.Error())
    return err
  }

  return nil
}

func (v *FileSystem) GetReceipt(ctx context.Context, wallet *wallets.Wallet, guid string) ([]transactions.TransactionDetail, error) {

  tds := make([]transactions.TransactionDetail, 0)

  path := v.GetReceiptsDir(wallet) + v.Ps + guid + ".txt"

  _, err := os.Stat(path)
  if err != nil {
    if os.IsNotExist(err) {
      logger.L(ctx).Warnf("Receipt %s for wallet %s does not exist", guid, wallet.Name)
      return nil, perror.New(perror.ERROR_RECEIPT_NOT_EXISTS, "Receipt " + guid + " does not exist for wallet " + wallet.Name)
    }

    return nil, err
  }

  file, err := os.Open(path)
  if err != nil {
    logger.L(ctx).Errorf("Failed to open %s: %s", path, err.Error())
    return nil, perror.New(perror.ERROR_READ_TRANSACTIONS, "Failed to get receipt: " + err.Error())
  }

  defer file.Close()

  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    txt := scanner.Text()
    logger.L(ctx).Debugf("Read text %s", txt)

    parts := strings.Split(txt, ",")
    if (len(parts) != 4) {
      logger.L(ctx).Warnf("Invalid receipt record, skipping it: %s", txt)
      continue
    }

    td := transactions.TransactionDetail{}

    sn, _ := strconv.Atoi(parts[0])
    td.Sn = uint32(sn)
    td.Denomination, _ = strconv.Atoi(parts[1])
    td.PownString = parts[2]
    td.Result = parts[3]

    tds = append(tds, td)
  }

	return tds, nil
}

func (v *FileSystem) DeleteTransactionsAndReceipts(ctx context.Context, wallet *wallets.Wallet) error {
  logger.L(ctx).Debugf("Deleting Transactions For Wallet %s", wallet.Name)
  path := v.GetTransactionFilePath(wallet)

  err := os.Remove(path)
  if err != nil {
    logger.L(ctx).Errorf("Failed to delete transactions: %s", err.Error())
    return perror.New(perror.ERROR_DELETE_TRANSACTIONS, "Failed to delete transactions: " + err.Error())
  }


  trpath := v.GetReceiptsDir(wallet)
  err = os.RemoveAll(trpath)
  if err != nil {
    logger.L(ctx).Errorf("Failed to delete transactions: %s", err.Error())
    return perror.New(perror.ERROR_DELETE_RECEIPTS, "Failed to delete receipts: " + err.Error())
  }

  // Init it from scratch
  err = v.InitTransactionsFile(ctx, wallet) 
  if err != nil {
    return err
  }

  return nil

}

func (v *FileSystem) GetTransactions(ctx context.Context, wallet *wallets.Wallet) ([]transactions.Transaction, error) {
  logger.L(ctx).Debugf("Getting Transactions For Wallet %s", wallet.Name)
  path := v.GetTransactionFilePath(wallet)

  file, err := os.Open(path)
  if err != nil {
    logger.L(ctx).Errorf("Failed to open %s: %s", path, err.Error())
    return nil, perror.New(perror.ERROR_READ_TRANSACTIONS, "Failed to get transactions: " + err.Error())
  }

  defer file.Close()


  placeholder := "**********"

  runningBalance := 0
  trs := make([]transactions.Transaction, 0)
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    txt := scanner.Text()

    // GoLang Regxes don't support looks behind, so we do a workaround
    txt = strings.ReplaceAll(txt, "\\,", placeholder)

    parts := strings.Split(txt, ",")
    if (len(parts) < 5) {
      logger.L(ctx).Debugf("Invalid transaction record, skipping it: %s", txt)
      continue
    }

    tr := &transactions.Transaction{}
    message := parts[0]

    tr.Message = strings.ReplaceAll(message, placeholder, ",")
    ts, _ := strconv.Atoi(parts[1])
    datetime := time.Unix(int64(ts), 0)
    tr.Datetime = datetime
    tr.Amount, _ = strconv.Atoi(parts[2])
    tr.ReceiptID = parts[3]
    tr.Type = parts[4]

    runningBalance += tr.Amount
    tr.RunningBalance = runningBalance

    trs = append(trs, *tr)
  }


  if err := scanner.Err(); err != nil {
    logger.L(ctx).Errorf("Failed to read %s: %s", path, err.Error())
    return nil, perror.New(perror.ERROR_READ_TRANSACTIONS, "Failed to scan transactions: " + err.Error())
  }

  if runningBalance != wallet.Balance {
    logger.L(ctx).Debugf("Balance needs to be ajusted. Running balance %d, balance %d", runningBalance, wallet.Balance)
    
    adj := wallet.Balance - runningBalance
    t := transactions.New(adj, "Balance Adjustment", "Adjustment", "-")
    t.RunningBalance = wallet.Balance
    err = v.DoAppendTransaction(ctx, wallet, t)
    if err != nil {
      logger.L(ctx).Debugf("Failed to append adjustment for %s", wallet.Name)
      return nil, err
    }

    logger.L(ctx).Debugf("Adjusted")
    trs = append(trs, *t)
  }

  return trs, nil
}

func (v *FileSystem) GetTransactionFilePath(wallet *wallets.Wallet) string {
  return v.RootPath + v.Ps + wallet.Name + v.Ps + TRANSACTIONS_FILENAME
}

func (v *FileSystem) GetReceiptsDir(wallet *wallets.Wallet) string {
  return v.RootPath + v.Ps + wallet.Name + v.Ps + config.DIR_RECEIPTS
}

func (v *FileSystem) GetLastTransaction(ctx context.Context, wallet *wallets.Wallet) (*transactions.Transaction, error) {
  ts, err := v.GetTransactions(ctx, wallet)
  if err != nil {
    return nil, perror.New(perror.ERROR_READ_TRANSACTIONS, "Failed to get transactions: " + err.Error())
  }

  if len(ts) == 0 {
    return nil, nil
  }

  return &ts[len(ts) - 1], nil
}

func (v *FileSystem) InitTransactionsFile(ctx context.Context, wallet *wallets.Wallet) error {
  path := v.GetTransactionFilePath(wallet)

  _, err := os.Stat(path)
  if err != nil {
    if os.IsNotExist(err) {
      logger.L(ctx).Debugf("Transactions don't exist. Creating them")
      file, err := os.Create(path)
      if err != nil {
        logger.L(ctx).Errorf("Failed to create transactions file: %s", err.Error())
        return err
      }

      file.Close()
    } else {
      logger.L(ctx).Errorf("Failed to stat transactions file %s: %s", path, err)
      return err
    }
  }

  dir := v.GetReceiptsDir(wallet)
  _, err = os.Stat(dir)
  if err != nil {
    if os.IsNotExist(err) {
      logger.L(ctx).Debugf("Creating Receipts dir")
      err := os.Mkdir(dir, 0755)
      if err != nil {
        logger.L(ctx).Errorf("Failed to create receipts dir: %s", err.Error())
        return err
      }
    }
  }

 
  return nil
}

