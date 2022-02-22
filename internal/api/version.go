package api

import (
	"context"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
)



type VersionCall struct {
  Parent
}

func NewVersion(progressChannel chan interface{}) *VersionCall {
  return &VersionCall{
    *NewParent(progressChannel),
  }
}

func (v *VersionCall) Run(ctx context.Context) {
  s := raida.NewVersion(v.ProgressChannel)
  response, err := s.Version(ctx)
  if err != nil {
    v.SendError(ctx, err)
    return
  }
   
  v.SendResult(response)
}

func VersionReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  task, err := tasks.CreateTask(ctx)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  instance := NewVersion(task.ProgressChannel)

  task.Run(ctx, instance)

  SuccessResponse(ctx, w, task)
}
