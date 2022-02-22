package unpacker

import (
	"context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
)


type BinaryUnpacker struct {
}




func NewBinaryUnpacker() *BinaryUnpacker {
  return &BinaryUnpacker{}
}

func (v *BinaryUnpacker) GetName() string {
  return "binary"
}

func (v *BinaryUnpacker) Unpack(ctx context.Context, data []byte) ([]cloudcoin.CloudCoin, error) {
  logger.L(ctx).Debugf("Unpacking data, size %d", len(data))

  coins, err := UnmarshalBinary(ctx, data)
  if err != nil {
    return nil, perror.New(perror.ERROR_UNPACKER_NOT_MATCH, "No match:" + err.Error())
  }

  logger.L(ctx).Debugf("Unpacked %d coins", len(coins))

  return coins, nil
}
