package perror

import (
	"encoding/json"
	"strconv"
)

type ProgramError struct {
  Code int `json:"code"`
  Message string `json:"message"`
  Details []string `json:"details"`
}

func New(code int, message string) *ProgramError {
  return &ProgramError{
    Code: code,
    Message: message,
  }
}

func NewWithDetails(code int, message string, details []string) *ProgramError {
  return &ProgramError{
    Code: code,
    Message: message,
    Details: details,
  }
}

func (s ProgramError) Error() string {
  return "ProgramError: " + strconv.Itoa(s.Code) + ", " + s.Message + ""
}

func (s ProgramError) ToJson() string {
  json, err := json.Marshal(s)
  if err != nil {
    return "json encode error"
  }

  return string(json)
}
