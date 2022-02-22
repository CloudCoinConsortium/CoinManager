package api

import (
	"context"
	"regexp"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)


type Parent struct {
  ProgressChannel chan interface{}
}

type CoinRequest struct {
  Sn int `json:"sn"`
  Ans []string `json:"ans"`
  Pans []string `json:"pans"`
  CoinType int `json:"coin_type"`
}

func NewParent(progressChannel chan interface{}) *Parent {
  return &Parent{
    ProgressChannel: progressChannel,
  }
}

func (v *Parent) GetTotalIterations(coinCount int, strideSize int) int {
  return utils.GetTotalIterations(coinCount, strideSize)
}

func (v *Parent) SendProgress(message string, progress int) {
  pb := tasks.ProgressBatch{
    Status: config.TASK_STATUS_RUNNING,
    Message: message,
    Progress: progress,
    Code: 0,
    Data: nil,
  }

  v.ProgressChannel <- pb
}

func (v *Parent) SendResult(result interface{}) {
  pb := tasks.ProgressBatch{
    Status: config.TASK_STATUS_COMPLETED,
    Message: "Command Completed",
    Code: 0,
    Data: result,
    Progress: 100,
  }

  v.ProgressChannel <- pb
}

func (v *Parent) SendError(ctx context.Context, err error) {
  var message string
  var code int
  var details interface{}

  switch err.(type) {
  case (*perror.ProgramError):
    perr := err.(*perror.ProgramError)
    code = perr.Code
    message = perr.Message
    details = perr.Details
  default:
    code = perror.ERROR_INTERNAL
    message = err.Error()
    details = nil
  }

  logger.L(ctx).Debugf("Sending error %d %s", code, message)

  pb := tasks.ProgressBatch{
    Status: "error",
    Message: message,
    Code: code,
    Progress: 100,
    Data: details,
  }

  v.ProgressChannel <- pb
}

func ValidateCoin(v interface{}) error {
  coinRequest := v.(CoinRequest)

  err := validation.ValidateStruct(&coinRequest,
    validation.Field(&coinRequest.Sn, validation.Required, validation.Min(1), validation.Max(config.TOTAL_COINS)),
    validation.Field(&coinRequest.Ans, validation.Required, validation.Length(config.TOTAL_RAIDA_NUMBER, config.TOTAL_RAIDA_NUMBER), validation.Each(validation.Match(regexp.MustCompile("^[a-fA-F0-9]{32}$")))),
    validation.Field(&coinRequest.CoinType, validation.Min(0), validation.Max(65535)),
  )

  return err
}


func ValidateCoinWithPans(v interface{}) error {
  coinRequest := v.(CoinRequest)

  err := validation.ValidateStruct(&coinRequest,
    validation.Field(&coinRequest.Sn, validation.Required, validation.Min(1), validation.Max(config.TOTAL_COINS)),
    validation.Field(&coinRequest.Ans, validation.Required, validation.Length(config.TOTAL_RAIDA_NUMBER, config.TOTAL_RAIDA_NUMBER), validation.Each(validation.Match(regexp.MustCompile("^[a-fA-F0-9]{32}$")))),
    validation.Field(&coinRequest.Pans, validation.Required, validation.Length(config.TOTAL_RAIDA_NUMBER, config.TOTAL_RAIDA_NUMBER), validation.Each(validation.Match(regexp.MustCompile("^[a-fA-F0-9]{32}$")))),
  )

  return err
}

func ValidateTag(v interface{}) error {
  return validation.Validate(v, validation.Length(1, config.MAX_TAG_LENGTH), validation.Match(regexp.MustCompile("^[\\p{L}0-9_ ;^+()%$@#&:!-]*$")))
}

func ValidateWallet(v interface{}) error {
  return validation.Validate(v, validation.Required, validation.Length(config.WALLET_NAME_MIN_LENGTH, config.WALLET_NAME_MAX_LENGTH))
}

func ValidateSkyWallet(v interface{}) error {
  return validation.Validate(v, validation.Required, validation.Length(config.SKYWALLET_NAME_MIN_LENGTH, config.SKYWALLET_NAME_MAX_LENGTH), is.DNSName)
}

func ValidateGUID(v interface{}) error {
  return validation.Validate(v, validation.Required, validation.Match(regexp.MustCompile("^[a-fA-F0-9]{32}$")))
}
