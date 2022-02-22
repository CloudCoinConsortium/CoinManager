package testecho

import (
	"context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	parent "github.com/CloudCoinConsortium/superraidaclientbackend/internal/tests"
)

type Caller struct {
  parent.Parent
}

func New() *Caller {
  return &Caller{
    parent.Parent{},
  }
}

func (v *Caller) DoCommand(ctx context.Context, args []string) (interface{}, error) {
  logger.L(ctx).Debugf("Test echo")

  r := raida.NewEcho(nil)
  r.SetPrivateActiveRaidaList(&raida.RaidaList.PrimaryRaidaList)
  out, err := r.Echo(ctx)
  if err != nil {
    return "", err
  }

  return out, nil
}
