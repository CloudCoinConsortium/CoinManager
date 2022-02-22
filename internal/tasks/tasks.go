package tasks

import (
	"context"
	"math"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/gofrs/uuid"
)


type ProgressBatch struct {
  Status string
  Message string
  Code int
  Data interface{}
  Progress int
}

type Task struct {
  Id string `json:"id"`
  Status string `json:"status"`
  Progress float64 `json:"progress"`
  Started int64 `json:"started"`
  Finished int64 `json:"finished"`
  Message string `json:"message"`
  Data interface{} `json:"data"`
  ItemCounters map[int]int `json:"-"`

  TotalIterations int `json:"-"`
  ProgressChannel chan interface{} `json:"-"`
}

var tasks map[string]*Task
var lock = sync.RWMutex{}

func InitTasks() {
  tasks = make(map[string]*Task)

  logger.L(nil).Debugf("Tasks initialized")
}

func GetTask(id string) (*Task, error) {
  lock.RLock()
  defer lock.RUnlock()

  task, ok := tasks[id]
  if !ok {
    return nil,  perror.New(perror.ERROR_TASK_NOT_FOUND, "Task does not exist")
  }

  return task, nil
}

func CreateTask(ctx context.Context) (*Task, error) {
  id, _ := uuid.NewV4()

  task := Task{}
  task.Status = config.TASK_STATUS_RUNNING
  task.Id = id.String()
  task.Progress = 0
  task.Started = time.Now().Unix()
  task.Finished = 0
  task.ProgressChannel = make(chan interface{}, 1)
  task.TotalIterations = config.TOTAL_RAIDA_NUMBER

  lock.Lock()
  tasks[task.Id] = &task
  lock.Unlock()

  logger.L(ctx).Debugf("Created task %s", task.Id)

  go WaitChannel(ctx, &task)

  CompletedTaskCollector()

  return &task, nil
}

func WaitChannel(ctx context.Context, task *Task) {

  pct := 0.0
  for {
    select {
    case packet := <- task.ProgressChannel:
      ourPacket := packet.(ProgressBatch)
      //logger.L(ctx).Debugf("Task %s received progress: %s (added %d point)", task.Id, ourPacket.Message, ourPacket.Progress)

      if ourPacket.Status == "error" {
        logger.L(ctx).Debugf("Task %s error. code %d" , task.Id, ourPacket.Code)
        task.SetError(ourPacket.Code, ourPacket.Message, ourPacket.Data)
        return
      }

      progress := ourPacket.Progress
      if progress != 0 {
        pct = (100.0 / float64(task.TotalIterations)) * float64(progress)
      } 

      //logger.L(ctx).Debugf("Appending progress %v%%. Total so far:%v%%", pct, task.Progress)

      task.Status = ourPacket.Status
      task.Progress += pct
      task.Message = ourPacket.Message
      task.Progress = math.Round(task.Progress * 100) / 100

      // Show dots about RAIDA
      if task.Status == config.TASK_STATUS_RUNNING && ourPacket.Data != nil {
        switch ourPacket.Data.(type) {
        case int:
          rIdx := ourPacket.Data.(int)
          //logger.L().Debugf("Task completed r %d", rIdx)

          if (len(task.ItemCounters) == 0) {
            task.ItemCounters = make(map[int]int, config.TOTAL_RAIDA_NUMBER)
          }

          task.ItemCounters[rIdx]++
          s := ""
          //for v := 0; v < len(task.ItemCounters); v++ {
          for v := 0; v < config.TOTAL_RAIDA_NUMBER; v++ {
            val, ok := task.ItemCounters[v]
            if !ok {
              s += "0"
              continue
            }

            s += strconv.Itoa(val)
          }

          task.Data = s

        default:
          task.Data = ourPacket.Data
        }
      } else {
        task.Data = ourPacket.Data
      }

      if (task.Status == config.TASK_STATUS_COMPLETED) {
        task.Message = "Command Completed"
        task.Progress = 100
        task.Finished = time.Now().Unix()
        return
      }

    case <- time.After(time.Second * config.TASK_TIMEOUT):
      logger.L(ctx).Debugf("Task %d timed out", task.Id)
      task.SetError(perror.ERROR_TASK_TIMEOUT, "Task timed out", nil)
      return
    }
  }
}

func CompletedTaskCollector() {
  lock.Lock()
  defer lock.Unlock()

  expireTime := time.Now().Add(time.Second * config.COMPLETED_TASK_LIFETIME)
  for id, task := range(tasks) {
    if task.Finished == 0 {
      continue
    }

    finishedTs := time.Unix(task.Finished, 0)
    if finishedTs.After(expireTime) {
      logger.L(nil).Debugf("Cleaning task %s", id)
      delete(tasks, id)
    }

  }
}

func (t *Task) SetError(code int, errorText string, details interface{}) {
  var rTexts []string
  
  if details != nil {
    rTexts = details.([]string)
  }

  t.Message = "Error Occured"
  t.Status = "error"
  t.Data = perror.NewWithDetails(code, errorText, rTexts)
  t.Progress = 100
  t.Finished = time.Now().Unix()
}

func (t *Task) SetTotalIterations(total int) {
  t.TotalIterations = total
}


func (t *Task) Run(ctx context.Context, runnable Runnable) {
  _, file, line, _ := runtime.Caller(1)

  dir := filepath.Dir(file)
  if dir != "" {
    dir = filepath.Base(dir)
  }
  filename := filepath.Base(file)

  shortPath := dir + "/" + filename

  logger.L(ctx).Debugf("Running task %s, %s:%d", t.Id, shortPath, line)

  go runnable.Run(ctx)
}
