package unpacker

import (
	"context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
)


type JsonUnpacker struct {
}




func NewJsonUnpacker() *JsonUnpacker {
  return &JsonUnpacker{}
}

func (v *JsonUnpacker) GetName() string {
  return "json"
}

func (v *JsonUnpacker) Unpack(ctx context.Context, data []byte) ([]cloudcoin.CloudCoin, error) {
  logger.L(ctx).Debugf("Unpacking data len %d", len(data))

  coins, err := UnmarshalJson(ctx, data)
  if err != nil {
    return nil, err
  }

  logger.L(ctx).Debugf("Unpacked %d coins", len(coins))

  return coins, err
}
