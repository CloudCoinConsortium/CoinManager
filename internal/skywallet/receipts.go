package skywallet

import (
	"context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
)



func (v *SkyWalleter) AddReceipt(ctx context.Context, w *wallets.SkyWallet, st *wallets.StatementTransaction) error {
  logger.L(ctx).Debugf("Adding StatementTransaction for %s", w.Name)
  err := storage.GetDriver().AppendSkyTransactionDetails(ctx, w, st)
  if err != nil {
    return perror.New(perror.ERROR_FAILED_TO_UPDATE_WALLET_CONTENTS, "Failed to append transaction: " + err.Error())
  }

  return nil
}

func (v *SkyWalleter) GetReceipt(ctx context.Context, w *wallets.SkyWallet, guid string) (*wallets.StatementTransaction, error) {
  logger.L(ctx).Debugf("Getting StatementTransaction %s for %s", guid, w.Name)
  st, err := storage.GetDriver().GetSkyTransactionDetails(ctx, w, guid)
  if err != nil {
    return nil, perror.New(perror.ERROR_FAILED_TO_UPDATE_WALLET_CONTENTS, "Failed to get transaction: " + err.Error())
  }

  return st, nil
}
