package skywallet

import (
	"context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
)


func (v *SkyWalleter) AppendSenderHistory(ctx context.Context, sender string) error {
  logger.L(ctx).Debugf("Saving sender %s in history", sender)


  senders, err := storage.GetDriver().GetSenderHistory(ctx, ".*")
  if err != nil {
    return perror.New(perror.ERROR_ADD_SENDER_HISTORY, "Failed to get sender history: " + err.Error())
  }

  for _, lsender := range(senders) {
    if lsender == sender {
      logger.L(ctx).Debugf("This sender is in the history already")
      return nil
    }
  }


  err = storage.GetDriver().AppendSenderHistory(ctx, sender)
  if err != nil {
    return perror.New(perror.ERROR_ADD_SENDER_HISTORY, "Failed to add sender history: " + err.Error())
  }

  return nil
}

func (v *SkyWalleter) GetSenderHistory(ctx context.Context, pattern string) ([]string, error) {
  logger.L(ctx).Debugf("Get history, pattern %s", pattern)

  records, err := storage.GetDriver().GetSenderHistory(ctx, pattern)
  if err != nil {
    return nil, perror.New(perror.ERROR_GET_SENDER_HISTORY, "Failed to get sender history: " + err.Error())
  }

  return records, nil
}

