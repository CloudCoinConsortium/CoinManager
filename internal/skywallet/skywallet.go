package skywallet

import (
	"context"
	"strconv"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/dnsservice"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/worker"
	"github.com/golang/freetype/truetype"
)


type SkyWalleter struct {
  worker.Worker
  progressChannel chan interface{}
  font *truetype.Font
  rfont *truetype.Font
}

type Card struct {
  CardNumber string
  CardPIN string
  CardData string
}

func New(progressChannel chan interface{}) (*SkyWalleter) {
  return &SkyWalleter{
    *worker.New(progressChannel),
    progressChannel,
    nil,
    nil,
  }
}

func (v *SkyWalleter) RegisterNew(ctx context.Context, cc *cloudcoin.CloudCoin) error {
  ds := dnsservice.New(v.progressChannel)
  err := ds.RegisterName(ctx, cc)
  if err != nil {
    logger.L(ctx).Errorf("Failed to register name %s: %s", cc.GetSkyName(), err.Error())
    return err
  }

  err = storage.GetDriver().CreateSkyWallet(ctx, cc)
  if err != nil {
    logger.L(ctx).Errorf("Failed to create skywallet on FS. Will try to reverse")
    return err
  }

  return nil
}

func (v *SkyWalleter) RegisterNewCard(ctx context.Context, cc *cloudcoin.CloudCoin) (*Card, error) {
  logger.L(ctx).Debugf("Registering a new card for cc %d", cc.Sn)

  swallet, _ := v.GetIDOnly(ctx, cc.GetSkyName())
  if swallet != nil {
    return nil, perror.New(perror.ERROR_WALLET_ALREADY_EXISTS, "Skywallet already exists")
  }

  rand := utils.GenerateStrNumber(12)
  pin := utils.GenerateStrNumber(4)

  preCardNumber := "401" + rand
  reversed := utils.ReverseString(preCardNumber)

  logger.L(ctx).Debugf("rand %s pin %s (%s : %s)", rand, pin, preCardNumber, reversed)
  total := 0
  for idx, v := range(reversed) {
    num, _ := strconv.Atoi(string(v))
    if ((idx + 3) % 2) > 0 {
      num *= 2
      if num > 9 {
        num -= 9
      }
    }

    total += num
  }

  remainder := 10 - (total % 10)
  if remainder == 10 {
    remainder = 0
  }

  cardNumber := preCardNumber + strconv.Itoa(remainder)
  tmpNumber := ""
  for idx, ch := range (cardNumber) {
    tmpNumber += string(ch)
    logger.L(ctx).Debugf("i=%d %s", idx, tmpNumber)
    if ((idx + 1) % 4) == 0 {
      tmpNumber += " "
    }
  }
  cardNumber = tmpNumber

  // 13 (12 + 1 month)
  expDate := time.Now().AddDate(config.CARD_EXPIRATION_IN_YEARS, 0, 0)
  expString := expDate.Format("1/06")

  ip := cloudcoin.GetIPFromSn(cc.Sn)
  
  _ = expString

  logger.L(ctx).Debugf("%s %s total %d rand %s rem %d %s", cc.GetSkyName(), ip, total, rand, remainder, cardNumber)
  logger.L(ctx).Debugf("Setting pans")
  for i := 0; i < config.TOTAL_RAIDA_NUMBER; i++ {
    seed := strconv.Itoa(i) + strconv.Itoa(int(cc.Sn)) + rand + pin
    pan := utils.GetMD5(seed)

    logger.L(ctx).Debugf("seed %s (%s)", seed, pan)
    cc.SetPan(i, pan)
  }

  logger.L(ctx).Debugf("Powning")


  ncoins := make([]cloudcoin.CloudCoin, 1)
  ncoins[0] = *cc

  for idx, an := range(cc.Ans) {
    logger.L(ctx).Debugf("br%d %s", idx, an)
  }

  // Register dns
  ds := dnsservice.New(v.progressChannel)
  err := ds.RegisterName(ctx, cc)
  if err != nil {
    logger.L(ctx).Errorf("Failed to register name %s: %s", cc.GetSkyName(), err.Error())
    return nil, err
  }

  // Pown
  s := raida.NewPown(v.progressChannel)
  _, err = s.Pown(ctx, ncoins)
  if err != nil {
    logger.L(ctx).Debugf("Failed to pown ID coin %s", err.Error())
    return nil, err
  }


  logger.L(ctx).Debugf("powned %s", cc.GetPownString())
  for idx, an := range(cc.Ans) {
    logger.L(ctx).Debugf("ar%d %s", idx, an)
  }

  isAuthentic, _, _ := cc.IsAuthentic()
  if !isAuthentic {
    logger.L(ctx).Errorf("Coin is counterfeit %s", cc.GetPownString())
    return nil, perror.New(perror.ERROR_COIN_COUTERFEIT, "ID coin is counterfeit or failed to pown itself: " + cc.GetPownString())
  }


  bdata, err := v.DrawIDCard(ctx, cc, cardNumber, pin, expString, ip)
  if err != nil {
    logger.L(ctx).Debugf("Failed to draw card: %s", err.Error())
    storage.GetDriver().CreateSkyWallet(ctx, cc)
    return nil, perror.New(perror.ERROR_DRAW_PNG, "Failed to draw PNG card. The ID coin is saved as binary")
  }

  
  err = storage.GetDriver().CreateSkyWalletWithData(ctx, cc, bdata)
  if err != nil {
    logger.L(ctx).Debugf("Failed to save coin: %s", err.Error())
    return nil, err
  }
  

  logger.L(ctx).Debugf("ID created %v", bdata)

  
  return nil, nil
}

func (v *SkyWalleter) UpdateBalance(ctx context.Context, wallet *wallets.SkyWallet) error {
  b := raida.NewBalance(v.progressChannel)
  response, err := b.Balance(ctx, wallet.IDCoin)
  if err != nil {
    logger.L(ctx).Errorf("Failed to get balance: %s", err.Error())
    return perror.New(perror.ERROR_GET_BALANCE, "Failed to get balance: " + err.Error())
  }

  wallet.Denominations[1] = response.D1
  wallet.Denominations[5] = response.D5
  wallet.Denominations[25] = response.D25
  wallet.Denominations[100] = response.D100
  wallet.Denominations[250] = response.D250
  wallet.Balance = response.Total

  logger.L(ctx).Debugf("Updated Balance for %s: %v", wallet.Name, response)

  return nil
}

func (v *SkyWalleter) DeleteStatements(ctx context.Context, name string) error {
  logger.L(ctx).Debugf("Deleting statements for skywallet %s", name)

  skyWallet, err := v.GetIDOnly(ctx, name)
  if err != nil {
    return err
  }


  logger.L(ctx).Debugf("Wallet %s", skyWallet.Name)
  ds := raida.NewDeleteAllStatements(v.progressChannel)
  err = ds.DeleteAllStatements(ctx, skyWallet.IDCoin)
  if err != nil {
    logger.L(ctx).Errorf("Failed to delete all statements: %s", err.Error())
    return perror.New(perror.ERROR_FAILED_TO_DELETE_ALL_STATEMENTS, "Failed to delete statements on the RAIDA")
  }

  return nil
  

}

func (v *SkyWalleter) GetIDOnly(ctx context.Context, name string) (*wallets.SkyWallet, error) {
  wallet, err := storage.GetDriver().GetSkyWallet(ctx, name)
  if err != nil {
    logger.L(ctx).Errorf("Failed to get skywallet %s: %s", name, err.Error())
    return nil, err
  }

  return wallet, nil
}

func (v *SkyWalleter) GetByID(ctx context.Context, sn uint32) (*wallets.SkyWallet, error) {
  wallets, err := storage.GetDriver().GetSkyWallets(ctx)
  if err != nil {
    logger.L(ctx).Errorf("Failed to get skywallets: %s", err.Error())
    return nil, err
  }

  for _, wallet := range(wallets) {
    if wallet.IDCoin.Sn == sn {
      return &wallet, nil
    }
  }


  return nil, perror.New(perror.ERROR_WALLET_NOT_FOUND, "Wallet with sn " + strconv.Itoa(int(sn)) + " not found")
}

func (v *SkyWalleter) GetWithBalance(ctx context.Context, name string) (*wallets.SkyWallet, error) {
  logger.L(ctx).Debugf("Skywalleter GetWithBalance for %s", name)

  wallet, err := storage.GetDriver().GetSkyWallet(ctx, name)
  if err != nil {
    logger.L(ctx).Errorf("Failed to get skywallet %s: %s", name, err.Error())
    return nil, err
  }

  err = v.UpdateBalance(ctx, wallet)
  if err != nil {
    logger.L(ctx).Errorf("Failed to update balance: %s", err.Error())
    return nil, err
  }

  return wallet, nil
}

func (v *SkyWalleter) Get(ctx context.Context, name string) (*wallets.SkyWallet, error) {
  logger.L(ctx).Debugf("Skywalleter for %s", name)

  wallet, err := v.GetWithBalance(ctx, name)
  if err != nil {
    return nil, err
  }

  s := raida.NewShowStatement(v.progressChannel)
  sresponse, err := s.ShowStatement(ctx, wallet.IDCoin)
  if err != nil {
    return nil, err
  }

  wallet.Statements = make([]wallets.Statement, len(sresponse.Items))

  for idx, lv := range(sresponse.Items) {
    ttype := v.GetStatementTransactionType(lv.TransactionType)

    wallet.Statements[idx] = wallets.Statement{
      Guid: lv.Guid,
      Type: ttype,
      Amount: lv.Amount,
      Balance: lv.Balance,
      Time: lv.TimeStamp,
      Memo: lv.Memo,
    }
  }

  logger.L(ctx).Debugf("r %v", sresponse)

  return wallet, nil
}

func (v *SkyWalleter) GetWithContents(ctx context.Context, name string) (*wallets.SkyWallet, error) {
  sw, err := v.Get(ctx, name)
  if err != nil {
    return nil, err
  }

  err = v.UpdateSkyWalletContents(ctx, sw)
  if err != nil {
    return nil, err
  }

  return sw, nil
}

func (v *SkyWalleter) GetWithContentsOnly(ctx context.Context, name string) (*wallets.SkyWallet, error) {
  sw, err := v.GetIDOnly(ctx, name)
  if err != nil {
    return nil, err
  }

  err = v.UpdateSkyWalletContents(ctx, sw)
  if err != nil {
    return nil, err
  }

  // Calc balance
  for _, ccs := range(sw.CoinsByDenomination) {
    for _, cc := range(ccs) {
      d := cc.GetDenomination()
      sw.Balance += d
    }
  }
  logger.L(ctx).Debugf("Calculated balance for %s %d", name, sw.Balance)

  return sw, nil
}

func (v *SkyWalleter) UpdateSkyWalletContents(ctx context.Context, skyWallet *wallets.SkyWallet) error {
  // Empty previous values
  for _, d := range([]int{1, 5, 25, 100, 250}) {
    skyWallet.CoinsByDenomination[d] = make([]*cloudcoin.CloudCoin, 0)
    skyWallet.Denominations[d] = 0
  }

  sr := raida.NewShowCoinsByDenomination(v.progressChannel)
  response, err := sr.ShowCoinsByDenomination(ctx, skyWallet.IDCoin)
  if err != nil {
    logger.L(ctx).Errorf("Failed to show registry: %s", err.Error())
    return err
  }

  sns := response.SerialNumbers

  for _, sn := range(sns) {
    d := cloudcoin.GetDenomination(sn)
    cc := cloudcoin.NewFromData(sn)
    skyWallet.CoinsByDenomination[d] = append(skyWallet.CoinsByDenomination[d], cc)
    skyWallet.Denominations[d]++
  }

  return nil
}

func (v *SkyWalleter) Update(ctx context.Context, wallet *wallets.SkyWallet) error {
  logger.L(ctx).Debugf("Updating Skywallet %s", wallet.Name)

  err := v.UpdateBalance(ctx, wallet)
  if err != nil {
    return err
  }

  err = v.UpdateSkyWalletContents(ctx, wallet)
  if err != nil {
    return err
  }

  return nil
}
