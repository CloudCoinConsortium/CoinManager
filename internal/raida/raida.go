package raida

import (
	"context"
	"encoding/binary"
	"strconv"
	"sync"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"

	//	"fmt"
	"math"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	//	"config"
	//	"time"
)

type RAIDA struct {
	TotalNumber     int
	SideSize        int
	DetectionAgents []DetectionAgent
  PrivateActiveRaidaList *[]string
  BaseTimeouts map[uint16]int
}

type ServerList struct {
  ActiveRaidaList []*string
  PrimaryRaidaList []string
  BackupRaidaList []string
  Mutex *sync.Mutex
}

func New(progressChannel chan interface{}) *RAIDA {
	DetectionAgents := make([]DetectionAgent, config.TOTAL_RAIDA_NUMBER)
	for idx, _ := range DetectionAgents {
		DetectionAgents[idx] = *NewDetectionAgent(idx, progressChannel)
	}

	sideSize := int(math.Sqrt(config.TOTAL_RAIDA_NUMBER))
	if sideSize*sideSize != config.TOTAL_RAIDA_NUMBER {
		panic("Invalid RAIDA Configuration")
	}

  var baseTimeouts = map[uint16]int {
    COMMAND_ECHO: 2,
//    COMMAND_DETECT: 1,
//    COMMAND_POWN: 1,
//    COMMAND_DEPOSIT: 1,
//    COMMAND_GETTICKET: 1,
    COMMAND_FIX: 5000,
  }

	return &RAIDA{
		TotalNumber:     config.TOTAL_RAIDA_NUMBER,
		DetectionAgents: DetectionAgents,
		SideSize:        sideSize,
    BaseTimeouts:   baseTimeouts,
	}
}

/*
type Common struct {
	Server	string `json:"server"`
	Version string `json:"version"`
	Time	string `json:"time"`
}
*/

func (r *RAIDA) SetPrivateActiveRaidaList(list *[]string) {
  r.PrivateActiveRaidaList = list
  for idx, hostp := range(*list) {
    r.DetectionAgents[idx].ActiveURL = hostp
  }
}

func (r *RAIDA) FormErrorResults(code int, message string) []Result {
	results := make([]Result, r.TotalNumber)
	for idx := 0; idx < r.TotalNumber; idx++ {
    results[idx] = Result{
      Index: idx,
      ErrCode: code,
      Message: message,
    }
  }

  return results
}

func (r *RAIDA) SendRequest(ctx context.Context, bs [][]byte, bodyLen int) []Result {
  return r.SendRequestToSpecificRS(ctx, bs, bodyLen, nil)
}

func (r *RAIDA) InRList(idx int, rss []int) bool {
  if rss == nil {
    return true
  }

  for _, v := range(rss) {
    if v == idx {
      return true
    }
  }

  return false
}

func (r *RAIDA) SendRequestToSpecificRS(ctx context.Context, bs [][]byte, bodyLen int, rss []int) []Result {
  if rss != nil {
    logger.L(ctx).Debugf("Sending request to specific RAIDA servers %v", rss) 
  }

  var firstTimeout int
  isEcho := false
  for i := 0; i < r.TotalNumber; i++ {
    if !r.InRList(i, rss) {
      continue
    }

    b := bs[i]
    if len(b) < config.RAIDA_REQUEST_HEADER_SIZE {
      return r.FormErrorResults(config.REMOTE_RESULT_ERROR_SKIPPED, "Data is less that a header size for Raida" + strconv.Itoa(i))
    }

    command := binary.BigEndian.Uint16(b[4:6])
    commandName, ok := r.GetCommandName(command)
    if !ok {
      return r.FormErrorResults(config.REMOTE_RESULT_ERROR_SKIPPED, "Command " +strconv.Itoa(int(command)) + " is not supported for Raida" + strconv.Itoa(i))
    }

    timeout := r.CalcTimeout(command)
    //logger.L(ctx).Debugf("Raida%d Doing Command 0x%x: %s (timeout %d)", i, command, commandName, timeout)

    if command == COMMAND_ECHO {
      isEcho = true
    }

    firstTimeout, ok = r.BaseTimeouts[command]
    if !ok {
      firstTimeout = config.GLOBAL_FIRST_TIMEOUT_MS
    }

    r.DetectionAgents[i].SetTimeout(timeout)
    r.DetectionAgents[i].SetContext(commandName)
  }

	done := make(chan Result)
	for idx, _ := range r.DetectionAgents {
    if !r.InRList(idx, rss) {
      continue
    }

    b := bs[idx]

		go func(agent *DetectionAgent, b []byte) {
			agent.SendRequest(ctx, b, bodyLen, done, nil)
		}(&r.DetectionAgents[idx], b)
	}

	results := make([]Result, r.TotalNumber)
	chanResults := make([]*Result, r.TotalNumber)

  logger.L(ctx).Debugf("Requests to the RAIDA have been sent. Setting the first timeout to %d ms", config.GLOBAL_FIRST_TIMEOUT_MS)

  // We need to set errcodes before waiting from channel. Because the wait can time out and the rest of the response will be dropped without being set to 'SKIPPED'
	for i := 0; i < r.TotalNumber; i++ {
    if !r.InRList(i, rss) {
      chanResults[i] = &Result{}
      chanResults[i].Index = i
      chanResults[i].ErrCode = config.REMOTE_RESULT_ERROR_SKIPPED
    }
  }


  nextTimeout := config.GLOBAL_NEXT_2ND_DEVIATION_TIMEOUT_MS
  timeout := firstTimeout

  if !isEcho {
    logger.L(ctx).Debugf("First timeout is set to %d ms, Next timeout is previosly calculated, 2nd deviation: %dms", firstTimeout, nextTimeout)
    if nextTimeout == 0 {
      logger.L(ctx).Errorf("Critical error: 2nd deviation timeout was not calculated. Setting it to default %d", config.GLOBAL_NEXT_TIMEOUT_MS);
      nextTimeout = config.GLOBAL_NEXT_TIMEOUT_MS
    }
  }

BL:
	for i := 0; i < r.TotalNumber; i++ {
    if !r.InRList(i, rss) {
      continue
    }

    if !isEcho {
      select {
      case rcv := <-done:
        chanResults[i] = &rcv
        if chanResults[i] != nil && chanResults[i].ErrCode != config.REMOTE_RESULT_ERROR_SKIPPED && timeout == config.GLOBAL_FIRST_TIMEOUT_MS {
          logger.L(ctx).Debugf("Received first response from RAIDA %d. Setting next timeout to 2nd deviation %dms", chanResults[i].Index, nextTimeout)
          timeout = nextTimeout
        }
      case <- time.After(time.Duration(timeout) * time.Millisecond):
        var w string
        if i == 0 {
          w = "First"
        } else {
          w = "Next"
        }

        logger.L(ctx).Errorf("%s Wait timeout expired (%d ms). We no longer wait for other responses from RAIDA servers, %d response(s) discarded", w, timeout, (r.TotalNumber - i))
        break BL
      }
    } else {
      rcv := <-done
      chanResults[i] = &rcv
    }
	}

	logger.L(ctx).Debugf("Requests to the RAIDA completed")
  // First, default them as if they did not reply aniything
  for i := 0; i < r.TotalNumber; i++ {
    if !r.InRList(i, rss) {
      results[i] = Result{
        Index: i,
        ErrCode: config.REMOTE_RESULT_ERROR_SKIPPED,
      }
    } else {
      results[i] = Result{
        Index: i,
        ErrCode: config.REMOTE_RESULT_ERROR_TIMEOUT,
      }
    }
  }


	for _, result := range chanResults {
    if result == nil {
      continue
    }
    logger.L(ctx).Debugf("Response from RAIDA%d:  %v", result.Index, result)
		results[result.Index] = *result
	}

	return results
}

// Read remaining bytes from the previous request
func (r *RAIDA) ReadBytesFromDA(ctx context.Context, idx int, n int) Result {
	done := make(chan Result)

//  agent := r.DetectionAgents[idx]
  go func(agent *DetectionAgent) {
			agent.ReadBytes(ctx, n, done)
	}(&r.DetectionAgents[idx])

  chanResults := <-done

	logger.L(ctx).Debugf("Done readbytes request")

	return chanResults
}











/*
func (r *RAIDA) SendDefinedRequestRaw(url string, params []map[string]string) []Result {
	return r.sendDefinedRequest(url, params, nil, true, false, true)
}

func (r *RAIDA) SendDefinedRequestNoWait(url string, params []map[string]string, i interface{}) []Result {
	return r.sendDefinedRequest(url, params, i, false, false, false)
}

func (r *RAIDA) SendDefinedRequest(url string, params []map[string]string, i interface{}) []Result {
	return r.sendDefinedRequest(url, params, i, true, false, false)
}

func (r *RAIDA) SendDefinedRequestPost(url string, params []map[string]string, i interface{}) []Result {
	return r.sendDefinedRequest(url, params, i, true, true, false)
}

func (r *RAIDA) sendDefinedRequest(url string, params []map[string]string, i interface{}, wait bool, post bool, raw bool) []Result {
	logger.Info("Doing request " + url)

	done := make(chan Result)
	var doneIssued chan bool
	if !wait {
		doneIssued = make(chan bool)
	} else {
		doneIssued = nil
	}
	for idx, agent := range r.DetectionAgents {
		if params[idx] == nil {
			logger.Debug("Skipping Raida " + strconv.Itoa(idx))
			go func(idx int) {
				r := &Result{Index: idx, ErrCode: config.REMOTE_RESULT_ERROR_SKIPPED}
				if doneIssued != nil {
					doneIssued <- true
				}
				done <- *r
			}(idx)
			continue
		}

		if raw {
			go func(agent DetectionAgent, idx int) {
				agent.SendRequestRaw(url, params[idx], done, doneIssued, post)
			}(agent, idx)
		} else {
			go func(agent DetectionAgent, idx int) {
				agent.SendRequest(url, params[idx], done, doneIssued, post, reflect.TypeOf(i))
			}(agent, idx)
		}
	}

	if !wait {
		logger.Debug("Don't need to wait for completion. We will only wait till the requests are sent")
		for i := 0; i < r.TotalNumber; i++ {
			<-doneIssued
		}

		return nil
	}

	results := make([]Result, r.TotalNumber)
	chanResults := make([]Result, r.TotalNumber)
	for i := 0; i < r.TotalNumber; i++ {
		chanResults[i] = <-done
	}

	logger.Info("Done request " + url)
	for _, result := range chanResults {
		results[result.Index] = result
	}

	return results
}
*/

func (r *RAIDA) SetTimeout(timeout int) {
  for idx, _ := range(r.DetectionAgents) {
    r.DetectionAgents[idx].SetTimeout(timeout)
  }
}

func (r *RAIDA) TotalServers() int {
	return r.TotalNumber
}

func (r *RAIDA) GetSideSize() int {
	return r.SideSize
}

func (r *RAIDA) GetTimeoutMultiplier(command uint16) int {
  var command2config = map[uint16]int {
    COMMAND_ECHO: config.ECHO_TIMEOUT_MULT,
  }

  v, ok := command2config[command]
  if !ok {
    return config.DEFAULT_TIMEOUT_MULT
  }

  return v
}

func (r *RAIDA) CalcTimeout(command uint16) int {
  _, ok := r.GetCommandName(command)
  if !ok {
    return config.DEFAULT_BASE_TIMEOUT
  }

  var baseTimeout int
  baseTimeout, ok = r.BaseTimeouts[command]
  if !ok {
    baseTimeout = config.DEFAULT_BASE_TIMEOUT
  }

  mult := float32(r.GetTimeoutMultiplier(command)) / 100.0
  newTimeout := mult * float32(baseTimeout)

  return int(newTimeout)
}
