package api

import (
	"net/http"
	"regexp"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/skywallet"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	//"github.com/go-ozzo/ozzo-validation/v4/is"
)

//func (v *WalletsCall) Run() {
//    s := raida.NewEcho(v.ProgressChannel)
//    s.Echo()
//}

type SenderHistoryResponse struct {
  Names []string `json:"names"`
}

type SenderHistoryRequest struct {
	Pattern string `json:"pattern"`
}

func (s *SenderHistoryRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Pattern, validation.Required),
  )

  return err
}

func SenderHistoryReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  logger.L(ctx).Debugf("Sender History Request")

  query := r.URL.Query()

  pattern := ".*"
  rpattern, ok := query["pattern"]
  if ok && len(rpattern) > 0 {
    pattern = rpattern[0]
  }

  logger.L(ctx).Debugf("pattern %s", pattern)

  senderHistoryRequest := SenderHistoryRequest{}
  senderHistoryRequest.Pattern = pattern
  err := senderHistoryRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }


  _, err = regexp.Compile(pattern)
  if err != nil {
    logger.L(ctx).Errorf("Invalid regex %s: %s", pattern, err.Error())
    ErrorResponse(ctx, w, perror.New(perror.ERROR_INVALID_REGEX, "Invalid regex pattern"))
    return
  }


  sw := skywallet.New(nil)
  records, err := sw.GetSenderHistory(ctx, pattern)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  response := &SenderHistoryResponse{}
  response.Names = records
  

  SuccessResponse(ctx, w, response)
  /*
*/
//  SuccessResponse(w, wallets)
}
