package filesystem

import (
	"bufio"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/unpacker"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
)

//if strings.HasSuffix(fname, config.CC_FILE_BINARY_EXTENSION) {
//      fname = fname[:len(fname) - len(config.CC_FILE_BINARY_EXTENSION)]
//    }

/* Interface Functions */

func (v *FileSystem) CreateSkyWallet(ctx context.Context, cc *cloudcoin.CloudCoin) error {
  name := cc.GetSkyName()
  if (name == "") {
    return perror.New(perror.ERROR_NO_SKYWALLET_NAME, "Skywallet name is not defined")
  }

  logger.L(ctx).Debugf("Creating Sky Wallet \"%s\" SN:%d ", name, cc.Sn)

  path := v.SkyRootPath + v.Ps + name
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		return perror.New(perror.ERROR_WALLET_ALREADY_EXISTS, "SkyWallet already exists")
	}

  err = v.MkDir(ctx, path)
  if err != nil {
    logger.L(ctx).Errorf("Failed to create wallet folder: %s", err.Error())
    return err
  }

  err = v.SaveSkyCoin(ctx, cc)
  if err != nil {
    logger.L(ctx).Errorf("Failed to save sky coin: %s", err.Error())
    return err
  }

  return nil
}

func (v *FileSystem) CreateSkyWalletWithData(ctx context.Context, cc *cloudcoin.CloudCoin, data []byte) error {
  name := cc.GetSkyName()
  if (name == "") {
    return perror.New(perror.ERROR_NO_SKYWALLET_NAME, "Skywallet name is not defined")
  }

  logger.L(ctx).Debugf("Creating Sky Wallet \"%s\" SN:%d ", name, cc.Sn)

  path := v.SkyRootPath + v.Ps + name
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		return perror.New(perror.ERROR_WALLET_ALREADY_EXISTS, "SkyWallet already exists")
	}

  err = v.MkDir(ctx, path)
  if err != nil {
    logger.L(ctx).Errorf("Failed to create wallet folder: %s", err.Error())
    return err
  }

  err = v.SaveSkyCoinWithData(ctx, cc, data)
  if err != nil {
    logger.L(ctx).Errorf("Failed to save sky coin: %s", err.Error())
    return err
  }

  return nil
}

func (v *FileSystem) GetFirstSkyWallet(ctx context.Context) (*wallets.SkyWallet, error) {
  logger.L(ctx).Debugf("Getting First SkyWallet")

  path := v.SkyRootPath 

  files, err := ioutil.ReadDir(path)
  if err != nil {
    logger.L(ctx).Errorf("Failed to read SkyWallet dir %s: %s", path, err.Error())
    return nil, perror.New(perror.ERROR_FS_DRIVER_FAILED_TO_READ_FOLDER, "Failed to read folder " + path + ": " + err.Error())
  }

  for _, f := range(files) {
    fname := f.Name()

    if fname == SENDER_HISTORY_FILENAME {
      continue
    }

    logger.L(ctx).Debugf("Reading file %s", fname)
    skywallet, err := v.GetSkyWallet(ctx, fname)
    if err != nil {
      logger.L(ctx).Warnf("Failed to read skywallet %s. Skipping it: %s", fname, err.Error())
      continue
    }

    return skywallet, nil
  }

  return nil, nil
}

func (v *FileSystem) GetSkyWallets(ctx context.Context) ([]wallets.SkyWallet, error) {
  logger.L(ctx).Debugf("Getting Sky Wallets")

  path := v.SkyRootPath 

  files, err := ioutil.ReadDir(path)
  if err != nil {
    logger.L(ctx).Errorf("Failed to read SkyWallet dir %s: %s", path, err.Error())
    return nil, perror.New(perror.ERROR_FS_DRIVER_FAILED_TO_READ_FOLDER, "Failed to read folder " + path + ": " + err.Error())
  }

  mySkyWallets := make([]wallets.SkyWallet, 0)
  for _, f := range(files) {
    fname := f.Name()

    if fname == SENDER_HISTORY_FILENAME {
      continue
    }

    logger.L(ctx).Debugf("Reading file %s", fname)
    skywallet, err := v.GetSkyWallet(ctx, fname)
    if err != nil {
      logger.L(ctx).Warnf("Failed to read skywallet %s. Skipping it: %s", fname, err.Error())
      continue
    }

    mySkyWallets = append(mySkyWallets, *skywallet)
  }


  return mySkyWallets, nil
}

func (v *FileSystem) GetIDCoinPath(name string) string {
  return v.SkyRootPath + v.Ps + name + v.Ps + name + config.CC_FILE_BINARY_EXTENSION
}

func (v *FileSystem) GetIDCoinPathPng(name string) string {
  return v.SkyRootPath + v.Ps + name + v.Ps + name + config.CC_FILE_PNG_EXTENSION
}

func (v *FileSystem) GetSkyReceiptsDir(name string) string {
  return v.SkyRootPath + v.Ps + name + v.Ps + config.DIR_RECEIPTS
}

func (v *FileSystem) GetSkyReceiptFile(name string, guid string) string {
  return v.GetSkyReceiptsDir(name) + v.Ps + guid + ".txt"
}

func (v *FileSystem) GetSenderHistoryFilename() string {
  return v.SkyRootPath + v.Ps + SENDER_HISTORY_FILENAME
}

func (v *FileSystem) GetSkyWallet(ctx context.Context, name string) (*wallets.SkyWallet, error) {
  logger.L(ctx).Debugf("Getting SkyWallet %s", name)

  isPNG := false
  path := v.GetIDCoinPath(name)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
    path = v.GetIDCoinPathPng(name)
  	_,  err = os.Stat(path)
    if os.IsNotExist(err) {
      logger.L(ctx).Errorf("ID Coin %s does not exist", path)
	  	return nil, perror.New(perror.ERROR_WALLET_NOT_FOUND, "SkyWallet ID coin not found")
    }

    isPNG = true
    logger.L(ctx).Debug("The ID coin is PNG")
  }

  bytes, err := ioutil.ReadFile(path)
  if err != nil {
    logger.L(ctx).Errorf("Failed to read coin %s: %s", path, err.Error())
		return nil, perror.New(perror.ERROR_READING_COIN, "Failed to read ID coin")
  }

  var cc *cloudcoin.CloudCoin
  if !isPNG {
    cc, err = cloudcoin.NewFromBinarySingle(ctx, bytes)
    if err != nil {
      logger.L(ctx).Errorf("Failed to parse coin %s: %s", path, err.Error())
		  return nil, perror.New(perror.ERROR_READING_COIN, "Failed to parse ID coin")
    }
  } else {
    u := unpacker.New()
    ccs, err := u.Unpack(ctx, bytes)
    if err != nil {
      logger.L(ctx).Errorf("Failed to unpack ID coin %s: %s", path, err.Error())
	  	return nil, perror.New(perror.ERROR_READING_COIN, "Failed to ubpack ID PNG coin")
    }
    if len(ccs) != 1 {
      return nil, perror.New(perror.ERROR_READING_COIN, "Number of coins in the PNG must be one. We have: " + strconv.Itoa(len(ccs)))
    }

    cc = &ccs[0]
  }

  cc.SetSkyName(name)

  wallet := wallets.NewSkyWallet(name)
  wallet.SetIDCoin(cc)
	if isPNG {
    wallet.PNG = b64.StdEncoding.EncodeToString([]byte(bytes))
  }

  return wallet, nil
}

func (v *FileSystem) DeleteSkyWallet(ctx context.Context, name string) error {
  logger.L(ctx).Debugf("Deleting Sky Wallet %s", name)

  path := v.SkyRootPath + v.Ps + name
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return perror.New(perror.ERROR_WALLET_NOT_FOUND, "SkyWallet not found")
  }

	err = os.RemoveAll(path)
  if err != nil {
    logger.L(ctx).Errorf("Failed to delete %s: %s", path, err.Error())
    return perror.New(perror.ERROR_DELETE_WALLET, "Failed to delete SkyWallet: " + err.Error())
  }

  return nil
}

func (v *FileSystem) SaveSkyCoin(ctx context.Context, cc *cloudcoin.CloudCoin) error {
  path := v.GetIDCoinPath(cc.GetSkyName())

  logger.L(ctx).Debugf("Saving coin %d: %s", cc.Sn, path)

  _, err := os.Stat(path)
  if err == nil {
    logger.L(ctx).Errorf("Coin exists: %d", cc.Sn)
    return perror.New(perror.ERROR_COIN_ALREADY_EXISTS, "Coin " + strconv.Itoa(int(cc.Sn)) + " already exists")
  }

  if !os.IsNotExist(err) {
    logger.L(ctx).Errorf("Failed to stat %s: %s", path, err.Error())
    return err
  }

  data, err := cc.GetData()
  if err != nil {
    logger.L(ctx).Errorf("Failed to get coin data sn %d: %s", cc.Sn, err.Error())
    return err
  }

  err = os.WriteFile(path, data, 0644)
  if err != nil {
    logger.L(ctx).Errorf("Failed to save file %s: %s", path, err.Error())
    return err
  }

  return nil
}



func (v *FileSystem) SaveSkyCoinWithData(ctx context.Context, cc *cloudcoin.CloudCoin, data []byte) error {
  path := v.GetIDCoinPathPng(cc.GetSkyName())

  logger.L(ctx).Debugf("Saving coin with data %d: %s", cc.Sn, path)
  _, err := os.Stat(path)
  if err == nil {
    logger.L(ctx).Errorf("Coin exists: %d", cc.Sn)
    return perror.New(perror.ERROR_COIN_ALREADY_EXISTS, "Coin " + strconv.Itoa(int(cc.Sn)) + " already exists")
  }

  if !os.IsNotExist(err) {
    logger.L(ctx).Errorf("Failed to stat %s: %s", path, err.Error())
    return err
  }

  err = os.WriteFile(path, data, 0644)
  if err != nil {
    logger.L(ctx).Errorf("Failed to save file %s: %s", path, err.Error())
    return err
  }

  return nil
}









































func (v *FileSystem) AppendSenderHistory(ctx context.Context, record string) error {
  fname := v.GetSenderHistoryFilename()

  file, err := os.OpenFile(fname, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
  if err != nil {
    logger.L(ctx).Errorf("Failed to open sender history file %s: %s", fname, err.Error())
    return err
  }

  defer file.Close()

  _, err = file.WriteString(record + v.Ls)
  if err != nil {
    logger.L(ctx).Errorf("Failed to write to file %s: %s", file, err.Error())
    return err
  }

  logger.L(ctx).Debugf("Record has been written")

  return nil
}

func (v *FileSystem) GetSenderHistory(ctx context.Context, pattern string) ([]string, error) {
  fname := v.GetSenderHistoryFilename()

  records := make([]string, 0)

  file, err := os.OpenFile(fname, os.O_RDONLY, 0644)
  if err != nil {
    if errors.Is(err, os.ErrNotExist) {
      logger.L(ctx).Debugf("History file %s Filed hasn't been created yet. No records", fname)
      return records, nil
    }
    
    logger.L(ctx).Errorf("Failed to open sender history file %s: %s", fname, err.Error())
    return nil, err
  }

  defer file.Close()

  scanner := bufio.NewScanner(file)
  r, err := regexp.Compile(pattern)
  if err != nil {
    logger.L(ctx).Errorf("Invalid regexp")
    return nil, perror.New(perror.ERROR_INVALID_REGEX, "Invalid regexp")
  }

  for scanner.Scan() {
    if r.MatchString(scanner.Text()) {
      records = append(records, scanner.Text())
    }
  }

  err = scanner.Err()
  if err != nil {
    logger.L(ctx).Errorf("Failed to read file %s: %s", fname, err.Error())
    return nil, perror.New(perror.ERROR_FILESYSTEM, "Failed to read file " + fname + ": " + err.Error())
  }

  logger.L(ctx).Debugf("Records have been returned")

  return records, nil
}


func (v *FileSystem) AppendSkyTransactionDetails(ctx context.Context, w *wallets.SkyWallet, sd *wallets.StatementTransaction) error {
  dir := v.GetSkyReceiptsDir(w.Name)
	_, err := os.Stat(dir)
  if err != nil {
    if os.IsNotExist(err) {
      err = v.MkDir(ctx, dir)
      if err != nil {
        logger.L(ctx).Errorf("Failed to create receipts folder: %s", err.Error())
        return err
      }
    } else {
      return err
    }
  }

  path := v.GetSkyReceiptFile(w.Name, sd.ID)

  data, err := json.Marshal(sd)
  if err != nil {
    logger.L(ctx).Errorf("Failed to write json: %s", err.Error())
    return err
  }

  err = os.WriteFile(path, data, 0644)
  if err != nil {
    logger.L(ctx).Errorf("Failed to save file %s: %s", path, err.Error())
    return err
  }

  return nil
}

func (v *FileSystem) GetSkyTransactionDetails(ctx context.Context, w *wallets.SkyWallet, guid string) (*wallets.StatementTransaction, error) {
  path := v.GetSkyReceiptFile(w.Name, guid)

  data, err := ioutil.ReadFile(path)
  if err != nil {
    logger.L(ctx).Errorf("Failed to read trnsaction %s for %s: %s", guid, w.Name, err.Error())
    return nil, err
  }

  var st wallets.StatementTransaction

  err = json.Unmarshal(data, &st)
  if err != nil {
    logger.L(ctx).Errorf("Failed to read json: %s", err.Error())
    return nil, err
  }

  return &st, nil
}
