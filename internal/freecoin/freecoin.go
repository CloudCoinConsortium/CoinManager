package freecoin

import (
	"math/rand"
	"time"
  "context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/worker"
)

type Freecoiner struct {
  worker.Worker
  progressChannel chan interface{}
}

type FreecoinResult struct {
  Sn uint32
  Ans []string
}

func New(progressChannel chan interface{}) (*Freecoiner, error) {
  return &Freecoiner{
    *worker.New(progressChannel),
    progressChannel,
  }, nil

}

func (v *Freecoiner) Get(ctx context.Context) (*FreecoinResult, error) {
  logger.L(ctx).Debugf("Getting a free coin. We have %d attempts", config.MAX_FREECOIN_ATTEMPTS)

  rand.Seed(time.Now().UnixNano())

  for i := 0; i < config.MAX_FREECOIN_ATTEMPTS; i++ {
    logger.L(ctx).Debugf("Attempt #%d", i)

    fcr, err := v.GetCoin(ctx)
    if err == nil {
      logger.L(ctx).Debugf("Got coin successfully: %d", fcr.Sn)
      return fcr, nil
    }

    logger.L(ctx).Debugf("Attempt failed: %d", err.Error())
  }

  logger.L(ctx).Errorf("No more attempts. Giving up")

  return nil, perror.New(perror.ERROR_FREECOIN_ATTEMPTS_REACHED, "Failed to get FreeCoin")
}

func (v *Freecoiner) GetRandomSN(ctx context.Context) uint32 {
  sn := rand.Intn(config.MAX_FREECOIN_SN - config.MIN_FREECOIN_SN) + config.MIN_FREECOIN_SN
  logger.L(ctx).Debugf("Generated SN for Freecoin %d", sn)

  return uint32(sn)
}

func (v *Freecoiner) GetCoin(ctx context.Context) (*FreecoinResult, error) {

  sn := v.GetRandomSN(ctx)

  f := raida.NewFreecoin(v.progressChannel)
  response, err := f.Get(ctx, sn)
  if err != nil {
    logger.L(ctx).Errorf("Failed to get a freecoin: %s", err.Error())
    return nil, err
  }

  freecoinResult := &FreecoinResult{}
  freecoinResult.Sn = sn
  freecoinResult.Ans = response.Ans

  return freecoinResult, nil
}

func (v *Freecoiner) GetSpecificCoin(ctx context.Context, sn uint32) (*FreecoinResult, error) {
  f := raida.NewFreecoin(v.progressChannel)
  response, err := f.Get(ctx, sn)
  if err != nil {
    logger.L(ctx).Errorf("Failed to get a freecoin: %s", err.Error())
    return nil, err
  }

  freecoinResult := &FreecoinResult{}
  freecoinResult.Sn = sn
  freecoinResult.Ans = response.Ans

  return freecoinResult, nil
}
