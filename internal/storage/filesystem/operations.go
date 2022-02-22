package filesystem

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
)

/* Interface Functions */
func (v *FileSystem) GetCoins(ctx context.Context, wallet *wallets.Wallet, location int) ([]cloudcoin.CloudCoin, error) {
  logger.L(ctx).Debugf("Getting coins in location %d", location)

  folder := v.GetDirByLocationStatus(location)
  if folder == "" {
    return nil, perror.New(perror.ERROR_INVALID_LOCATION, "Invalid Coin Location")
  }

  return nil, nil
}

func (v *FileSystem) EmptyLocation(ctx context.Context, wallet *wallets.Wallet, location int) error {
  folder := v.GetDirByLocationStatus(location)
  if folder == "" {
    return perror.New(perror.ERROR_INVALID_LOCATION, "Invalid Coin Location")
  }

  path := v.RootPath + v.Ps + wallet.Name + v.Ps + config.DIR_TRASH
  v.MkDir(ctx, path)

  sfolder := v.RootPath + v.Ps + wallet.Name + v.Ps + folder
  files, err := ioutil.ReadDir(sfolder)
  if err != nil {
    return perror.New(perror.ERROR_FS_DRIVER_FAILED_TO_READ_FOLDER, "Failed to read suspect folder: " +  err.Error())
  }

  for _, f := range files {
    fpath := sfolder + v.Ps + f.Name()

    newPath := path + v.Ps + strconv.Itoa(int(time.Now().Unix())) + "_" + f.Name() 
    err = os.Rename(fpath, newPath)
    if err != nil {
      logger.L(ctx).Errorf("Failed to move file to Trash: %s", err.Error())
      return err
    }

  }

  return nil
}


func (v *FileSystem) MoveCoins(ctx context.Context, srcWallet, dstWallet *wallets.Wallet, coins []cloudcoin.CloudCoin) (int, error) {
  logger.L(ctx).Debugf("Moving %d notes from %s to %s", len(coins), srcWallet.Name, dstWallet.Name)

  for _, cc := range(coins) {
    coinPath := v.CoinCurrentPath(ctx, &cc, srcWallet.Name)
    if (coinPath == "") {
      return 0, perror.New(perror.ERROR_COIN_INVALID_LOCATION, "Failed to get location of the Coin #" + strconv.Itoa(int(cc.Sn)))
    }

    folder := v.GetCurrentDir(ctx, &cc)
    if folder == "" {
      return 0, perror.New(perror.ERROR_INVALID_LOCATION, "Invalid Coin Location for Coin #" + strconv.Itoa(int(cc.Sn)))
    }

    if folder != config.DIR_BANK && folder != config.DIR_FRACKED {
      return 0, perror.New(perror.ERROR_INVALID_LOCATION, "Invalid Coin Location for Coin #" + strconv.Itoa(int(cc.Sn)) + ". Only Bank and Fracked required")
    }

    logger.L(ctx).Debugf("path %s folder %s", coinPath, folder)
    if v.CoinExists(ctx, &cc, dstWallet.Name, config.DIR_BANK) {
      return 0, perror.New(perror.ERROR_COIN_ALREADY_EXISTS, "Coin " + strconv.Itoa(int(cc.Sn)) + " already exists in wallet " + dstWallet.Name + "/Bank")
    }

    if v.CoinExists(ctx, &cc, dstWallet.Name, config.DIR_FRACKED) {
      return 0, perror.New(perror.ERROR_COIN_ALREADY_EXISTS, "Coin " + strconv.Itoa(int(cc.Sn)) + " already exists in wallet " + dstWallet.Name + "/Fracked")
    }
  }

  amount := 0
  for _, cc := range(coins) {
    folder := v.GetCurrentDir(ctx, &cc)
    err := v.MoveCoin(ctx, &cc, srcWallet.Name, dstWallet.Name, folder)
    if err != nil {
      logger.L(ctx).Errorf("Failed to move coin #%d: %s", cc.Sn, err.Error())
      continue
    }

    amount += cc.GetDenomination()
  }

  return amount, nil
}

func (v *FileSystem) ReadCoin(ctx context.Context, wallet *wallets.Wallet, cc *cloudcoin.CloudCoin) error {
  coinPath := v.CoinCurrentPath(ctx, cc, wallet.Name)
  if (coinPath == "") {
    return perror.New(perror.ERROR_COIN_INVALID_LOCATION, "Failed to get location of the Coin #" + strconv.Itoa(int(cc.Sn)))
  }

  logger.L(ctx).Debugf("Reading Coin %d: %s", cc.Sn, coinPath)


  data, err := os.ReadFile(coinPath) 
  if err != nil {
    logger.L(ctx).Errorf("Failed to read coin %s: %s", coinPath, err.Error())
    return err
  }

  if (len(data) < 32) {
    return perror.New(perror.ERROR_INVALID_HEADER_SIZE, "Invalid coin")
  }

  ccb, err := cloudcoin.NewFromBinary(ctx,data[0:32], data[32:])
  if err != nil {
    return err
  }

  if (ccb.Sn != cc.Sn) {
    return perror.New(perror.ERROR_SN_MISMATCH, "Coin serial number in the filename " + strconv.Itoa(int(cc.Sn))  + " differs from the one in the file " + strconv.Itoa(int(ccb.Sn)))
  }

  location := cc.GetLocationStatus()
  //pownString := cc.GetPownString()
  
  *cc = *ccb

  cc.SetLocationStatus(location)
  //cc.SetPownString(pownString)

  return nil
}

// Forcibly sets location (status)
func (v *FileSystem) SetLocation(ctx context.Context, wallet *wallets.Wallet, coins []cloudcoin.CloudCoin, location int) error {
  logger.L(ctx).Debugf("Setting location for total:%d coin(s) in the wallet \"%s\" to %d", len(coins), wallet.Name, location)

  for _, cc := range(coins) {
    folder := v.GetDirByLocationStatus(location)
    if folder == "" {
      return perror.New(perror.ERROR_INVALID_LOCATION, "Invalid Coin Location")
    }

    if v.CoinExists(ctx, &cc, wallet.Name, folder) {
      if (folder == config.DIR_SENT || folder == config.DIR_COUNTERFEIT) {
        logger.L(ctx).Debugf("removing dest coin, because it is duplicated")
        err := v.RemoveDupCoin(ctx, &cc, wallet.Name, folder)
        if err != nil {
          logger.L(ctx).Errorf("Failed to remove dup coin: %s", err.Error())
          return err
        }
      } else {
        return perror.New(perror.ERROR_COIN_ALREADY_EXISTS, "Coin " + strconv.Itoa(int(cc.Sn)) + " already exists in " + folder)
      }
    }
  }

  for idx, coin := range(coins) {
    targetDir := v.GetDirByLocationStatus(location)

    logger.L(ctx).Debugf("Moving coin %d targetDir %s, targetLocation %d", coin.Sn, targetDir, location)

    var err error
    if (location == config.COIN_LOCATION_STATUS_LIMBO) {
      err = v.SaveCoinWithPans(ctx, &coin, wallet.Name, targetDir)
    } else {
      err = v.SaveCoin(ctx, &coin, wallet.Name, targetDir)
    }
    if err != nil {
      logger.L(ctx).Debugf("Failed to save coin %d", coin.Sn)
      return perror.New(perror.ERROR_SAVE_COIN, "Failed to save coin " + strconv.Itoa(int(coin.Sn)) + ": " +  err.Error())
    }

    err = v.RemoveCoin(ctx, &coin, wallet.Name)
    if err != nil {
      logger.L(ctx).Debugf("Failed to remove coin %d", coin.Sn)
      return err
    }

    coins[idx].SetLocationStatus(location)
  }

  return nil
}

func (v *FileSystem) UpdateStatus(ctx context.Context, wallet *wallets.Wallet, coins []cloudcoin.CloudCoin) error {
  return v.UpdateStatusCommon(ctx, wallet, coins, true)
}

func (v *FileSystem) UpdateStatusForNewCoin(ctx context.Context, wallet *wallets.Wallet, coins []cloudcoin.CloudCoin) error {
  return v.UpdateStatusCommon(ctx, wallet, coins, false)
}

// Updates the status of the coin based on its grade status
func (v *FileSystem) UpdateStatusCommon(ctx context.Context, wallet *wallets.Wallet, coins []cloudcoin.CloudCoin, needRemove bool) error {
  logger.L(ctx).Debugf("Updating status for total:%d coin(s) in the wallet \"%s\" (need remove orig %v)", len(coins), wallet.Name, needRemove)

  for _, cc := range(coins) {
    targetDir := v.GetTargetDir(ctx, &cc)
    if targetDir == "" {
      return perror.New(perror.ERROR_COIN_INVALID_LOCATION, "Failed to get location of the Coin #" + strconv.Itoa(int(cc.Sn)))
    }

    if (targetDir == config.DIR_COUNTERFEIT) {
      if v.CoinExists(ctx, &cc, wallet.Name, targetDir) {
        logger.L(ctx).Debugf("Removing dest coin, because it is duplicated")
        err := v.RemoveDupCoin(ctx, &cc, wallet.Name, targetDir)
        if err != nil {
          logger.L(ctx).Errorf("Failed to remove dup coin: %s", err.Error())
          return err
        }
      }
    }

    if v.CoinExists(ctx, &cc, wallet.Name, targetDir) {
      return perror.New(perror.ERROR_COIN_ALREADY_EXISTS, "Coin " + strconv.Itoa(int(cc.Sn)) + " already exists in " + targetDir)
    }
  }

  for idx, coin := range(coins) {
    targetLocation := v.GetTargetLocation(ctx, &coin)
    targetDir := v.GetTargetDir(ctx, &coin)

    logger.L(ctx).Debugf("Moving coin %d targetDir %s, targetLocation %d", coin.Sn, targetDir, targetLocation)

    var err error
    if (targetLocation == config.COIN_LOCATION_STATUS_LIMBO) {
      err = v.SaveCoinWithPans(ctx, &coin, wallet.Name, targetDir)
    } else {
      err = v.SaveCoin(ctx, &coin, wallet.Name, targetDir)
    }
    if err != nil {
      logger.L(ctx).Debugf("Failed to save coin %d", coin.Sn)
      return perror.New(perror.ERROR_SAVE_COIN, "Failed to save coin " + strconv.Itoa(int(coin.Sn)) + ": " +  err.Error())
    }

    if (needRemove) {
      err = v.RemoveCoin(ctx, &coin, wallet.Name)
      if err != nil {
        logger.L(ctx).Debugf("Failed to remove coin %d", coin.Sn)
        return err
      }
    }

    coins[idx].SetLocationStatus(targetLocation)
  }

  return nil
}

func (v *FileSystem) UpdateCoins(ctx context.Context, wallet *wallets.Wallet, coins []cloudcoin.CloudCoin) error {
  logger.L(ctx).Debugf("Updating coins for total:%d coin(s) in the wallet \"%s\"", len(coins), wallet.Name)

  for _, cc := range(coins) {
    targetDir := v.GetTargetDir(ctx, &cc)
    if targetDir == "" {
      return perror.New(perror.ERROR_COIN_INVALID_LOCATION, "Failed to get location of the Coin #" + strconv.Itoa(int(cc.Sn)))
    }

    if !v.CoinExists(ctx, &cc, wallet.Name, targetDir) {
      return perror.New(perror.ERROR_COIN_INVALID_LOCATION, "Coin " + strconv.Itoa(int(cc.Sn)) + " doesn't exist in " + targetDir)
    }
  }

  for _, coin := range(coins) {
    targetLocation := v.GetTargetLocation(ctx, &coin)
    targetDir := v.GetTargetDir(ctx, &coin)

    logger.L(ctx).Debugf("Updating coin %d targetDir %s, targetLocation %d", coin.Sn, targetDir, targetLocation)
    err := v.RemoveCoin(ctx, &coin, wallet.Name)
    if err != nil {
      logger.L(ctx).Debugf("Failed to remove coin %d", coin.Sn)
      return err
    }

    if (targetLocation == config.COIN_LOCATION_STATUS_LIMBO) {
      err = v.SaveCoinWithPans(ctx, &coin, wallet.Name, targetDir)
    } else {
      err = v.SaveCoin(ctx, &coin, wallet.Name, targetDir)
    }
    if err != nil {
      logger.L(ctx).Debugf("Failed to save coin %d", coin.Sn)
      return perror.New(perror.ERROR_SAVE_COIN, "Failed to save coin " + strconv.Itoa(int(coin.Sn)) + ": " +  err.Error())
    }
  }

  return nil
}


func (v *FileSystem) PutInImported(ctx context.Context, wallet *wallets.Wallet, filePath string) error {
  _, file := filepath.Split(filePath)

  tag := strconv.Itoa(int(time.Now().Unix()))

  newPath := v.RootPath + v.Ps + wallet.Name + v.Ps + config.DIR_IMPORTED + v.Ps + tag + "." + file

  logger.L(ctx).Debugf("Moving %s to %s", filePath, newPath)

  _, err := os.Stat(newPath)
  if err == nil {
    logger.L(ctx).Errorf("File exists: %d", newPath)
    return perror.New(perror.ERROR_COIN_ALREADY_EXISTS, "File in Imported already exists")
  }

  if !os.IsNotExist(err) {
    logger.L(ctx).Errorf("Failed to stat %s: %s", newPath, err.Error())
    return err
  }

  err = os.Rename(filePath, newPath)
  if err != nil {
    logger.L(ctx).Errorf("Failed to move file to Imported: %s", err.Error())
    return err
  }

  return nil
}
/*
func (v *FileSystem) PutInImport(wallet *wallets.Wallet, coins []cloudcoin.CloudCoin) error {
  logger.L(ctx).Debugf("Putting %d coins in the wallet \"%s\" (Import)", len(coins), wallet.Name)

  for _, coin := range(coins) {
    if v.CoinExists(&coin, wallet.Name, config.DIR_IMPORT) {
      return perror.New(perror.ERROR_COIN_ALREADY_EXISTS, "Coin " + strconv.Itoa(int(coin.Sn)) + " already exists in " + config.DIR_IMPORT)
    }

    if v.CoinExists(&coin, wallet.Name, config.DIR_BANK) {
      return perror.New(perror.ERROR_COIN_ALREADY_EXISTS, "Coin " + strconv.Itoa(int(coin.Sn)) + " already exists in " + config.DIR_BANK)
    }

    if v.CoinExists(&coin, wallet.Name, config.DIR_FRACKED) {
      return perror.New(perror.ERROR_COIN_ALREADY_EXISTS, "Coin " + strconv.Itoa(int(coin.Sn)) + " already exists in " + config.DIR_FRACKED)
    }

    if v.CoinExists(&coin, wallet.Name, config.DIR_SUSPECT) {
      return perror.New(perror.ERROR_COIN_ALREADY_EXISTS, "Coin " + strconv.Itoa(int(coin.Sn)) + " already exists in " + config.DIR_SUSPECT)
    }
  }

  for idx, coin := range(coins) {
    err := v.SaveCoin(&coin, wallet.Name, config.DIR_IMPORT)
    if err != nil {
      logger.L(ctx).Debugf("Failed to save coin %d", coin.Sn)
      return perror.New(perror.ERROR_SAVE_COIN, "Failed to save coin " + strconv.Itoa(int(coin.Sn)) + ": " +  err.Error())
    }

    coins[idx].SetLocationStatus(config.COIN_LOCATION_STATUS_IMPORT)
  }

  return nil
}

*/

func (v *FileSystem) CoinExistsInTheWallet(ctx context.Context, wallet *wallets.Wallet, cc *cloudcoin.CloudCoin) bool {
  if v.CoinExists(ctx, cc, wallet.Name, config.DIR_SUSPECT) {
    logger.L(ctx).Warnf("Coin %d already exists in Suspect", cc.Sn)
    return true
  }
  if v.CoinExists(ctx, cc, wallet.Name, config.DIR_BANK) {
    logger.L(ctx).Warnf("Coin %d already exists in Bank", cc.Sn)
    return true
  }

  if v.CoinExists(ctx, cc, wallet.Name, config.DIR_FRACKED) {
    logger.L(ctx).Warnf("Coin %d already exists in Fracked", cc.Sn)
    return true
  }

  return false
}

func (v *FileSystem) CoinExistsInTheWalletAndAuthentic(ctx context.Context, wallet *wallets.Wallet, cc *cloudcoin.CloudCoin) bool {
  if v.CoinExists(ctx, cc, wallet.Name, config.DIR_BANK) {
    logger.L(ctx).Warnf("Coin %d already exists in Bank", cc.Sn)
    return true
  }

  if v.CoinExists(ctx, cc, wallet.Name, config.DIR_FRACKED) {
    logger.L(ctx).Warnf("Coin %d already exists in Fracked", cc.Sn)
    return true
  }

  return false
}

func (v *FileSystem) PutInSuspect(ctx context.Context, wallet *wallets.Wallet, coins []cloudcoin.CloudCoin) error {
  logger.L(ctx).Debugf("Moving %d coins in the wallet \"%s\" (Suspect)", len(coins), wallet.Name)


  for idx, coin := range(coins) {
    if v.CoinExists(ctx, &coin, wallet.Name, config.DIR_SUSPECT) {
      logger.L(ctx).Warnf("Coin %d already exists in Suspect. Writing it over", coin.Sn)
      srcPath := v.CoinPath(&coin, wallet.Name, config.DIR_SUSPECT)
      dstPath := v.CoinPath(&coin, wallet.Name, config.DIR_TRASH)
      err := os.Rename(srcPath, dstPath)
      if err != nil {
        logger.L(ctx).Errorf("Failed to move duplicated coin from the Suspect folder. This operation is fatal since there is no defined behavior for it: %s", err)
        return perror.New(perror.ERROR_SAVE_COIN, "Failed to move duplicated coin " + strconv.Itoa(int(coin.Sn)) + ". This operation is fatal")
      }
    }

    err := v.SaveCoinWithPans(ctx, &coin, wallet.Name, config.DIR_SUSPECT)
    if err != nil {
      logger.L(ctx).Debugf("Failed to save coin %d", coin.Sn)
      return perror.New(perror.ERROR_SAVE_COIN, "Failed to save coin " + strconv.Itoa(int(coin.Sn)) + ": " +  err.Error())
    }

    coins[idx].SetLocationStatus(config.COIN_LOCATION_STATUS_SUSPECT)
  }

  return nil
}

func (v *FileSystem) MoveCoin(ctx context.Context, cc *cloudcoin.CloudCoin, srcWalletName, dstWalletName string, dstDir string) error {
  srcPath := v.CoinPath(cc, srcWalletName, dstDir)
  dstPath := v.CoinPath(cc, dstWalletName, dstDir)

  logger.L(ctx).Debugf("Moving %s to %s", srcPath, dstPath)
  err := os.Rename(srcPath, dstPath)
  if err != nil {
    logger.L(ctx).Errorf("Failed to rename")
    return err
  }

  return nil
}

func (v *FileSystem) SaveCoin(ctx context.Context, cc *cloudcoin.CloudCoin, walletName string, dstDir string) error {
  path := v.CoinPath(cc, walletName, dstDir)

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

func (v *FileSystem) SaveCoinWithPans(ctx context.Context, cc *cloudcoin.CloudCoin, walletName string, dstDir string) error {
  path := v.CoinPath(cc, walletName, dstDir)

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

  data, err := cc.GetDataWithPans()
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

func (v *FileSystem) RemoveCoin(ctx context.Context, cc *cloudcoin.CloudCoin, walletName string) error {
  path := v.CoinCurrentPath(ctx, cc, walletName)
  if path == "" {
    logger.L(ctx).Errorf("Failed to get location of the coin")
    return perror.New(perror.ERROR_COIN_INVALID_LOCATION, "Failed to get location of the Coin")
  }

  logger.L(ctx).Debugf("Removing coin %s", path)
  err := os.Remove(path)
  if err != nil {
    return perror.New(perror.ERROR_COIN_INVALID_LOCATION, "Failed to remove coin " + strconv.Itoa(int(cc.Sn)) + ": " + err.Error())
  }

  return nil
}

func (v *FileSystem) CoinCurrentPath(ctx context.Context, cc *cloudcoin.CloudCoin, walletName string) string {
  status := cc.GetLocationStatus()

  dir := v.GetDirByLocationStatus(status)
  logger.L(ctx).Debugf("Got dir %s by status %d", dir, status)
  if dir == "" {
    return ""
  }

  return v.CoinPath(cc, walletName, dir)
}

func (v *FileSystem) GetDirByLocationStatus(status int) string {
  var dir string

  switch (status) {
  case config.COIN_LOCATION_STATUS_UNKNOWN:
    return ""
  case config.COIN_LOCATION_STATUS_SUSPECT:
    dir = config.DIR_SUSPECT
  case config.COIN_LOCATION_STATUS_BANK:
    dir = config.DIR_BANK
  case config.COIN_LOCATION_STATUS_FRACKED:
    dir = config.DIR_FRACKED
  case config.COIN_LOCATION_STATUS_IMPORT:
    dir = config.DIR_IMPORT
  case config.COIN_LOCATION_STATUS_COUNTERFEIT:
    dir = config.DIR_COUNTERFEIT
  case config.COIN_LOCATION_STATUS_SENT:
    dir = config.DIR_SENT
  case config.COIN_LOCATION_STATUS_LIMBO:
    dir = config.DIR_LIMBO
  default:
    return ""
  }

  return dir
}

func (v *FileSystem) GetTargetDir(ctx context.Context, cc *cloudcoin.CloudCoin) string {
  var location int

  location = v.GetTargetLocation(ctx, cc)
  strLocation := v.GetDirByLocationStatus(location)

  return strLocation
}

func (v *FileSystem) GetCurrentDir(ctx context.Context, cc *cloudcoin.CloudCoin) string {
  var location int

  location = cc.GetLocationStatus()
  strLocation := v.GetDirByLocationStatus(location)

  return strLocation
}




func (v *FileSystem) GetTargetLocation(ctx context.Context, cc *cloudcoin.CloudCoin) int {
  var location int

  status := cc.GetGradeStatus()

  logger.L(ctx).Debugf("Getting location for #%d, %s", cc.Sn, cc.GetGradeStatusString())
  switch (status) {
  case config.COIN_STATUS_AUTHENTIC:
    location = config.COIN_LOCATION_STATUS_BANK
  case config.COIN_STATUS_FRACKED:
    location = config.COIN_LOCATION_STATUS_FRACKED
  case config.COIN_STATUS_COUNTERFEIT:
    location = config.COIN_LOCATION_STATUS_COUNTERFEIT
  case config.COIN_STATUS_LIMBO:
    location = config.COIN_LOCATION_STATUS_LIMBO
  case config.COIN_STATUS_UNKNOWN:
    location = config.COIN_LOCATION_STATUS_TRASH
  default:
    location = config.COIN_LOCATION_STATUS_TRASH
  }

  return location
}

func (v *FileSystem) GetLocationByDir(dir string) int {
  switch (dir) {
  case config.DIR_BANK:
    return config.COIN_LOCATION_STATUS_BANK
  case config.DIR_FRACKED:
    return config.COIN_LOCATION_STATUS_FRACKED
  case config.DIR_COUNTERFEIT:
    return config.COIN_LOCATION_STATUS_COUNTERFEIT
  case config.DIR_LIMBO:
    return config.COIN_LOCATION_STATUS_LIMBO
  case config.DIR_SENT:
    return config.COIN_LOCATION_STATUS_SENT
  default:
    return config.COIN_LOCATION_STATUS_UNKNOWN

  }

}

func (v *FileSystem) CoinPath(cc *cloudcoin.CloudCoin, walletName string, dir string) string {
  return v.RootPath + v.Ps + walletName + v.Ps + dir + v.Ps + cc.GetName()
}

func (v *FileSystem) CoinExists(ctx context.Context, cc *cloudcoin.CloudCoin, walletName string, dir string) bool {
  path := v.CoinPath(cc, walletName, dir)

  logger.L(ctx).Debugf("Checking path %s", path)

  _, err := os.Stat(path)
  if err == nil {
    return true
  }
  
  if os.IsNotExist(err) {
    return false
  }
  
  // Rarely happens. Regard as 'not exists'
  logger.L(ctx).Debugf("Weird error %s", err.Error())

  return false
}

func (v *FileSystem) RemoveDupCoin(ctx context.Context, cc *cloudcoin.CloudCoin, walletName string, folder string) error {
  path := v.CoinPath(cc, walletName, folder)
  logger.L(ctx).Debugf("Removing duplicated path %s", path)

  err := os.Remove(path)
  if err != nil {
    return err
  }

  return nil

}
