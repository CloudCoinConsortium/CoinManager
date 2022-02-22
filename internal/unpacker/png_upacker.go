package unpacker

import (
	"context"
	"encoding/binary"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
)

type PNGUnpacker struct {
}


func NewPNGUnpacker() *PNGUnpacker {
  return &PNGUnpacker{}
}

func (v *PNGUnpacker) GetName() string {
  return "PNG"
}

func (v *PNGUnpacker) Unpack(ctx context.Context, data []byte) ([]cloudcoin.CloudCoin, error) {
  byteValue := data

  logger.L(ctx).Debugf("PNG Unpacking data size %d", len(data))
  idx := utils.BasicPNGChecks(ctx, data)
	if idx == -1 {
    logger.L(ctx).Debugf("PNG Unpacker can't find PNG structure")
    return nil, perror.New(perror.ERROR_UNPACKER_NOT_MATCH, "No match")
  }

	i := 0
	var length int
	for {
		sidx := idx + 4 + i
		if sidx >= len(byteValue) {
			logger.L(ctx).Debugf("Failed to find stack in the PNG file")
      return nil, perror.New(perror.ERROR_NO_CLOUDCOINS_IN_STACK, "CloudCoin not found in the PNG file")
		}

		length = int(binary.BigEndian.Uint32(byteValue[sidx:]))
		if length == 0 {
			i += 12
			if i > len(byteValue) {
        return nil, perror.New(perror.ERROR_NO_CLOUDCOINS_IN_STACK, "CloudCoin not found in the PNG file")
			}
		}

		f := sidx + 4
		l := sidx + 8
		sig := string(byteValue[f:l])
		if sig == "cLDc" {
			crcSig := binary.BigEndian.Uint32(byteValue[sidx+8+length:])
			calcSig := utils.CalcCrc32(byteValue[f : f+length+4])

			if crcSig != calcSig {
				logger.L(ctx).Debugf("CRC32 is incorrect")
        return nil, perror.New(perror.ERROR_PNG_SIGNATURE, "Invalid PNG signature")
			}

			break
		}

		i += length + 12
		if i > len(byteValue) {
      return nil, perror.New(perror.ERROR_NO_CLOUDCOINS_IN_STACK, "CloudCoin not found in the PNG file")
		}
	}

	stringStack := string(byteValue[idx + 4 + i + 8 : idx + 4 + i + 8 + length])
	newByteValue := []byte(stringStack)

  coins, err := UnmarshalJson(ctx, newByteValue)
  if err != nil {
    logger.L(ctx).Debugf("Not JSON. Trying binary format in the PNG file")

    coins, err = UnmarshalBinary(ctx, newByteValue)
    logger.L(ctx).Debugf("sz %d", len(newByteValue))
    if err != nil {
      return nil, perror.New(perror.ERROR_VALIDATE, "Invalid stack in the PNG file. " + err.Error())
    }
  }

  return coins, nil
}

