package api

import (
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gorilla/mux"
)

type GetTaskRequest struct {
	Id string
}

func (s *GetTaskRequest) Validate() error {
  err := validation.ValidateStruct(s,
    validation.Field(&s.Id, validation.Required, is.UUIDv4 ),
  )

  return err
}

func TaskReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
	vars := mux.Vars(r)

  id, ok := vars["id"]
  if !ok {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_PARAMS_TASKID_MISSING, "TaskID Missing"))
    return
  }

  getTaskRequest := GetTaskRequest{
    Id: id,
  }
  err := getTaskRequest.Validate()
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_VALIDATE, "Validation error. " + err.Error()))
    return
  }

  task, err := tasks.GetTask(id)
  if err != nil {
    ErrorResponse(ctx, w, err)
    return
  }

  SuccessResponse(ctx, w, task)
}

