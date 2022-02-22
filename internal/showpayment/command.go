package showpayment

import (
	"regexp"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
)

type Caller struct {
}

func (v *Caller) DoCommand(args []string) (interface{}, error) {
  if len(args) != 1 {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "GUID is required for showpayment")
  }

  guid := args[0]

  r := regexp.MustCompile("^[a-fA-F0-9]{32}$")
  if !r.Match([]byte(guid)) {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "GUID is invalid")
  }

  sp := New(nil)
  out, err := sp.Show(nil, guid)
  if err != nil {
    return "", err
  }

  return out, nil
}
