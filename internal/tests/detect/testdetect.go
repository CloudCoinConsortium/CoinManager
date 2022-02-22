package testdetect

import (
	"context"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
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


func (v *Caller) DoCommand(ctx context.Context, args []string) (interface{}, error) {
  logger.L(ctx).Debugf("Test detect")

  if len(args) != 1 {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "Coin number is required")
  }

  totalCoins, _ := strconv.Atoi(args[0])
  if totalCoins > v.MaxCoins {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "MaxCoins exceeded")
  }


  coins := make([]cloudcoin.CloudCoin, 0)
  for i := v.StartSN; i < v.StartSN + totalCoins; i++ {
    coin := v.GetCoin(uint32(i))
    coins = append(coins, *coin)
  }


  r := raida.NewDetect(nil)
  out, err := r.Detect(ctx, coins)
  if err != nil {
    return "", err
  }

  return out, nil
}
