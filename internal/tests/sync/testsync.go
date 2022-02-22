package testsync

import (
	"context"
	"io/ioutil"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/monitor"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	parent "github.com/CloudCoinConsortium/superraidaclientbackend/internal/tests"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/unpacker"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
)



type Caller struct {
  parent.Parent
}

func New() *Caller {
  return &Caller{
    *parent.New(),
  }
}

type Response struct {
  PreHealthCheck *raida.ShowCoinsByDenominationRawOutput
  PostHealthCheck *raida.ShowCoinsByDenominationRawOutput
}


func (v *Caller) DoCommand(ctx context.Context, args []string) (interface{}, error) {
  logger.L(ctx).Debugf("Test Sync")

  if len(args) < 2 {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "Path and Command are required")
  }

  path := args[0]
  command := args[1]

  bytes, err := ioutil.ReadFile(path)
  if err != nil {
    logger.L(ctx).Errorf("Failed to read coin %s: %s", path, err.Error())
		return nil, perror.New(perror.ERROR_READING_COIN, "Failed to read ID coin")
  }

  u := unpacker.New()
  ccs, err := u.Unpack(ctx, bytes)
  if err != nil {
    logger.L(ctx).Errorf("Failed to unpack ID coin %s: %s", path, err.Error())
   	return nil, perror.New(perror.ERROR_READING_COIN, "Failed to ubpack ID PNG coin")
  }
  if len(ccs) != 1 {
    return nil, perror.New(perror.ERROR_READING_COIN, "Number of coins in the PNG must be one. We have: " + strconv.Itoa(len(ccs)))
  }

  cc := &ccs[0]

  sr := raida.NewShowCoinsByDenomination(nil)
  response, err := sr.ShowCoinsByDenominationRaw(ctx, cc)
  if err != nil {
    logger.L(ctx).Errorf("Failed to show registry: %s", err.Error())
    return nil, err
  }

  r := &Response{}
  r.PreHealthCheck = response

  if command == "1" {
    return r, nil
  }

  // transfer
  if command == "2" {
    if len(args) != 3 {
      return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "RAIDA to test is required")
    }

    rN, _ := strconv.Atoi(args[2])

    logger.L(ctx).Debugf("transfer %d", rN)
    for i := 0; i < 25; i++ {
      if i == rN {
        raida.RaidaList.ActiveRaidaList[i] = &raida.RaidaList.PrimaryRaidaList[i]
      } else {
        raida.RaidaList.ActiveRaidaList[i] = nil
      }
    }

    coins := make([]cloudcoin.CloudCoin, 2)
    cc0 := cloudcoin.NewFromData(2301812)
    cc1 := cloudcoin.NewFromData(4301919)

    coins[0] = *cc0
    coins[1] = *cc1

    toSn := 62720
    receiptID, _ := utils.GenerateReceiptID()

    tr := raida.NewTransfer(nil)
    trr, err := tr.Transfer(ctx, cc, coins, uint32(toSn), receiptID, "test")
    logger.L(ctx).Debugf("tr %v, err %v", trr, err)
 /*   for i := 0; i < 25; i++ {
      raida.RaidaList.ActiveRaidaList[i] = &raida.RaidaList.PrimaryRaidaList[i]
    }
    */
    monitor.DoMonitorTask(ctx)
    sr2 := raida.NewShowCoinsByDenomination(nil)
    response2, err := sr2.ShowCoinsByDenominationRaw(ctx, cc)
    if err != nil {
      logger.L(ctx).Errorf("Failed to show registry: %s", err.Error())
      return nil, err
    }
    r.PostHealthCheck = response2

    return r, nil
  }


  // owner add/delete
  if command == "3" {
    if len(args) != 4 {
      return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "RAIDA to test is required")
    }

    rN, _ := strconv.Atoi(args[2])
    addS := args[3]

    var add bool
    if addS == "add" {
      add = true
    } else {
      add = false
    }

    coins := make([]cloudcoin.CloudCoin, 2)
    cc0 := cloudcoin.NewFromData(2301812)
    cc1 := cloudcoin.NewFromData(4301919)

    coins[0] = *cc0
    coins[1] = *cc1

    ows := raida.NewSync(nil)
    so, err := ows.Sync(ctx, rN, *cc, coins, add)

    logger.L(ctx).Debugf("so %v err %v", so, err)


    sr2 := raida.NewShowCoinsByDenomination(nil)
    response2, err := sr2.ShowCoinsByDenominationRaw(ctx, cc)
    if err != nil {
      logger.L(ctx).Errorf("Failed to show registry: %s", err.Error())
      return nil, err
    }
    r.PostHealthCheck = response2

    return r, nil
  }

/*
  sr2 := raida.NewShowCoinsByDenomination(nil)
  response2, err := sr2.ShowCoinsByDenominationRaw(cc)
  if err != nil {
    logger.L().Errorf("Failed to show registry: %s", err.Error())
    return nil, err
  }
  r.PostHealthCheck = response2
*/
  //totalCoins, _ := strconv.Atoi(args[0])
  return r, nil
/*
  if totalCoins > v.MaxCoins {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "MaxCoins exceeded")
  }

  if len(args) - 1 != totalCoins {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "You need to pass pownstrings")
  }
  
  coins := make([]cloudcoin.CloudCoin, 0)
  for i := v.StartSN; i < v.StartSN + totalCoins; i++ {
    coin := v.GetCoin(uint32(i))
    coins = append(coins, *coin)
  }



  r := raida.NewDetect(nil)
  out, err := r.Detect(coins)
  if err != nil {
    return "", err
  }

  rr := &FResponse{}
  rr.DetectPre = out
  
  rr.PreAns = make(map[uint32][]string)
  rr.PostAns = make(map[uint32][]string)
  for _, cc := range(coins) {
    rr.PreAns[cc.Sn] = make([]string, 25)
    rr.PostAns[cc.Sn] = make([]string, 25)
    for i :=0; i < 25;i++ {
      rr.PreAns[cc.Sn][i] = cc.Ans[i]
    }
  }

  // Doing fix
  plen := len(coins)
  fbatches := make([]fix.FixBatch, plen)
  for bn, _ := range(fbatches) {
    fbatches[bn].CoinsPerRaida = make(map[int][]cloudcoin.CloudCoin, 0)
  }

  var batchNumber int
  for idx, pi := range(args[1:]) {
    cc := &coins[idx]

    cc.SetPownString(pi)
    batchNumber = idx / config.MAX_NOTES_TO_SEND

    logger.L().Debugf("Will try to fix coin %d: %s. Batch %d", cc.Sn, cc.PownString, batchNumber)
    for ridx, status := range(cc.Statuses) {
      logger.L().Debugf("coin %d r%d st=%d", cc.Sn, ridx, status)

      if status == config.COIN_STATUS_COUNTERFEIT {
        logger.L().Debugf("coin %d failed on raida %d", cc.Sn, ridx)
        coins[idx].SetAn(ridx, "00000000000000000000000000000099")
        _, ok := fbatches[batchNumber].CoinsPerRaida[ridx]
        if !ok {
          fbatches[batchNumber].CoinsPerRaida[ridx] = make([]cloudcoin.CloudCoin, 0)
        }

        fbatches[batchNumber].CoinsPerRaida[ridx] = append(fbatches[batchNumber].CoinsPerRaida[ridx], coins[idx])
      }

    }

    logger.L().Debugf("ccc %v", fbatches)
  }


  for bn, _ := range(fbatches) {
    logger.L().Debugf("batch #%d", bn)
    for ridx, _ := range(fbatches[bn].CoinsPerRaida) {
      logger.L().Debugf("raida%d", ridx)
      for _, cc := range(fbatches[bn].CoinsPerRaida[ridx]) {
        logger.L().Debugf("sn %d", cc.Sn) 
      }
    }
  }

  
  fixer, _ := fix.New(nil)
  fr, err := fixer.Fix(nil, fbatches)
  if err != nil {
    return "", err
  }
  
  rr.Fix = fr


  for _, cc := range(coins) {
    for i :=0; i < 25;i++ {
      rr.PostAns[cc.Sn][i] = cc.Ans[i]
    }
  }


  rd := raida.NewDetect(nil)
  out2, err := rd.Detect(coins)
  if err != nil {
    return "", err
  }

  rr.DetectPost = out2
  
  return rr, nil
  */
}
