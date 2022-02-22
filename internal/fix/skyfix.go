package fix

import (
	"context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
)

type SkyFixResult struct {
  AllFixed bool `json:"allfixed"`
  NumberFixes int `json:"number_fixed"`
  NumberFailures int `json:"number_failed"`
}

func (v *Fixer) SkyFix(ctx context.Context, wallet *wallets.SkyWallet, raidas []int) (*SkyFixResult, error) {
  logger.L(ctx).Debugf("Fixing ID coin from %s", wallet.Name)

  fr := &SkyFixResult{}

  cc := wallet.IDCoin

  tickets, err := v.UpdateTickets(ctx, nil, []cloudcoin.CloudCoin{*cc})
  if err != nil {
    logger.L(ctx).Errorf("Failed to get tickets: %s", err.Error())
    return nil, err
  }

  for _, ridx := range(raidas) {
    logger.L(ctx).Debugf("Fixing on raida %d", ridx)
    err := v.FixSkyCCOnRaida(ctx, ridx, cc, tickets)
    if err != nil {
      logger.L(ctx).Debugf("Failed to fix on raida %d: %s", ridx, err.Error())
      fr.NumberFailures++
      continue
    }

    fr.NumberFixes++
  }

  if cc.GetGradeStatus() == config.COIN_STATUS_AUTHENTIC {
    fr.AllFixed = true
  } 

  logger.L(ctx).Debugf("Fix result %v", cc.GetPownString())

  return fr, nil
}

func (v *Fixer) FixSkyCCOnRaida(ctx context.Context, rIdx int, cc *cloudcoin.CloudCoin, tickets []string) error {
  logger.L(ctx).Debugf("Fixing coins on Raida %d", rIdx)
  
  d := raida.NewFix(v.progressChannel)
  _, err := d.ProcessSkyFix(ctx, rIdx, cc, tickets)
  if err != nil {
    logger.L(ctx).Errorf("Failed to fix coins: %s", err.Error())
    return err
  }

  return nil
}
