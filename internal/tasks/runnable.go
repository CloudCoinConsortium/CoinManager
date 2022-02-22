package tasks

import "context"

type Runnable interface {
  Run(context.Context)
}
