package filesystem

import (
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
  "context"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
)


type FileSystem struct {
  RootPath string
  SkyRootPath string
  Ps string
  Ls string
}

func New() *FileSystem {
  return &FileSystem{}
}

type WConfig struct {
  Email string `toml:"email"`
  PasswordHash string `toml:"passwordhash"`
}

/* Interface Functions */

func (v *FileSystem) InitWallets(ctx context.Context) error {
  v.Ps = string(os.PathSeparator)
  if runtime.GOOS == "windows" {
    v.Ls = "\r\n"
  } else {
    v.Ls = "\n"
  }

  v.RootPath = config.ROOT_PATH + v.Ps + WALLETS_DIR
  err := v.MkDir(ctx, v.RootPath)
  if err != nil {
    return err
  }

  v.SkyRootPath = config.ROOT_PATH + v.Ps + SKYWALLETS_DIR
  err = v.MkDir(ctx, v.SkyRootPath)
  if err != nil {
    return err
  }

  wallets, err := v.GetWallets(ctx)
  if err != nil {
    return err
  }

  if len(wallets) == 0 {
    logger.L(ctx).Debugf("Creating default Wallet")
    if err := v.CreateWallet(ctx, config.DEFAULT_WALLET_NAME, "", ""); err != nil {
      logger.L(ctx).Errorf("Failed to create Default Wallet: " + err.Error())
      return perror.New(perror.ERROR_FAILED_TO_CREATE_WALLET, "Failed to create default wallet")
    }
  }

  return nil
}

func (v *FileSystem) CreateWallet(ctx context.Context, name string, email string, password string) error {
  logger.L(ctx).Debugf("Creating Wallet \"%s\" email %s", name, email)

  path := v.RootPath + v.Ps + name
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		return perror.New(perror.ERROR_WALLET_ALREADY_EXISTS, "Wallet already exists")
	}

  err = v.MkDir(ctx, path)
  if err != nil {
    logger.L(ctx).Errorf("Failed to create wallet folder: %s", err.Error())
    return err
  }

  passwordHash := utils.GetHash(password)
  err = v.SaveConfig(ctx, name, email, passwordHash)
  if err != nil {
    logger.L(ctx).Errorf("Failed to save wallet config: %s", err.Error())
    return err
  }


  folders := []string{
    DIR_BANK,
    DIR_FRACKED,
    DIR_COUNTERFEIT,
    DIR_LIMBO,
    DIR_IMPORT,
    DIR_TRASH,
    DIR_IMPORTED,
    DIR_SENT,
    DIR_SUSPECT,
  }

  for _, folder := range(folders) {
    fpath := path + v.Ps + folder
    if err := v.MkDir(ctx, fpath); err != nil {
      return err
    }
  }

  wallet := wallets.New(name, email, passwordHash)
  err = v.InitTransactionsFile(ctx, wallet) 
  if err != nil {
    return err
  }

  return nil
}

func (v *FileSystem) GetWallets(ctx context.Context) ([]wallets.Wallet, error) {
  logger.L(ctx).Debugf("Getting Wallets")

	files, err := ioutil.ReadDir(v.RootPath)
	if err != nil {
    logger.L(ctx).Errorf("Failed to read folder " + v.RootPath + ": " + err.Error())
    return nil, perror.New(perror.ERROR_FS_DRIVER_FAILED_TO_READ_FOLDER, "Failed to read folder " + v.RootPath + ": " + err.Error())
	}

  myWallets := make([]wallets.Wallet, 0)
	for _, f := range files {
		fname := f.Name()
    logger.L(ctx).Debugf("Reading Wallet \"%s\"", fname)

    config, err := v.GetConfig(ctx, fname)
    if err != nil {
      logger.L(ctx).Debugf("Failed to read and parse %s config file: %s", fname, err.Error())
      continue
    }

    wallet := wallets.New(fname, config.Email, config.PasswordHash)
    myWallets = append(myWallets, *wallet)
  }

  return myWallets, nil
}


func (v *FileSystem) GetWallet(ctx context.Context, name string) (*wallets.Wallet, error) {
  logger.L(ctx).Debugf("Getting Wallet %s", name)

  path := v.RootPath + v.Ps + name
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, perror.New(perror.ERROR_WALLET_NOT_FOUND, "Wallet not found")
	}

  config, err := v.GetConfig(ctx, name)
  if err != nil {
    logger.L(ctx).Debugf("Failed to read and parse %s config file: %s", name, err.Error())
    return nil, perror.New(perror.ERROR_WALLET_CONFIG_PARSE, "Failed to parse wallet config: " + err.Error())
  }

  wallet := wallets.New(name, config.Email, config.PasswordHash)

  return wallet, nil
}

func (v *FileSystem) GetWalletWithBalance(ctx context.Context, name string) (*wallets.Wallet, error) {
  logger.L(ctx).Debugf("Getting Wallet %s with balance", name)

  wallet, err := v.GetWallet(ctx, name)
  if err != nil {
    return nil, err
  }

  err = v.UpdateWalletBalance(ctx, wallet)
  if err != nil {
    return nil, err
  }

  return wallet, nil
}

func (v *FileSystem) GetWalletWithContents(ctx context.Context, name string) (*wallets.Wallet, error) {
  wallet, err := v.GetWallet(ctx, name)
  if err != nil {
    return nil, err
  }

  err = v.UpdateWalletContents(ctx, wallet)
  if err != nil {
    return nil, err
  }

  return wallet, nil
}

func (v *FileSystem) UpdateWalletBalance(ctx context.Context, wallet *wallets.Wallet) error {
  logger.L(ctx).Debugf("Updating Wallet %s Balance", wallet.Name)
  err := v.UpdateWalletContentsInDir(ctx, wallet, config.DIR_BANK)
  if err != nil {
    return err
  }

  err = v.UpdateWalletContentsInDir(ctx, wallet, config.DIR_FRACKED)
  if err != nil {
    return err
  }

  return nil
}

func (v *FileSystem) UpdateWalletContents(ctx context.Context, wallet *wallets.Wallet) error {
  logger.L(ctx).Debugf("Updating Wallet %s Contents", wallet.Name)

  err := v.UpdateWalletBalance(ctx, wallet)
  if err != nil {
    return err
  }

  transactions, err := v.GetTransactions(ctx, wallet)
  if err != nil {
    return err
  }

  wallet.Transactions = transactions

  return nil
}

func (v *FileSystem) ReadCoinsInLocation(ctx context.Context, wallet *wallets.Wallet, locationStatus int) ([]cloudcoin.CloudCoin, int, error) {
  dir := v.GetDirByLocationStatus(locationStatus)
  if dir == "" {
    logger.L(ctx).Errorf("Failed to find location status for %d", locationStatus)
    return nil, 0, perror.New(perror.ERROR_FILESYSTEM, "Invalid location")
  }

  path := v.RootPath + v.Ps + wallet.Name + v.Ps + dir

	files, err := ioutil.ReadDir(path)
	if err != nil {
    logger.L(ctx).Errorf("Failed to read folder " + path + ": " + err.Error())
    return nil, 0, perror.New(perror.ERROR_FS_DRIVER_FAILED_TO_READ_FOLDER, "Failed to read folder " + path + ": " + err.Error())
	}
 
  ignoredCoins := 0
  coins := make([]cloudcoin.CloudCoin, 0)
  for _, file := range(files) {
    name := file.Name()

    logger.L(ctx).Debugf("name %s", name)
    parts := strings.Split(name, ".")
    if (len(parts) < 5) {
      logger.L(ctx).Debugf("Invalid file in dir %s. Skipping it", name)
      ignoredCoins++
      continue
    }
    sn, err := strconv.Atoi(parts[3])
    if err != nil {
      logger.L(ctx).Debugf("Failed to convert Serial Number. Skipping coin")
      ignoredCoins++
      continue
    }

    cc := cloudcoin.NewFromData(uint32(sn))
    cc.SetLocationStatus(locationStatus)

    err = v.ReadCoin(ctx, wallet, cc)
    if err != nil {
      logger.L(ctx).Errorf("Failed to read coin %d: %s", cc.Sn, err)
      ignoredCoins++
      continue
    }

    coins = append(coins, *cc)
  }

  return coins, ignoredCoins, nil
}

func (v *FileSystem) UpdateWalletContentsForLocation(ctx context.Context, wallet *wallets.Wallet, locationStatus int) error {

  dir := v.GetDirByLocationStatus(locationStatus)
  if dir == "" {
    logger.L(ctx).Errorf("Failed to find location status for %d", locationStatus)
    return perror.New(perror.ERROR_FILESYSTEM, "Invalid location")
  }

  return v.UpdateWalletContentsInDir(ctx, wallet, dir)
}

func (v *FileSystem) UpdateWalletContentsInDir(ctx context.Context, wallet *wallets.Wallet, dir string) error {
  logger.L(ctx).Debugf("Updating Wallet %s in %s", wallet.Name, dir)


  location := v.GetLocationByDir(dir)

  path := v.RootPath + v.Ps + wallet.Name + v.Ps + dir
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return perror.New(perror.ERROR_DIRECTORY_NOT_EXIST, "Directory " + path + " not found")
  }

	files, err := ioutil.ReadDir(path)
	if err != nil {
    logger.L(ctx).Errorf("Failed to read folder " + path + ": " + err.Error())
    return perror.New(perror.ERROR_FS_DRIVER_FAILED_TO_READ_FOLDER, "Failed to read folder " + path + ": " + err.Error())
	}
 
  for _, file := range(files) {
    name := file.Name()

    //logger.L(ctx).Debugf("Reading filename %s", name)
    
    parts := strings.Split(name, ".")
    if (len(parts) < 5) {
      logger.L(ctx).Debugf("Invalid file in dir %s. Skipping it", name)
      continue
    }

    denomination, err := strconv.Atoi(parts[0])
    if err != nil {
      logger.L(ctx).Debugf("Failed to convert denomination. Skipping coin")
      continue
    }

    sn, err := strconv.Atoi(parts[3])
    if err != nil {
      logger.L(ctx).Debugf("Failed to convert Serial Number. Skipping coin")
      continue
    }

    if (cloudcoin.GetDenomination(uint32(sn)) != denomination) {
      logger.L(ctx).Debugf("Denomination (%d) doesn't match serial number (%d). Skipping coin", denomination, sn)
      continue
    }

    cc := cloudcoin.NewFromData(uint32(sn))
    cc.SetLocationStatus(location)

    wallet.Balance += denomination
    wallet.Denominations[denomination]++
    wallet.CoinsByDenomination[denomination] = append(wallet.CoinsByDenomination[denomination], cc)
    wallet.Contents = append(wallet.Contents, uint32(sn))
  }

  return nil
}

func (v *FileSystem) RenameWallet(ctx context.Context, wallet *wallets.Wallet, name string) error {
  logger.L(ctx).Debugf("Renaming Wallet %s to %s", wallet.Name, name)

  path := v.RootPath + v.Ps + wallet.Name
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return perror.New(perror.ERROR_WALLET_NOT_FOUND, "Wallet not found")
  }

  newPath := v.RootPath + v.Ps + name
  err = os.Rename(path, newPath)
  if err != nil {
    logger.L(ctx).Errorf("Failed to rename wallet: (new path %s) %s", newPath, err.Error())
    return perror.New(perror.ERROR_RENAME_WALLET, "Failed to rename wallet: " + err.Error())
  }

  return nil
}

func (v *FileSystem) DeleteWallet(ctx context.Context, name string) error {
  logger.L(ctx).Debugf("Deleting Wallet %s", name)

  path := v.RootPath + v.Ps + name
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return perror.New(perror.ERROR_WALLET_NOT_FOUND, "Wallet not found")
  }


  bankDir := path + v.Ps + config.DIR_BANK
	files, err := ioutil.ReadDir(bankDir)
	if err != nil {
    logger.L(ctx).Errorf("Failed to read folder " + bankDir + ": " + err.Error())
    return perror.New(perror.ERROR_FS_DRIVER_FAILED_TO_READ_FOLDER, "Failed to read folder " + bankDir + ": " + err.Error())
	}
  
  if (len(files) > 0) {
    logger.L(ctx).Errorf("Folder is not empty %s", bankDir)
    return perror.New(perror.ERROR_WALLET_NOT_EMPTY, "Wallet not empty")
  }

  frackedDir := path + v.Ps + config.DIR_FRACKED
	files, err = ioutil.ReadDir(frackedDir)
	if err != nil {
    logger.L(ctx).Errorf("Failed to read folder " + frackedDir + ": " + err.Error())
    return perror.New(perror.ERROR_FS_DRIVER_FAILED_TO_READ_FOLDER, "Failed to read folder " + frackedDir + ": " + err.Error())
	}

  if (len(files) > 0) {
    logger.L(ctx).Errorf("Folder is not empty %s", frackedDir)
    return perror.New(perror.ERROR_WALLET_NOT_EMPTY, "Wallet not empty")
  }

	err = os.RemoveAll(path)
  if err != nil {
    logger.L(ctx).Errorf("Failed to delete %s: %s", path, err.Error())
    return perror.New(perror.ERROR_DELETE_WALLET, "Failed to delete wallet: " + err.Error())
  }

  return nil
}

/* Internal Functions */
func (v *FileSystem) MkDir(ctx context.Context, path string) error {
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		return nil
	}

  logger.L(ctx).Debugf("Creating Folder %s", path)
	err = os.Mkdir(path, 0700)
	if err != nil {
    logger.L(ctx).Debugf("Created %s", path)
    return perror.New(perror.ERROR_FAILED_TO_CREATE_DIRECTORY, "Failed to create folder: " + err.Error())
	}

  return nil
}

func (v *FileSystem) SaveConfig(ctx context.Context, name, email, passwordHash string) error {
  configPath := v.RootPath + v.Ps + name + v.Ps + config.WALLET_CONFIG_NAME

  config := &WConfig{
    Email: email,
    PasswordHash: passwordHash,
  }

  f, err := os.Create(configPath)
  if err != nil {
    return err
  }

  defer f.Close()
  if err := toml.NewEncoder(f).Encode(&config); err != nil {
    return err
  }

  logger.L(ctx).Debugf("Wallet Config Saved")

  return nil

}

func (v *FileSystem) GetConfig(ctx context.Context, name string) (*WConfig, error) {
  var conf WConfig

  configPath := v.RootPath + v.Ps + name + v.Ps + config.WALLET_CONFIG_NAME

  data, err := ioutil.ReadFile(configPath)
  if err != nil {
    logger.L(ctx).Debugf("Failed to read wallet %s config file %s: %s", name, configPath, err.Error())
    return nil, err
  }

	if _, err := toml.Decode(string(data), &conf); err != nil {
    logger.L(ctx).Debugf("Failed to parse wallet %s config file: %s", name, err.Error())
		return nil, err
	}

  logger.L(ctx).Debugf("Wallet Config Read")

  return &conf, nil
}


