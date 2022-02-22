package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
)

type Response struct {
  Payload interface{} `json:"payload"`
  Status string `json:"status"`
}

type DummyPayload struct {
}

func ErrorResponse(ctx context.Context, w http.ResponseWriter, err error) {
  rq := Response{}
  rq.Status = "error"
  rq.Payload = err

	logger.L(ctx).Debugf("Response Error %s", err.Error())

  json.NewEncoder(w).Encode(rq)
}

func SuccessResponse(ctx context.Context, w http.ResponseWriter, payload interface{}) {
  rq := Response{}
  rq.Status = "success"

  if payload == nil {
    rq.Payload = DummyPayload{}
  } else {
    rq.Payload = payload
  }

  err := json.NewEncoder(w).Encode(rq)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_ENCODE_OUTPUT_JSON, "Failed to encode output JSON: " + err.Error()))
  }

	logger.L(ctx).Debugf("Response successfull")
}
