package raida

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
)

type DetectionAgent struct {
	index int
  conn net.Conn
  progressChannel chan interface{}
  ActiveURL string
  Timeout int
  CurrentContext string // CommandName mostly
}

type Result struct {
	ErrCode int
	Message string
	Index   int
	Data    []byte
  RequestTime int64
}

func (da *DetectionAgent) log(ctx context.Context, message string, a ...interface{}) {
	prefix := "RAIDA" + strconv.Itoa(da.index)

  if da.CurrentContext != "" {
    prefix += " command " + da.CurrentContext 
  }

  message = prefix + ", " + message
	logger.L(ctx).Debugf(message, a...)
}

func (da *DetectionAgent) logError(ctx context.Context, message string, a ...interface{}) {
	prefix := "RAIDA" + strconv.Itoa(da.index)

  if da.CurrentContext != "" {
    prefix += " command " + da.CurrentContext
  }

  message = prefix + ", " + message
	logger.L(ctx).Errorf(message, a...)
}

func (da *DetectionAgent) RecordTimeBytes(ctx context.Context, startTs int64, l int) int64 {
	endTs := time.Now().UnixNano()
	diff := (endTs - startTs) / 1000000

	da.log(ctx, "Read %d bytes, Request Took %d ms", l, diff)

  return diff
}


func NewDetectionAgent(index int, progressChannel chan interface{}) *DetectionAgent {

  RaidaList.Mutex.Lock()
  hostp := RaidaList.ActiveRaidaList[index]
  RaidaList.Mutex.Unlock()

  var activeURL string
  if hostp != nil {
    activeURL = *hostp
  } else {
    activeURL = ""
  }

	return &DetectionAgent{
		index: index,
		conn: nil,
    progressChannel: progressChannel,
    ActiveURL: activeURL,
    Timeout : config.DEFAULT_BASE_TIMEOUT,
	}
}

func (da *DetectionAgent) SendProgress() {
  if da.progressChannel == nil {
    return
  }

  pb := tasks.ProgressBatch{
    Status: "running",
    Message: "Last received response was from Raida " + strconv.Itoa(da.index),
    Code: 0,
    Data: da.index,
    Progress: 1,
  }

  da.progressChannel <- pb
}

func (da *DetectionAgent) InitConn(ctx context.Context) error {

  if (da.ActiveURL == "") {
    return perror.New(perror.ERROR_RAIDA_UNAVAILABLE, "RAIDA is Marked as Unavailable")
  } 
  
  hostURL := da.ActiveURL

  // TODO:delete it
  //port := 30000 + da.index
  //hostURL = config.DEFAULT_DOMAIN + ":" +strconv.Itoa(port)
  // Delete it

  da.log(ctx, "Init %s", hostURL)
  c, err := net.Dial("udp", hostURL)
  if err != nil {
    da.logError(ctx, "Failed to init DA " + hostURL)
    return perror.New(perror.ERROR_RAIDA_UNAVAILABLE, "Failed to Dial: " + err.Error())
  }

  da.conn = c

  return nil
}

func (da *DetectionAgent) SetContext(command string) {
  da.CurrentContext = command
}
func (da *DetectionAgent) SetTimeout(timeout int) {
  da.Timeout = timeout
}

func (da *DetectionAgent) SendRequest(ctx context.Context, params []byte, bodyLen int, done chan Result, doneIssued chan bool) {
  result := &Result{}
	result.Index = da.index

  err := da.InitConn(ctx)
  if err != nil {
    da.logError(ctx, "This RAIDA is not available. It will not be contacted: %s", err.Error())
    result.ErrCode = config.REMOTE_RESULT_ERROR_SKIPPED

    da.SendProgress()
  	done <- *result
    return
  }

  echoByte0 := int(params[12])
  echoByte1 := int(params[13])


  //sl := da.index * 1
  //if da.index == 1 {
  //  time.Sleep(time.Duration(5) * time.Second)
  //}

	startTs := time.Now().UnixNano()
  da.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(da.Timeout)))

	//fmt.Printf("p=%v\n",params)

  inputLength := len(params)
  packets := utils.GetUDPPacketCount(inputLength)

  for p := 0; p < packets; p++ {
    offset := p * config.MAX_RAIDA_DATAGRAM_SIZE
    length := offset + config.MAX_RAIDA_DATAGRAM_SIZE
    if length > inputLength {
      length = inputLength
    }

    n, err := da.conn.Write(params[offset:length])
    da.log(ctx, "Sent command %s: %v (sent %d bytes, total %d bytes, packets %d/%d). bufoffset/length=%d/%d Timeout %ds", da.CurrentContext, params, n, inputLength, p, packets, offset, length, da.Timeout)
    if err != nil {
      da.logError(ctx, "Failed to write to UDP socket: " + err.Error())
      result.ErrCode = config.REMOTE_RESULT_ERROR_COMMON

      da.SendProgress()
  	  done <- *result
      return
    }
  }


  expectedBytes := config.RAIDA_RESPONSE_HEADER_SIZE + bodyLen
/*
  if (da.index == 0) {
    expectedBytes += 1
  }*/

  buf := make([]byte, config.MAX_DGRAM_SIZE)
  readBytes := 0
  for {
    sbuf := buf[readBytes:]
    n, err := da.conn.Read(sbuf)
    if err != nil {
      e, ok := err.(net.Error)
      if !ok || !e.Timeout() {
        da.logError(ctx, "Error %s", err.Error())
        result.ErrCode = config.REMOTE_RESULT_ERROR_COMMON
      } else {
        da.logError(ctx, "Timeout")
        result.ErrCode = config.REMOTE_RESULT_ERROR_TIMEOUT
      }

			da.RecordTimeBytes(ctx, startTs, readBytes)
      da.SendProgress()
    	done <- *result
      return
    }

    readBytes += n

    if readBytes >= expectedBytes {
      break
    }
  }

  requestTime := da.RecordTimeBytes(ctx, startTs, readBytes)
  da.SendProgress()

  rIdx := int(buf[0])
  if rIdx != da.index {
    da.log(ctx, "Invalid rIdx returned %d for us", rIdx)
    result.ErrCode = config.REMOTE_RESULT_ERROR_COMMON
    done <- *result
    return
  }

  rEchoByte0 := int(buf[6])
  rEchoByte1 := int(buf[6])

  if (echoByte0 != rEchoByte0 || echoByte1 != rEchoByte1) {
    da.log(ctx, "Invalid echo bytes received 0x%x:0x%x, excpected 0x%x:0x%x", rEchoByte0, rEchoByte1, echoByte0, echoByte1)
    result.ErrCode = config.REMOTE_RESULT_ERROR_COMMON
    done <- *result
    return
  }

	result.Message = "OK"
	result.ErrCode = config.REMOTE_RESULT_ERROR_NONE
  result.Data = buf[:readBytes]
  result.RequestTime = requestTime
 	done <- *result

}

func (da *DetectionAgent) ReadBytes(ctx context.Context, bodyLen int, done chan Result) {
  expectedBytes := bodyLen
  da.log(ctx, "Reading more %d bytes", expectedBytes)

  result := &Result{}
	result.Index = da.index

  buf := make([]byte, config.MAX_DGRAM_SIZE)
  readBytes := 0
  for {
    sbuf := buf[readBytes:]
    n, err := da.conn.Read(sbuf)

    if err != nil {
      e, ok := err.(net.Error)
      if !ok || !e.Timeout() {
        da.logError(ctx, "Error: %s", e.Error())
        result.ErrCode = config.REMOTE_RESULT_ERROR_COMMON
      } else {
        da.logError(ctx, "Timeout")
        result.ErrCode = config.REMOTE_RESULT_ERROR_TIMEOUT
      }

    	done <- *result
      return
    }

    readBytes += n
    da.log(ctx, "Read %d v=%d (totalRead %d)", n, da.index, readBytes)

    if readBytes >= expectedBytes {
      da.log(ctx, "Packed read successfully")
      break
    }

  }

	result.Message = "OK"
	result.ErrCode = config.REMOTE_RESULT_ERROR_NONE

  result.Data = buf[:readBytes]
 	done <- *result
}



