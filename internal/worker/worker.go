package worker

import (
	"context"
	"os"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
)



type Worker struct {
  progressChannel chan interface{}
  BreakInBankCallback func(context.Context, *wallets.SkyWallet) error
}


func New(progressChannel chan interface{}) (*Worker) {
  return &Worker{
    progressChannel: progressChannel,
    BreakInBankCallback: nil,
  }
}

func (v *Worker) SetBreakInBankCallback(f func(context.Context, *wallets.SkyWallet) error) {
  v.BreakInBankCallback = f
}

func (v *Worker) BreakInBank(ctx context.Context, wallet *wallets.SkyWallet) error {
  cc, err := v.PickCoinToBreak(wallet)
  if err != nil {
    logger.L(ctx).Errorf("Failed to find a coin to break")
    return err
  }

  logger.L(ctx).Debugf("Picked CC %d denomination %d for breaking", cc.Sn, cc.GetDenomination())

  sns, err := v.GetSNsFromChangeServer(ctx, cc.GetDenomination())
  if err != nil {
    logger.L(ctx).Errorf("Failed to get sns: %s", err.Error())
    return err
  }

  for _, sn := range(sns) {
    logger.L(ctx).Debugf("picked %d", sn)
  }

  breaker := raida.NewBreakInBank(nil)
  _, err = breaker.BreakInBank(ctx, wallet.IDCoin, cc.Sn, sns)
  if err != nil {
    logger.L(ctx).Errorf("Failed to make actual break for coins: %s", err.Error())
    return err
  }

  // Update Sky Wallet
  if v.BreakInBankCallback != nil {
    err := v.BreakInBankCallback(ctx, wallet)
    if err != nil {
      logger.L(ctx).Errorf("Failed to update skywallet after making change: %s", err.Error())
      return err
    }
  }

  return nil
}

func (v *Worker) Break(ctx context.Context, wallet *wallets.Wallet) error {
  logger.L(ctx).Debugf("Breaking coin in %s", wallet.Name)

  cc, err := v.PickCoinToBreak(wallet)
  if err != nil {
    logger.L(ctx).Errorf("Failed to find a coin to break")
    return err
  }

  logger.L(ctx).Debugf("Picked CC %d denomination %d for breaking", cc.Sn, cc.GetDenomination())

  sns, err := v.GetSNsFromChangeServer(ctx, cc.GetDenomination())
  if err != nil {
    logger.L(ctx).Errorf("Failed to get sns: %s", err.Error())
    return err
  }

  /*
  // Generating PGs
  pgs := make([]string, 0)
  for i := 0; i < config.TOTAL_RAIDA_NUMBER; i++ {
    pg, _ := utils.GeneratePG()
    pgs = append(pgs, pg)
  }
  */

  // Read Coin's ANs from the disk
  err = storage.GetDriver().ReadCoin(ctx, wallet, cc)
  if err != nil {
    logger.L(ctx).Debugf("Failed to read coin %d: %s", cc.Sn, err.Error())
    return err
  }

  // Create Coins
  coins := make([]cloudcoin.CloudCoin, 0)
  for _, sn := range(sns) {
    cc := cloudcoin.NewFromData(sn)
    cc.GenerateMyPans()

    coins = append(coins, *cc)
  }

  for _, cc := range(coins) {
    logger.L(ctx).Debugf("pre cc %d (%ds) %s %s", cc.Sn, cc.GetDenomination(), cc.GetGradeStatusString(), cc.GetPownString())
  }

  // No Progress Channel
  
  breaker := raida.NewBreak(nil)
  _, err = breaker.BreakCoins(ctx, cc, coins)
  if err != nil {
    logger.L(ctx).Debugf("Failed to make actual break for coins: %s", err.Error())
    return err
  }
  // Construct CloudCoins from PGs 

  for idx, _ := range(coins) {
    coins[idx].Grade()
    coins[idx].SetAnsToPansIfPassed()

    logger.L(ctx).Debugf("cc %d (%ds) %s %s", coins[idx].Sn, coins[idx].GetDenomination(), coins[idx].GetGradeStatusString(), coins[idx].GetPownString())
    ok, _, _ := coins[idx].IsAuthentic()
    if !ok {
      logger.L(ctx).Warnf("Coin %d is not authentic: %s", coins[idx].Sn, coins[idx].GetPownString)
    }
  }

  for idx, _ := range(coins) {
    logger.L(ctx).Debugf("xcc %d (%ds) %s %s %d", coins[idx].Sn, coins[idx].GetDenomination(), coins[idx].GetGradeStatusString(), coins[idx].GetPownString(), coins[idx].GetGradeStatus())
  }
  logger.L(ctx).Debugf("Updating coins status")
  err = storage.GetDriver().UpdateStatusForNewCoin(ctx, wallet, coins)
  if err != nil {
    logger.L(ctx).Errorf("Failed to update status of the changed coins: %s", err.Error())
    return err
  }

  logger.L(ctx).Debugf("Moving orig coin %d to sent", cc.Sn)
  err = storage.GetDriver().SetLocation(ctx, wallet, []cloudcoin.CloudCoin{*cc}, config.COIN_LOCATION_STATUS_SENT)
  if err != nil {
    logger.L(ctx).Errorf("Failed to move coin to Sent: %s", err.Error())
    return err
  }

  // Refresh Wallet
  err = storage.GetDriver().UpdateWalletBalance(ctx, wallet)
  if err != nil {
    logger.L(ctx).Errorf("Failed to Update Wallet balance %s: %s", wallet.Name, err.Error())
    return err
  }

  logger.L(ctx).Debugf("Break completed")

  return nil
}


func (v *Worker) GetSNsFromChangeServer(ctx context.Context, denomination int) ([]uint32, error) {
  // No Progress Channel
  showChange := raida.NewShowChange(nil)
  so, err := showChange.ShowChange(ctx, denomination)
  if err != nil {
    logger.L(ctx).Errorf("Failed to make change: %s", err.Error())
    return nil, err
  }

  if (len(so.SerialNumbers) == 0) {
    logger.L(ctx).Errorf("No coins on the Change Server for this denomination")
    return nil, perror.New(perror.ERROR_EMPTY_CHANGE_SERVER, "No coins on the Change server")
  }

  logger.L(ctx).Debugf("Picked %d notes from Change Server", len(so.SerialNumbers))


  csns1 := make([]uint32, 0)
  csns5 := make([]uint32, 0)
  csns25 := make([]uint32, 0)
  csns100 := make([]uint32, 0)

  for _, sn := range(so.SerialNumbers) {
    logger.L(ctx).Debugf("sn %d:%d", sn, cloudcoin.GetDenomination(sn))
    switch cloudcoin.GetDenomination(sn) {
    case 1:
      csns1 = append(csns1, sn)
    case 5:
      csns5 = append(csns5, sn)
    case 25:
      csns25 = append(csns25, sn)
    case 100:
      csns100 = append(csns100, sn)
    default:
      logger.L(ctx).Debugf("Igonoring coin %d, d:%d", sn, cloudcoin.GetDenomination(sn))
      continue
    }
  }


  // Pick proper SNs for denomination
  var sns []uint32
  switch denomination {
	case 5:
		sns = cloudcoin.CoinsGetA(csns1, 5)
	case 25:
		sns = cloudcoin.CoinsGet25B(csns5, csns1)
	case 100:
		sns = cloudcoin.CoinsGet100E(csns25, csns5, csns1)
	case 250:
		sns = cloudcoin.CoinsGet250F(csns100, csns25, csns5, csns1)
  default:
    return nil, perror.New(perror.ERROR_INVALID_DENOMINATION, "Can't break coin of denomination: " + strconv.Itoa(denomination))
	}

  if sns == nil {
    logger.L(ctx).Errorf("Failed to collect coins from ShowChange response for denomination %d", denomination)
    return nil, perror.New(perror.ERROR_COLLECTING_COINS_FROM_CHANGE, "Failed to collect change from ShowChange")
  }

  total := 0
  for _, sn := range(sns) {
    logger.L(ctx).Debugf("Picked coin %d (%d) for breaking", sn, cloudcoin.GetDenomination(sn))
    total += cloudcoin.GetDenomination(sn)
  }

  if total != denomination {
    logger.L(ctx).Errorf("Collected wrong total %d. Needed %d", total, denomination)
    return nil, perror.New(perror.ERROR_COLLECTING_COINS_FROM_CHANGE, "Failed to collect needed total. Picked only " + strconv.Itoa(total))
  }

  return sns, nil
}

func (v *Worker) PickCoinToBreak(wallet wallets.WalletInterface) (*cloudcoin.CloudCoin, error) {
 
  dens := wallet.GetCoinsByDenominations()

  if (len(dens[250]) > 0) {
    return dens[250][0], nil
  }

  if (len(dens[100]) > 0) {
    return dens[100][0], nil
  }

  if (len(dens[25]) > 0) {
    return dens[25][0], nil
  }

  if (len(dens[5]) > 0) {
    return dens[5][0], nil
  }

  return nil, perror.New(perror.ERROR_FAILED_TO_FIND_COIN_TO_BREAK, "No coin to break")
}

func (v *Worker) GetExpCoins(ctx context.Context, amount int, totals map[int]int, isLoose bool) (map[int]int, error) {
  savedAmount := amount

  for k, v := range(totals) {
    if k == 1 {
      logger.L(ctx).Debugf("Notes %d", v)
      break
    }
  }

  var exp1, exp5, exp25, exp100, exp250 int
  for i := 0; i < 2; i++ {
    exp1 = 0
    exp5 = 0
    exp25 = 0
    exp100 = 0

    if i == 0 && amount >= 250 && totals[250] > 0 {
      if (amount / 250) < totals[250] {
        exp250 = amount / 250
      } else {
        exp250 = totals[250]
      }

      amount -= (exp250 * 250)
    }

    if amount >= 100 && totals[100] > 0 {
      if (amount / 100) < totals[100] {
        exp100 = amount / 100
      } else {
        exp100 = totals[100]
      }

      amount -= (exp100 * 100)
    }

    if amount >= 25 && totals[25] > 0 {
      if (amount / 25) < totals[25] {
        exp25 = amount / 25
      } else {
        exp25 = totals[25]
      }

      amount -= (exp25 * 25)
    }

    if amount >= 5 && totals[5] > 0 {
      if (amount / 5) < totals[5] {
        exp5 = amount / 5
      } else {
        exp5 = totals[5]
      }

      amount -= (exp5 * 5)
    }

    if amount >= 1 && totals[1] > 0 {
      if (amount / 1) < totals[1] {
        exp1 = amount 
      } else {
        exp1 = totals[1]
      }

      amount -= (exp1)
    }

    logger.L(ctx).Debugf("%d/%d/%d/%d/%d amount %d", exp1, exp5, exp25, exp100, exp250, amount)
    if amount == 0 {
      break
    }

    if i == 1 || exp250 == 0 {
      if isLoose {
        break
      }

      logger.L(ctx).Debugf("Can't collect needed amount. Rest %d", amount)
      return nil, perror.New(perror.ERROR_FAILED_TO_PICK_COINS, "Failed to collect coins")
    }

    exp250--
    amount = savedAmount - exp250 * 250
  }

  rv := make(map[int]int, 0)
  rv[1] = exp1
  rv[5] = exp5
  rv[25] = exp25
  rv[100] = exp100
  rv[250] = exp250

  return rv, nil
}

func (v *Worker) GetCoinsToDealWithNoRead(ctx context.Context, wallet *wallets.Wallet, amount int) ([]cloudcoin.CloudCoin, error) {
  return v.GetCoinsToDealWithCommon(ctx, wallet, amount, false)
}

func (v *Worker) GetCoinsToDealWith(ctx context.Context, wallet *wallets.Wallet, amount int) ([]cloudcoin.CloudCoin, error) {
  return v.GetCoinsToDealWithCommon(ctx, wallet, amount, true)
}

func (v *Worker) GetCoinsToDealWithCommon(ctx context.Context, wallet *wallets.Wallet, amount int, needRead bool) ([]cloudcoin.CloudCoin, error) {
  logger.L(ctx).Debugf("Getting %d coins from %s", amount, wallet.Name)

  exps, err := v.GetHandleExps(ctx, wallet, amount)
  if err != nil {
    return nil, err
  }

  coinsToExport := make([]cloudcoin.CloudCoin, 0)
  for denomination, ccs := range(wallet.CoinsByDenomination) {
    for _, cc := range(ccs) {

      //logger.L(ctx).Debugf("coin %d ps %s, location %d, grade %s", cc.Sn, cc.PownString, cc.GetLocationStatus(), cc.GetGradeStatusString())
      logger.L(ctx).Debugf("coin %d location %d", cc.Sn, cc.GetLocationStatus())
      if needRead {
        err = storage.GetDriver().ReadCoin(ctx, wallet, cc)
        if err != nil {
          logger.L(ctx).Debugf("Failed to read coin %d: %s", cc.Sn, err.Error())
          continue
        }
      }

      if exps[denomination] > 0 {
        exps[denomination]--
        coinsToExport = append(coinsToExport, *cc)

        logger.L(ctx).Debugf("Added Coin %d d:%d", cc.Sn, cc.GetDenomination())
      }
    }
  }

  for d, v := range(exps) {
    if v != 0 {
      logger.L(ctx).Debugf("We forgot to pick d:%d. Left %d", d, v)
      return nil, perror.New(perror.ERROR_FAILED_TO_PICK_COINS, "Failed to pick coins")
    }
  }

  return coinsToExport, nil
}

func (v *Worker) GetSkyCoinsToDealWith(ctx context.Context, wallet *wallets.SkyWallet, amount int) ([]cloudcoin.CloudCoin, error) {
  logger.L(ctx).Debugf("Getting %d skycoins from %s", amount, wallet.Name)

  exps, err := v.GetHandleSkyExps(ctx, wallet, amount)
  if err != nil {
    return nil, err
  }

  coinsToExport := make([]cloudcoin.CloudCoin, 0)
  for denomination, ccs := range(wallet.CoinsByDenomination) {
    for _, cc := range(ccs) {
      if exps[denomination] > 0 {
        exps[denomination]--
        coinsToExport = append(coinsToExport, *cc)

        logger.L(ctx).Debugf("Added SkyCoin %d d:%d", cc.Sn, cc.GetDenomination())
      }
    }
  }

  for d, v := range(exps) {
    if v != 0 {
      logger.L(ctx).Debugf("We forgot to pick d:%d. Left %d", d, v)
      return nil, perror.New(perror.ERROR_FAILED_TO_PICK_COINS, "Failed to pick coins")
    }
  }

  return coinsToExport, nil
}

func (v *Worker) GetHandleSkyExps(ctx context.Context, wallet *wallets.SkyWallet, amount int) (map[int]int, error) {
  exps, err := v.GetExpCoins(ctx, amount, wallet.Denominations, false)
  if err != nil {
    perr := err.(*perror.ProgramError)
    if perr.Code == perror.ERROR_FAILED_TO_PICK_COINS {
      logger.L(ctx).Debugf("Can't pick coins. We will try to make change")
      err = v.BreakInBank(ctx, wallet)
      if err != nil {
        logger.L(ctx).Errorf("Failed to break coin for wallet %s: %s", wallet.Name, err.Error())
        return nil, perror.New(perror.ERROR_FAILED_TO_BREAK_COIN, "Failed to break coin")
      }

      // Now, try again
      exps, err = v.GetExpCoins(ctx, amount, wallet.Denominations, false)
      if err != nil {
        return nil, err
      }
    } else {
      return nil, err
    }
  }

  logger.L(ctx).Debugf("Exps %v", exps)

  return exps, nil
}

func (v *Worker) GetHandleExps(ctx context.Context, wallet *wallets.Wallet, amount int) (map[int]int, error) {
  exps, err := v.GetExpCoins(ctx, amount, wallet.Denominations, false)
  if err != nil {
    perr := err.(*perror.ProgramError)
    if perr.Code == perror.ERROR_FAILED_TO_PICK_COINS {
      logger.L(ctx).Debugf("Can't pick coins. We will try to make change")
      err = v.Break(ctx, wallet)
      if err != nil {
        logger.L(ctx).Errorf("Failed to break coin for wallet %s: %s", wallet.Name, err.Error())
        return nil, perror.New(perror.ERROR_FAILED_TO_BREAK_COIN, "Failed to break coin")
      }

      // Now, try again
      exps, err = v.GetExpCoins(ctx, amount, wallet.Denominations, false)
      if err != nil {
        return nil, err
      }
    } else {
      return nil, err
    }
  }

  //logger.L(ctx).Debugf("Exps %v", exps)

  return exps, nil
}



func (v *Worker) GetStatementTransactionType(itype int) string {
  ttype := "Unknown"

  switch itype {
  case config.STATEMENT_TYPE_DEPOSIT:
    ttype = "deposit"
  case config.STATEMENT_TYPE_SENT:
    ttype = "sent"
  case config.STATEMENT_TYPE_TRANSFER:
    ttype = "transfer"
  case config.STATEMENT_TYPE_WITHDRAW:
    ttype = "withdraw"
  case config.STATEMENT_TYPE_BREAK:
    ttype = "break"
  case config.STATEMENT_TYPE_JOIN:
    ttype = "join"
  }

  return ttype
}


func (v *Worker) SendProgress(message string) {
  if v.progressChannel == nil {
    return
  }

  pb := tasks.ProgressBatch{
    Status: "running",
    Message: message,
    Code: 0,
    Data: nil,
    Progress: 1,
  }

  v.progressChannel <- pb
}

func (v *Worker) ValidateFolder(ctx context.Context, path string) error {
  st, err := os.Stat(path)
  if err != nil {
    logger.L(ctx).Errorf("Failed to stat folder %s: %s", path, err.Error())
    return perror.New(perror.ERROR_BACKUP_FOLDER, "Failed to stat folder " + err.Error())
  }

  if !st.IsDir() {
    logger.L(ctx).Errorf("folder %s is not a directory", path)
    return perror.New(perror.ERROR_BACKUP_FOLDER, "Folder is not a directory")
  }

  return nil
}

func (v *Worker) InsertPNGData(ctx context.Context, bytes []byte, data []byte) ([]byte, error) {
  idx := utils.BasicPNGChecks(ctx, bytes)
	if idx == -1 {
    logger.L(ctx).Debugf("PNG Unpacker can't find PNG structure")
    return nil, perror.New(perror.ERROR_INVALID_PNG_HEADER, "Corrupted PNG header")
  }

  dlen := len(data)

  logger.L(ctx).Debugf("Embedding PNG data %d(%x) in size in databytes of %d size (idx %d)", dlen, dlen, len(bytes), idx)


  bcopy := make([]byte, len(bytes))
  copy(bcopy, bytes)

  s0 := bytes[0:idx + 4]

  b0 := byte((dlen >> 24) & 0xff)
  b1 := byte((dlen >> 16) & 0xff)
  b2 := byte((dlen >> 8) & 0xff)
  b3 := byte((dlen) & 0xff)

  logger.L(ctx).Debugf("%x %x %x %x", b0, b1, b2, b3)

  // Length
  s0 = append(s0, b0, b1, b2, b3)

  // cLDc
  s0 = append(s0, 0x63, 0x4c, 0x44, 0x63)
  
  // Data
  s0 = append(s0, data...)

  // CheckSum
  crc32 := utils.CalcCrc32(s0[idx + 8:])

  logger.L(ctx).Debugf("crc %x (%d %d)", crc32, len(s0), len(s0[idx+8:]))

  b0 = byte((crc32 >> 24) & 0xff)
  b1 = byte((crc32 >> 16) & 0xff)
  b2 = byte((crc32 >> 8) & 0xff)
  b3 = byte((crc32) & 0xff)

  s0 = append(s0, b0, b1, b2, b3)

  // Rest of the PNG
  s0 = append(s0, bcopy[idx+4:]...)

  return s0, nil
}
