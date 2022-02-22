package fix

import (
	"context"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/worker"
)

type Fixer struct {
  worker.Worker
  progressChannel chan interface{}
  Wallet *wallets.Wallet
}

type FixResult struct {
  TotalFixed  int `json:"total_fixed"`
  TotalErrors int `json:"total_errors"`
  TotalSkipped int `json:"total_skipped"`
  TotalFixesMade int `json:"total_fixes_made"`
  TotalFixesFailed int `json:"total_fixes_failed"`
}

func New(progressChannel chan interface{}) (*Fixer, error) {
  return &Fixer{
    *worker.New(progressChannel),
    progressChannel,
    nil,
  }, nil

}

func (v *Fixer) BatchFunction(ctx context.Context, coins []cloudcoin.CloudCoin) error {
  logger.L(ctx).Debugf("Calling batch function for fix")

  // Update coins in the same directory (to set more fracked) (Ignoring errors)
  _ = storage.GetDriver().UpdateCoins(ctx, v.Wallet, coins)
/*  if err != nil {
    logger.L(ctx).Errorf("Failed to set status for coins len=%d: %s", len(coins), err.Error())
    return err
  }
  */

  // Update status if the coin came authentic
  _ = storage.GetDriver().UpdateStatus(ctx, v.Wallet, coins)

  return nil
}

  
func (v *Fixer) CalcTotalBatch(batch FixBatch) int {
  total := 0
  
  for _, ccs := range(batch.CoinsPerRaida) {
    for _, cc := range(ccs) {
      total += cc.GetDenomination()
    }
  }

  return total
}

func (v *Fixer) Fix(ctx context.Context, wallet *wallets.Wallet, batches []FixBatch, noBf bool) (*FixResult, error) {
  if wallet == nil {
    logger.L(ctx).Debugf("Fixing %d batches of coins. No wallet", len(batches))
  } else {
    logger.L(ctx).Debugf("Fixing %d batches of coins from %s", len(batches), wallet.Name)
  }

  v.Wallet = wallet

  frGlobal := &FixResult{}
  for _, batch := range(batches) {
    fr, err := v.ProcessBatchFix(ctx, wallet, batch, noBf)
    if err != nil {
      logger.L(ctx).Warnf("Failed to fix coins %s", err.Error())
      frGlobal.TotalErrors += v.CalcTotalBatch(batch)
      continue
    }

    frGlobal.TotalErrors += fr.TotalErrors
    frGlobal.TotalSkipped += fr.TotalSkipped
    frGlobal.TotalFixed += fr.TotalFixed
    frGlobal.TotalFixesMade += fr.TotalFixesMade
    frGlobal.TotalFixesFailed += fr.TotalFixesFailed
  }

  return frGlobal, nil
}


func (v *Fixer) UpdateTickets(ctx context.Context, tickets []string, ccs []cloudcoin.CloudCoin) ([]string, error) {

  if tickets != nil {
    allset := true
    for idx, t := range(tickets) {
      if t == "" {
        allset = false
        logger.L(ctx).Debugf("Missing ticket for r%d. Will try to get it", idx)

        // No break here. We want to see all missing tickets
      }
    }

    if allset {
      logger.L(ctx).Debugf("We don't need tickets")
      return tickets, nil
    }
  }

  logger.L(ctx).Debugf("Need tickets")

  if len(ccs) == 0 {
    logger.L(ctx).Debugf("No cloudcoins to get tickets for")
    return tickets, nil
  }

  t := raida.NewGetTicket(v.progressChannel)
  response, err := t.GetTicket(ctx, ccs)
  if err != nil {
    logger.L(ctx).Errorf("Failed to get tickets: %s", err.Error())
    return nil, err
  }

  // We expect only 1 batch
  if len(response.Tickets) != 1 {
    return nil, perror.New(perror.ERROR_GET_TICKET, "Invalid number of batches in GetTicket. We expected only one. Got " + strconv.Itoa(len(response.Tickets)))
  }

  tickets = response.Tickets[0]


  return tickets, nil
}

// We make it exactly as GetTicket's stride. It should not be more than that!
func (v *Fixer) GetStrideSize() int {
  // packetSize = 1024 - header(22) - challenge(16) - signature(2)
  coinsLen := 1024 - 22 - 16 - 2

  // Sn + An
  coinItemSize := 3 + 16

  return coinsLen / coinItemSize
}

/*
func (v *Fixer) GetStrideSize() int {
  // packetSize = 1024 - header(22) - challenge(16) - signature(2)
  coinsLen := 1024 - 22 - 16 - 2

  // PG
  coinsLen -= 16

  // tickets
  coinsLen -= (8 * 25)

  // Sn 
  coinItemSize := 3

  return coinsLen / coinItemSize
}
*/

func (v *Fixer) ProcessBatchFix(ctx context.Context, wallet *wallets.Wallet, batch FixBatch, noBf bool) (*FixResult, error) {
  logger.L(ctx).Debugf("Fixing coins on %d raida(s)", len(batch.CoinsPerRaida))

  logger.L(ctx).Debugf("tickets %v", batch.Tickets)

  frLocal := &FixResult{}

  ccs := make([]cloudcoin.CloudCoin, 0)
  for _, coins := range(batch.CoinsPerRaida) {
    for _, cc := range(coins) {
      // Assume that the coin is in the Fracked
      cc.SetLocationStatus(config.COIN_LOCATION_STATUS_FRACKED)

      // Not necessary. We will generate PG later
      cc.GenerateMyPans()

      if wallet != nil {
        logger.L(ctx).Debugf("cc %d post %v", cc.Sn, cc.Ans)
        err := storage.GetDriver().ReadCoin(ctx, wallet, &cc)
        if err != nil {
          logger.L(ctx).Debugf("Failed to read coin %d: %s", cc.Sn, err.Error())
          frLocal.TotalSkipped += cc.GetDenomination()
          continue
        }
        logger.L(ctx).Debugf("cc %d post %v", cc.Sn, cc.Ans)
      }

      ccs = append(ccs, cc)
    }
  }

  if len(ccs) == 0 {
    logger.L(ctx).Debugf("No more coins to fix")
    return frLocal, nil
  }

  // The same ticket can be used multiple times
  tickets, err := v.UpdateTickets(ctx, batch.Tickets, ccs)
  if err != nil {
    logger.L(ctx).Errorf("Failed to get tickets: %s", err.Error())
    return nil, err
  }

  logger.L(ctx).Debugf("Updated tickets %v", tickets)

  for idx, ccs := range(batch.CoinsPerRaida) {
    logger.L(ctx).Debugf("Fixing batch on raida %d", idx)

    lfr, err := v.FixOnRaida(ctx, idx, ccs, tickets)
    if err != nil {
      logger.L(ctx).Debugf("Failed to fix on raida %d: %s", idx, err.Error())
      for _, lcc := range(ccs) {
        frLocal.TotalFixesFailed += lcc.GetDenomination()
      }
      continue
    }

    frLocal.TotalFixesMade += lfr.CoinsFixed
    frLocal.TotalFixesFailed += lfr.CoinsNotFixed

    if !noBf {
      v.BatchFunction(ctx, ccs)
    }
  }


  fsns := make(map[uint32]bool,0)

  // Collect status (graded on the RAIDA class already)
  for bidx, ccs := range(batch.CoinsPerRaida) {
    for _, cc := range(ccs) {
      logger.L(ctx).Debugf("Fix finally done for all batches. Raida#%d for cc %d: %s. Result: %s", bidx, cc.Sn, cc.GetPownString(), cc.GetGradeStatusString())
      // It is no enough to check if coin is Authentic. It can be authentic (e.g.) eppppnpppppppppppppppppp but failed to be fixed
      if cc.GetGradeStatus() == config.COIN_STATUS_AUTHENTIC {
        if cc.Statuses[bidx] == config.RAIDA_STATUS_PASS {
          // Only set to true if it wasn't set to false
          _, ok := fsns[cc.Sn]
          if !ok {
            fsns[cc.Sn] = true
          }
        } else {
          fsns[cc.Sn] = false
        }
      } else {
        // Don't set if it is 'true' already. Some coins can be in different bactches and be 'frack', 'frack but finally 'authentic'
        _, ok := fsns[cc.Sn]
        if !ok {
          fsns[cc.Sn] = false
        }

      }
    }
  }

  for k, v := range(fsns) {
    if v {
      frLocal.TotalFixed += cloudcoin.GetDenomination(k)
    } else {
      frLocal.TotalErrors += cloudcoin.GetDenomination(k)
    }
  }


  return frLocal, nil
}

func (v *Fixer) FixOnRaida(ctx context.Context, rIdx int, coins []cloudcoin.CloudCoin, tickets []string) (*raida.FixResult, error) {
  logger.L(ctx).Debugf("Fixing coins on Raida %d", rIdx)
  
  pg, _ := utils.GeneratePG()

  logger.L(ctx).Debugf("Generated pg %s", pg)

  d := raida.NewFix(v.progressChannel)
  // Will not set batch function. We call it manually
  lr, err := d.ProcessFix(ctx, rIdx, coins, tickets, pg)
  if err != nil {
    logger.L(ctx).Errorf("Failed to fix coins: %s", err.Error())
    return nil, err
  }

  return lr, nil
}
