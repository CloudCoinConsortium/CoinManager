package legacyraida

import (
	"bytes"
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
)

type DetectionAgent struct {
	index int
  progressChannel chan interface{}
  ActiveURL string
  Timeout int
}

type Result struct {
	ErrCode int
	Message string
	Index   int
	Data    string
}

func (da *DetectionAgent) log(ctx context.Context, message string, a ...interface{}) {
	prefix := "LegacyRAIDA" + strconv.Itoa(da.index)

  message = prefix + ", " + message
	logger.L(ctx).Debugf(message, a...)
}

func (da *DetectionAgent) logError(ctx context.Context, message string, a ...interface{}) {
	prefix := "LegacyRAIDA" + strconv.Itoa(da.index)

  message = prefix + ", " + message
	logger.L(ctx).Errorf(message, a...)
}

func (da *DetectionAgent) RecordTimeBytes(ctx context.Context, startTs int64, l int) {
	endTs := time.Now().UnixNano()
	diff := (endTs - startTs) / 1000000

	da.log(ctx, "Read %d bytes, Request Took %d ms", l, diff)
}


func NewDetectionAgent(index int, progressChannel chan interface{}) *DetectionAgent {
	return &DetectionAgent{
		index: index,
    progressChannel: progressChannel,
    ActiveURL: "https://raida" + strconv.Itoa(index) + "." + config.LEGACY_RAIDA_DOMAIN_NAME,
    Timeout : config.DEFAULT_BASE_TIMEOUT,
	}
}

func (da *DetectionAgent) SendProgress() {
  if da.progressChannel == nil {
    return
  }

  pb := tasks.ProgressBatch{
    Status: "running",
    Message: "Doing command on OldRaida " + strconv.Itoa(da.index),
    Code: 0,
    Data: da.index,
    Progress: 1,
  }

  da.progressChannel <- pb
}

func (da *DetectionAgent) SetTimeout(timeout int) {
  da.Timeout = timeout
}

func (da *DetectionAgent) SendRequest(ctx context.Context, command string, data string, done chan Result) {
  result := &Result{}
	result.Index = da.index

  da.log(ctx, "Sending %s request to legacy raida, timeout %d", command, da.Timeout)
  da.log(ctx, "data %s", data)

  startTs := time.Now().UnixNano()
  client := http.Client{
    Timeout: time.Duration(da.Timeout) * time.Second,
  }

  responseBody := bytes.NewBuffer([]byte(data))

  url := da.ActiveURL + "/service/" + command
  response, err := client.Post(url, "application/x-www-form-urlencoded", responseBody)
  if err != nil {
    da.logError(ctx, "Failed to send command: " + err.Error())
    result.ErrCode = config.REMOTE_RESULT_ERROR_COMMON
    da.SendProgress()
  	done <- *result
    return
  }

  defer response.Body.Close()

  body, err := ioutil.ReadAll(response.Body)
  if err != nil {
    da.logError(ctx, "Failed to read body: " + err.Error())

    e, ok := err.(net.Error)
    if !ok || !e.Timeout() {
      da.logError(ctx, "Error %s", err.Error())
      result.ErrCode = config.REMOTE_RESULT_ERROR_COMMON
    } else {
      da.logError(ctx, "Timeout")
      result.ErrCode = config.REMOTE_RESULT_ERROR_TIMEOUT
    }

    da.SendProgress()
  	done <- *result
    return
  }

  logger.L(ctx).Debugf("Read body %s", body)

	da.RecordTimeBytes(ctx, startTs, len(body))
  da.SendProgress()

	result.Message = "OK"
	result.ErrCode = config.REMOTE_RESULT_ERROR_NONE
  result.Data = string(body)
 	done <- *result

}
