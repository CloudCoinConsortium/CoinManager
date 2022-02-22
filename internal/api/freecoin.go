package api

import (
	"context"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/freecoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
)

type FreeCoinsponse struct {
  Sn int `json:"sn"`
  Ans []string `json:"ans"`
}

type FreeCoinCall struct {
  Parent
}

func NewFreeCoin(progressChannel chan interface{}) *FreeCoinCall {
  return &FreeCoinCall{
    *NewParent(progressChannel),
  }
}

func (v *FreeCoinCall) Run(ctx context.Context) {

  logger.L(ctx).Debugf("Getting Free Coin")

  freecoiner, err := freecoin.New(v.ProgressChannel)
  if err != nil {
    logger.L(ctx).Errorf("Failed to init freecoin: %s", err.Error())
    v.SendError(ctx, err)
    return
  }

  data, err := freecoiner.Get(ctx)
  if err != nil {
    v.SendError(ctx, err)
    return
  }

  v.SendResult(data)
}


func FreeCoinReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  logger.L(ctx).Debugf("FreeCoin Request")

  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewFreeCoin(task.ProgressChannel)
  task.Run(ctx, instance)

  SuccessResponse(ctx, w, task)
}

