package testfix

import (
	"context"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/fix"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	parent "github.com/CloudCoinConsortium/superraidaclientbackend/internal/tests"
)

type Caller struct {
  parent.Parent
}

func New() *Caller {
  return &Caller{
    *parent.New(),
  }
}

type FResponse struct {
  PreAns, PostAns map[uint32][]string
  DetectPre, DetectPost *raida.DetectOutput
  Fix *fix.FixResult
}

func (v *Caller) DoCommand(ctx context.Context, args []string) (interface{}, error) {
  logger.L(ctx).Debugf("Test Fix")

  if len(args) < 1 {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "Coin number is required")
  }

  totalCoins, _ := strconv.Atoi(args[0])

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

  for idx, pi := range(args[1:]) {
    cc := &coins[idx]

    cc.SetPownString(pi)
    for ridx, status := range(cc.Statuses) {
      if status == config.COIN_STATUS_COUNTERFEIT {
        coins[idx].SetAn(ridx, "00000000000000000000000000000088")
      }
    }
  }


  r := raida.NewDetect(nil)
  out, err := r.Detect(ctx, coins)
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

    logger.L(ctx).Debugf("Will try to fix coin %d: %s. Batch %d", cc.Sn, cc.PownString, batchNumber)
    for ridx, status := range(cc.Statuses) {
      logger.L(ctx).Debugf("coin %d r%d st=%d", cc.Sn, ridx, status)

      if status == config.COIN_STATUS_COUNTERFEIT {
        logger.L(ctx).Debugf("coin %d failed on raida %d", cc.Sn, ridx)
        coins[idx].SetAn(ridx, "00000000000000000000000000000099")
        _, ok := fbatches[batchNumber].CoinsPerRaida[ridx]
        if !ok {
          fbatches[batchNumber].CoinsPerRaida[ridx] = make([]cloudcoin.CloudCoin, 0)
        }

        fbatches[batchNumber].CoinsPerRaida[ridx] = append(fbatches[batchNumber].CoinsPerRaida[ridx], coins[idx])
      }

    }

    logger.L(ctx).Debugf("ccc %v", fbatches)
  }


  for bn, _ := range(fbatches) {
    logger.L(ctx).Debugf("batch #%d", bn)
    for ridx, _ := range(fbatches[bn].CoinsPerRaida) {
      logger.L(ctx).Debugf("raida%d", ridx)
      for _, cc := range(fbatches[bn].CoinsPerRaida[ridx]) {
        logger.L(ctx).Debugf("sn %d", cc.Sn) 
      }
    }
  }

  
  fixer, _ := fix.New(nil)
  fr, err := fixer.Fix(ctx, nil, fbatches, true)
  if err != nil {
    return "", err
  }
  
  rr.Fix = fr


  for _, cc := range(coins) {
    for i :=0; i < 25;i++ {
      rr.PostAns[cc.Sn][i] = cc.Ans[i]
    }
  }

  /*r := raida.NewEcho(nil)
  out, err := r.Echo()
  if err != nil {
    return "", err
  }
  */

  rd := raida.NewDetect(nil)
  out2, err := rd.Detect(ctx, coins)
  if err != nil {
    return "", err
  }

  rr.DetectPost = out2
  
  return rr, nil
}
