package api

import (
	"context"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
)



type EchoCall struct {
  Parent
}

func NewEcho(progressChannel chan interface{}) *EchoCall {
  return &EchoCall{
    *NewParent(progressChannel),
  }
}

func (v *EchoCall) Run(ctx context.Context) {
    s := raida.NewEcho(v.ProgressChannel)
    s.SetPrivateActiveRaidaList(&raida.RaidaList.PrimaryRaidaList)
    response, err := s.Echo(ctx)
    if err != nil {
      v.SendError(ctx, err)
      return
    }
   
    v.SendResult(response)
}

func EchoReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewEcho(task.ProgressChannel)

  task.Run(ctx, instance)

  SuccessResponse(ctx, w, task)
}
