package unpacker

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)


type Unpacker interface {
  Unpack(context.Context, []byte) ([]cloudcoin.CloudCoin, error)
  GetName() string
}

type MainUnpacker struct {
  unpackers []Unpacker
}

type JsonStackFormat struct {
  CloudCoin []JsonFormat `json:"cloudcoin"`
}

type JsonFormat struct {
  Sn json.Number `json:"sn"`
  Ans []string `json:"an"`
}

type JsonFormatForValidation struct {
  Sn int
  Ans []string
}

func New() *MainUnpacker {

  unpackers := make([]Unpacker, 3)

  unpackers[0] = NewJsonUnpacker()
  unpackers[1] = NewBinaryUnpacker()
  unpackers[2] = NewPNGUnpacker()

  return &MainUnpacker{
    unpackers: unpackers,
  }
}

func ValidateCoin(cc JsonFormatForValidation) error {
  err := validation.ValidateStruct(&cc,
    validation.Field(&cc.Sn, validation.Required, validation.Min(1), validation.Max(config.TOTAL_COINS)),
    validation.Field(&cc.Ans, validation.Required, validation.Length(config.TOTAL_RAIDA_NUMBER, config.TOTAL_RAIDA_NUMBER), validation.Each(validation.Match(regexp.MustCompile("^[a-fA-F0-9]{32}$")))),
  )

  return err
}

// This is a common function used in almost all unpackers (JSON, PNG, JPEG)
func UnmarshalJson(ctx context.Context, data []byte) ([]cloudcoin.CloudCoin, error) {
  var stack JsonStackFormat

  err := json.Unmarshal(data, &stack)
  if err != nil {
    return nil, perror.New(perror.ERROR_UNPACKER_NOT_MATCH, "No match")
  }

  if (len(stack.CloudCoin) == 0) {
    return nil, perror.New(perror.ERROR_NO_CLOUDCOINS_IN_STACK, "Empty Stack")
  }

  coins := make([]cloudcoin.CloudCoin, 0)
  for _, cc := range stack.CloudCoin {
    snInt, _ := cc.Sn.Int64()
    ccfv := JsonFormatForValidation{
      Sn: int(snInt),
      Ans: cc.Ans,
    }

    err := ValidateCoin(ccfv)
    if err != nil {
      return nil, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error())
    }

    rcc := cloudcoin.NewFromData(uint32(snInt))
    rcc.Ans = cc.Ans

    coins = append(coins, *rcc)
  }

  return coins, nil
}



func UnmarshalBinary(ctx context.Context, data []byte) ([]cloudcoin.CloudCoin, error) {
  header := data[0:32]

  coinBodySize := 416

  coinBytes := len(data) - len(header)
  totalCoins := coinBytes / coinBodySize

  if coinBytes % coinBodySize > 0 {
    return nil, perror.New(perror.ERROR_INVALID_COIN_BODY_SIZE, "Invalid body size")
  }

  logger.L(ctx).Debugf("Assuming %d coins", totalCoins)

  var coins []cloudcoin.CloudCoin

  coins = make([]cloudcoin.CloudCoin, 0)
  for i := 0; i < totalCoins; i++ {
    start := i * coinBodySize + 32
    end := start + coinBodySize

    cc, err := cloudcoin.NewFromBinary(ctx, header, data[start:end])
    if err != nil {
      return nil, err
    }

    coins = append(coins, *cc)
  }

  if totalCoins == 0 {
    return nil, perror.New(perror.ERROR_NO_CLOUDCOINS_IN_STACK, "No coins")
  }


  return coins, nil
}


















func (v *MainUnpacker) Unpack(ctx context.Context, data []byte) ([]cloudcoin.CloudCoin, error) {
  logger.L(ctx).Debug("Lets go")


  filetype := http.DetectContentType(data)
  logger.L(ctx).Debugf("Detected filetype %s", filetype)


  if filetype == "application/x-gzip" || filetype == "application/zip" {
    logger.L(ctx).Debugf("Will unzip the data")

    zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
    if err != nil {
      logger.L(ctx).Errorf("Failed to unzip data: %s", err.Error())
      return nil, perror.New(perror.ERROR_UNZIP, "Failed to unzip data")
    }

    ccs := make([]cloudcoin.CloudCoin, 0)
    for _, zipFile := range zipReader.File {
      logger.L(ctx).Debugf("Unzipped file %s", zipFile.Name)

      fbytes, err := v.ReadZip(zipFile)
      if err != nil {
        logger.L(ctx).Errorf("Failed to read file: %s", err.Error())
        return nil, perror.New(perror.ERROR_UNZIP, "Failed to read file " + zipFile.Name + " in the archive: " + err.Error())
      }

      logger.L(ctx).Debugf("Unzipped %d bytes", len(fbytes))

      lccs, err := v.DoUnpack(ctx, fbytes)
      if err != nil {
        logger.L(ctx).Errorf("Failed to unpack unzipped bytes...")
        return nil, err
      }

      ccs = append(ccs, lccs...)
    }

    return ccs, nil
  }

  return v.DoUnpack(ctx, data)
}

func (v *MainUnpacker) ReadZip(zf *zip.File) ([]byte, error) {
  f, err := zf.Open()
  if err != nil {
    return nil, err
  }

  defer f.Close()

  return ioutil.ReadAll(f)
}

func (v *MainUnpacker) DoUnpack(ctx context.Context, data []byte) ([]cloudcoin.CloudCoin, error) {
  for _, unpacker := range(v.unpackers) {
    logger.L(ctx).Debugf("Trying Unpacker %s", unpacker.GetName())

    coins, err := unpacker.Unpack(ctx, data)
    if err != nil {
      perr := err.(*perror.ProgramError)
      if perr.Code == perror.ERROR_UNPACKER_NOT_MATCH {
        logger.L(ctx).Debugf("Unpacker %s cant unpack this data. Trying next unpacker", unpacker.GetName())
        continue
      }

      logger.L(ctx).Errorf("Unpacker %s failed to unpack the data", unpacker.GetName())
      return nil, err
    }

    return coins, nil
  }


  return nil, perror.New(perror.ERROR_UNPACK, "Failed to find the Unpacker to decode this data")
}
