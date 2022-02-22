package api

import (
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
)

type InfoResponse struct {
  Version string `json:"version"`
}

func InfoReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  logger.L(ctx).Debugf("Info Request")

  info := &InfoResponse{
  }
  
  info.Version = config.VERSION

  SuccessResponse(ctx, w, info)
}

