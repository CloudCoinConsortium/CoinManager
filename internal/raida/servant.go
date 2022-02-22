package raida

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
)

type Servant struct {
	Raida       *RAIDA
	repairArray [][]int
  progressChannel chan interface{}
  batchFunction BatchFunc
}

type CoinOutput struct {
  Sn uint32 `json:"sn"`
  PownString string `json:"pownstring"`
  Result string `json:"result"`
}

type Error struct {
	Message string
}

func NewServant(progressChannel chan interface{}) *Servant {
	//fmt.Println("new servant")
	Raida := New(progressChannel)

	repairArray := make([][]int, Raida.TotalServers())

	return &Servant{
		Raida:       Raida,
		repairArray: repairArray,
    progressChannel: progressChannel,
	}

}

func (s *Servant) SetPrivateActiveRaidaList(list *[]string) {
  s.Raida.SetPrivateActiveRaidaList(list)
}

func (s *Servant) SetVersion(headerBytes []byte, version int) {
  headerBytes[6] = byte(version)
}

func (s *Servant) GetHeader(command uint16, rIdx int, cce *cloudcoin.CloudCoin) []byte {
  return s.GetHeaderCommon(command, config.COINID_TYPE_CLOUD, cce, rIdx)
}

func (s *Servant) GetHeaderSky(command uint16, rIdx int, cce *cloudcoin.CloudCoin) []byte {
  return s.GetHeaderCommon(command, config.COINID_TYPE_SKY, cce, rIdx)
}

func (s *Servant) GetHeaderCommon(command uint16, coinIDType int, cce *cloudcoin.CloudCoin, rIdx int) []byte {
  data := []byte{
    0, // CloudID
    0, // SplitID
    byte(rIdx), // RAIDA ID (Will be overwritten)
    0, // Shard ID
    0, 0, // Command
    0, // Version ?
    0, 0, // CoinID zeroes (Big Endian) (will be overwritten)
    0, 0, 0, // 3 bytes Reserved
    0x11, 0x11, // Echo 2 bytes
    0, 0x1, // UDP Number (Little Endian)
    0, // Encryption
    0, 0, // CoinID for encryption
    0, 0, 0, // The SN of coin whose AN was used as a shared secret for encryption.
  }

  binary.BigEndian.PutUint16(data[4:6], command)
  binary.BigEndian.PutUint16(data[7:9], uint16(coinIDType))

  if cce != nil {
    // Encryption
    data[16] = 1

    if cce.IsIDCoin() {
      binary.BigEndian.PutUint16(data[17:19], config.COINID_TYPE_SKY)
    } else {
      binary.BigEndian.PutUint16(data[17:19], config.COINID_TYPE_CLOUD)
    }

    data[19] = byte((cce.Sn >> 16) & 0xff)
    data[20] = byte((cce.Sn >> 8) & 0xff)
    data[21] = byte(cce.Sn & 0xff)
  }

  return data
}

func (s *Servant) UpdateHeaderUdpPackets(params [][]byte) {
  if params == nil {
    return
  }

  for i := 0; i < s.Raida.TotalNumber; i++ {
    if params[i] == nil {
      continue
    }

    packets := utils.GetUDPPacketCount(len(params[i]))

    // UDP Packet Count
    binary.BigEndian.PutUint16(params[i][14:16], uint16(packets))
  }
}

func (s *Servant) EncryptIfRequired(ctx context.Context, cce *cloudcoin.CloudCoin, ridx int, data []byte) ([]byte, error) {
  if cce == nil {
    logger.L(ctx).Warnf("Data for raida%d will not be encrypted", ridx)
    return data, nil
  }

  return utils.Encrypt(ctx, cce.Ans[ridx], cce.Sn, ridx, data)
}

func (s *Servant) GetEncryptionCoin(ctx context.Context) (*cloudcoin.CloudCoin, error) {
  logger.L(ctx).Debugf("Getting Encryption Coin")


  if config.ENCRYPTION_DISABLED {
    logger.L(ctx).Debugf("Encryption is disabled in the config file")
    return nil, nil
  }

  wallet, err := storage.GetDriver().GetFirstSkyWallet(ctx)
  if err != nil {
    return nil, err
  }

  if wallet == nil {
    logger.L(ctx).Warnf("No ID coins. Trying to find a local coin")

    lwallets, err := storage.GetDriver().GetWallets(ctx) 
    if err != nil {
      return nil, err
    }

    for _, lwallet := range(lwallets) {
      err := storage.GetDriver().UpdateWalletBalance(ctx, &lwallet)
      if err != nil {
        return nil, err
      }

      if lwallet.Balance == 0 {
        logger.L(ctx).Debugf("Wallet %s is empty skipping it", lwallet.Name)
        continue
      }

      for _, dsns := range(lwallet.CoinsByDenomination) {
        if len(dsns) > 0 {
          cc := dsns[0]

          logger.L(ctx).Debugf("Trying sn for encryption %d (wallet %s)", cc.Sn, lwallet.Name)

          err = storage.GetDriver().ReadCoin(ctx, &lwallet, cc)
          if err != nil {
            logger.L(ctx).Errorf("Failed to read coin %d for encryption: %s", cc.Sn, err.Error())
            return nil, err
          }

          logger.L(ctx).Debugf("Coin %d from wallet %s will be used for encryption", cc.Sn, lwallet.Name)
          return cc, nil

        }
      }
    }

    return nil, perror.New(perror.ERROR_NO_COINS, "No coins for encryption. Can't do encryption")
  }

  cc := wallet.IDCoin

  logger.L(ctx).Debugf("Got ID coin %d for encryption (skywallet %s)", cc.Sn, wallet.Name)

  return cc, nil
}


func (s *Servant) GetChallenge() []byte {
  return []byte{ 
    0xee, 0xee, 0xee, 0xee, 0xee, 0xee, 0xee, 0xee,
    0xee, 0xee, 0xee, 0xee, 0xee, 0xee, 0xee, 0xee,
  }
}

func (s *Servant) GetShortChallenge() []byte {
  return []byte{ 
    0xee, 0xee, 0xee, 0xee, 
  }
}

func (s *Servant) GetSignature() []byte {
  return []byte{ 
    0x3e, 0x3e,
  }
}



func (s *Servant) GetPownStringFromStatusArray(statuses []int) string {
	var b strings.Builder
	var c string

	for _, status := range statuses {
		switch status {
		case config.RAIDA_STATUS_UNTRIED:
			c = "u"
		case config.RAIDA_STATUS_PASS:
			c = "p"
		case config.RAIDA_STATUS_FAIL:
			c = "f"
		case config.RAIDA_STATUS_ERROR:
			c = "e"
		case config.RAIDA_STATUS_NORESPONSE:
			c = "n"
		default:
			c = "e"
		}

		fmt.Fprintf(&b, "%s", c)
	}

	return b.String()
}

func (s *Servant) IsQuorumCollected(statuses []int) bool {
  oks := 0
	for i := 0; i < s.Raida.TotalServers(); i++ {
    if (statuses[i] == config.RAIDA_STATUS_PASS) {
      oks++
    }
  }

  if (oks >= config.MIN_QUORUM_COUNT) {
    return true
  }

  return false
}

func (s *Servant) SetCoinsStatus(coins []cloudcoin.CloudCoin, rIdx int, status int) {
  if coins == nil {
    return 
  }

  for idx, _ := range(coins) {
    coins[idx].SetDetectStatus(rIdx, status)
  }
}

type BatchFunc func(context.Context, []cloudcoin.CloudCoin) error
type SuccessFunc func(context.Context, int, int, []cloudcoin.CloudCoin, []byte) (int, interface{})

func (s *Servant) SetBatchFunction(batchFunction BatchFunc) {
  s.batchFunction = batchFunction
}


// Function that can return ALL_FAIL, ALL_PASS or MIXED
func (v *Servant) CommonMixedSuccessFunction(ctx context.Context, idx int, status int, coins []cloudcoin.CloudCoin, rdata []byte) (int, interface{}) {
  cmdName := v.Raida.DetectionAgents[idx].CurrentContext

  if (status == RESPONSE_STATUS_ALL_PASS) {
    logger.L(ctx).Debugf("RAIDA%d (%s) AllPass Status %d", idx, cmdName, status)
    v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_PASS)
    return config.RAIDA_STATUS_PASS, nil
  } 
  
  if (status == RESPONSE_STATUS_MIX) {
    logger.L(ctx).Debugf("RAIDA%d (%s) Mix Status %d", idx, cmdName, status)
    result := v.ReadMixedResultsAndUpdateCoins(ctx, idx, coins, rdata[config.RAIDA_RESPONSE_HEADER_SIZE:])

    return result, nil
  }

  if (status == RESPONSE_STATUS_ALL_FAIL) {
    logger.L(ctx).Errorf("RAIDA%d (%s) AllFail Status %d", idx, cmdName, status)

    // The coins themselves are failed
    v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_FAIL)

    // It is ok have 'PASS' there. It is not about the coins. This status is about RAIDA response itself
    return config.RAIDA_STATUS_PASS, nil
  } 
  
  logger.L(ctx).Errorf("RAIDA%d (%s) Invalid Status %d", idx, cmdName, status)
  v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_ERROR)

  return config.RAIDA_STATUS_ERROR, nil
}

// Function that can return SUCCESS or FAIL
func (v *Servant) CommonSuccessFunction(ctx context.Context, idx int, status int, coins []cloudcoin.CloudCoin, rdata []byte) (int, interface{}) {
  cmdName := v.Raida.DetectionAgents[idx].CurrentContext
  if (status == RESPONSE_STATUS_SUCCESS) {
    logger.L(ctx).Debugf("RAIDA%d (command %s) Success Status %d, rest data len %d", idx, cmdName, status, len(rdata))
    v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_PASS)
    return config.RAIDA_STATUS_PASS, rdata[config.RAIDA_RESPONSE_HEADER_SIZE:]
  }

  if (status == RESPONSE_STATUS_FAIL || status == RESPONSE_STATUS_FAILED_AUTH) {
    logger.L(ctx).Debugf("Raida%d Fail Status %d", idx, status)
    v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_FAIL)
    return config.RAIDA_STATUS_FAIL, nil
  } 
  
  logger.L(ctx).Debugf("Raida%d Unknown Status %d", idx, status)
  v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_ERROR)

  return config.RAIDA_STATUS_ERROR, nil
}

func (s *Servant) ProcessGenericResponses(ctx context.Context, coins []cloudcoin.CloudCoin, results []Result, successFunc SuccessFunc) ([]int, []interface{}) {
  return s.ProcessGenericResponsesCommon(ctx, coins, results, successFunc, nil)
}
func (s *Servant) ProcessGenericResponsesIgnoreUntried(ctx context.Context, coins []cloudcoin.CloudCoin, results []Result, successFunc SuccessFunc, cce *cloudcoin.CloudCoin) ([]int, []interface{}) {
  return s.ProcessGenericResponsesCommonB(ctx, coins, results, successFunc, cce, true)
}
func (s *Servant) ProcessGenericResponsesCommon(ctx context.Context, coins []cloudcoin.CloudCoin, results []Result, successFunc SuccessFunc, cce *cloudcoin.CloudCoin) ([]int, []interface{}) {
  return s.ProcessGenericResponsesCommonB(ctx, coins, results, successFunc, cce, false)
}

func (s *Servant) ProcessGenericResponsesCommonB(ctx context.Context, coins []cloudcoin.CloudCoin, results []Result, successFunc SuccessFunc, cce *cloudcoin.CloudCoin, ignoreUntried bool) ([]int, []interface{}) {
  pownArray := make([]int, s.Raida.TotalServers())
  responses := make([]interface{}, s.Raida.TotalServers())

	for idx, result := range results {
    cmdName := s.Raida.DetectionAgents[idx].CurrentContext

		if result.ErrCode == config.REMOTE_RESULT_ERROR_NONE {
      status := int(result.Data[2])

      data := result.Data
      var err error
      if cce != nil {
        logger.L(ctx).Debugf("RAIDA%d Decrypting response for command %s with CoinID %d", idx, cmdName, cce.Sn)

        dataEnc := data[config.RAIDA_RESPONSE_HEADER_SIZE:]
        data, err = utils.Decrypt(ctx, cce.Ans[idx], cce.Sn, idx, dataEnc)
        if err != nil {
          logger.L(ctx).Debugf("Failed to decrypt response from RAIDA %d with coinID %d", idx, cce.Sn)
          pownArray[idx] = config.RAIDA_STATUS_ERROR
          s.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_ERROR)
          continue
        }

        data = append(result.Data[:config.RAIDA_REQUEST_HEADER_SIZE], data...)
      }

      mainStatus, rv := successFunc(ctx, idx, status, coins, data)
      logger.L(ctx).Debugf("RAIDA%d (command %s) ReceivedStatus %d (datalen %d), ResponseParser returned %d", idx, cmdName, status, len(result.Data), mainStatus)

      pownArray[idx] = mainStatus
      responses[idx] = rv

    } else if result.ErrCode == config.REMOTE_RESULT_ERROR_TIMEOUT {
      pownArray[idx] = config.RAIDA_STATUS_NORESPONSE

      s.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_NORESPONSE)
    } else if result.ErrCode == config.REMOTE_RESULT_ERROR_SKIPPED {
      pownArray[idx] = config.RAIDA_STATUS_UNTRIED
      if !ignoreUntried {
        s.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_UNTRIED)
      }
    } else {
      pownArray[idx] = config.RAIDA_STATUS_ERROR
      s.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_ERROR)
    }

  }

  pownString := s.GetPownStringFromStatusArray(pownArray)
  logger.L(ctx).Debugf("General Request Status %s", pownString)

  return pownArray, responses
}


func (s *Servant) ReadMixedResultsAndUpdateCoins(ctx context.Context, idx int, coins []cloudcoin.CloudCoin, data []byte) int {
  needMoreBytes := int(math.Ceil(float64(len(coins)) / 8.0))
  receivedExtraBytes := len(data) 

  logger.L(ctx).Debugf("Need more %d bytes to process. Received %d extra bytes", needMoreBytes, receivedExtraBytes)

 // data := rdata[config.RAIDA_RESPONSE_HEADER_SIZE:]
  logger.L(ctx).Debugf("extra %v", data)
  if receivedExtraBytes < needMoreBytes {
    diffBytes := needMoreBytes - receivedExtraBytes  
    logger.L(ctx).Debugf("Will download more %d bytes", diffBytes)

    result := s.Raida.ReadBytesFromDA(ctx, idx, diffBytes)
    if result.ErrCode == config.REMOTE_RESULT_ERROR_NONE {
      logger.L(ctx).Debugf("Read extra bytes: %d", len(result.Data))
      data = append(data, result.Data...)
    } else if result.ErrCode == config.REMOTE_RESULT_ERROR_TIMEOUT {
      logger.L(ctx).Errorf("Timeout while reading extra bytes")
      s.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_NORESPONSE)
      return config.RAIDA_STATUS_NORESPONSE
    } else {
      logger.L(ctx).Errorf("Error while reading extra bytes")
      s.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_ERROR)
      return config.RAIDA_STATUS_ERROR
    }
  }

  logger.L(ctx).Debugf("Processing extra data %d bytes", len(data))
  if (len(data) * 8 < len(coins)) {
    logger.L(ctx).Errorf("Invalid data from raida %d. Received mix results with length %d, but we sent %d notes", idx, len(data), len(coins))
    s.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_ERROR)
    return config.RAIDA_STATUS_ERROR
  }

  for cidx, _ := range(coins) {
    coinByteIdx := int(cidx / 8)
    coinBitIdx := cidx % 8
    byteValue := data[coinByteIdx]

    bitValue := byteValue & (1 << coinBitIdx)
    status := config.RAIDA_STATUS_FAIL
    if bitValue > 0 {
      status = config.RAIDA_STATUS_PASS
    }

    logger.L(ctx).Debugf("R%d cIdx %d (sn #%d) byteIdx %d val %d bitidx %d idOk %d (%v)", idx, cidx, coins[cidx].Sn, coinByteIdx, byteValue, coinBitIdx, bitValue, status)
	  coins[cidx].SetDetectStatus(idx, status)
  }

  return config.RAIDA_STATUS_PASS
}



/* Gets the most occured error */

func (s *Servant) GetErrorsFromResults(results []Result) []string {
  tresults := make([]string, s.Raida.TotalServers())

  for idx, result := range(results) {
    if result.ErrCode == config.REMOTE_RESULT_ERROR_NONE {
      status := int(result.Data[2])
      tresults[idx] = "RaidaCode#" + strconv.Itoa(status)
      continue
    }

    tresults[idx] = "NetworkCode#"+ strconv.Itoa(result.ErrCode) + ": " + result.Message
  }

  return tresults
}

func (s *Servant) GetProgramErrors(err error) []string {
  results := []string{}

  var code int
  var message string
  switch err.(type) {
  case (*perror.ProgramError):
    perr := err.(*perror.ProgramError)
    code = perr.Code
    message = perr.Message
  default:
    code = perror.ERROR_INTERNAL
    message = err.Error()
  }

  for i := 0; i < config.TOTAL_RAIDA_NUMBER; i++ {
    results = append(results, "ProgramCode#" + strconv.Itoa(code) + ": " + message)
  }

  return results
}

func (s *Servant) GetMessageFromStripesAndMirrors(ctx context.Context, data []string) (string, error) {
  logger.L(ctx).Debugf("assembling %v", data)

  parts := make([][]string, s.Raida.TotalServers())
  for i := 0; i < s.Raida.TotalServers(); i++ {
    chunks := strings.Split(data[i], config.MEMO_SEPARATOR)
    if len(chunks) != 3 {
      logger.L(ctx).Warnf("Skipping chunk for RAIDA%d", i)
      parts[i] = nil
      continue
    }

    parts[i] = make([]string, 3)
    parts[i][0] = strings.Trim(chunks[0], string(0x0))
    parts[i][1] = strings.Trim(chunks[1], string(0x0))
    parts[i][2] = strings.Trim(chunks[2], string(0x0))
  }

  message, err := s.AssembleMessage(ctx, parts)
  if err != nil {
    return "", err
  }

	bytes, err := base64.StdEncoding.DecodeString(message)
  if err != nil {
    return "", perror.New(perror.ERROR_DECODE_BASE64, "Failed to decode based64 memo")
  }

  message = string(bytes)

  logger.L(ctx).Debugf("Assembled %s", message)

  return message, nil

}

func (s *Servant) GetStripesAndMirrors(message string) []string {

	str := base64.StdEncoding.EncodeToString([]byte(message))

	vals := s.SplitMessage(str)
  
  memos := make([]string, s.Raida.TotalServers())
	for i := 0; i < s.Raida.TotalServers(); i++ {
    memos[i] = vals[i][0] + config.MEMO_SEPARATOR + vals[i][1] + config.MEMO_SEPARATOR + vals[i][2]
  }

  return memos
}

func (s *Servant) AssembleMessage(ctx context.Context, parts [][]string) (string, error) {
  collected := make([]string, s.Raida.TotalServers())

  var cs int

  cs = -1

  logger.L(ctx).Debugf("p %v", parts)
  // Assembling
	for i := 0; i < s.Raida.TotalServers(); i++ {
    if parts[i] == nil {
      continue
    }

    logger.L(ctx).Debugf("%d pp %v", i, parts[i])

    if parts[i][0] == "" || parts[i][1] == "" || parts[i][2] == "" {
      continue
    }

    cidx0 := i
    cidx1 := i + 3
    cidx2 := i + 6

    if cidx1 >= s.Raida.TotalServers() {
      cidx1 -= s.Raida.TotalServers()
    }

    if cidx2 >= s.Raida.TotalServers() {
      cidx2 -= s.Raida.TotalServers()
    }

    collected[cidx0] = parts[i][0]
    collected[cidx1] = parts[i][1]
    collected[cidx2] = parts[i][2]

    if (len(collected[cidx0]) != len(collected[cidx1]) || 
      len(collected[cidx0]) != len(collected[cidx2])) {
        logger.L(ctx).Warnf("Chunk length mismatch %d %d %d idxs=%d,%d,%d", len(collected[cidx0]), len(collected[cidx1]), len(collected[cidx2]), cidx0, cidx1, cidx2)
      return "", perror.New(perror.ERROR_ASSEMBLE_STRIPE_MIRRORS, "Chunk length mismatch")
    }

    if cs == -1 {
      cs = len(collected[cidx0])
    } else {
      if len(collected[cidx0]) != cs {
        return "", perror.New(perror.ERROR_ASSEMBLE_STRIPE_MIRRORS, "Chunk length is different across some RAIDA Servers")
      }
    }
  }

  if cs == -1 {
      return "", perror.New(perror.ERROR_ASSEMBLE_STRIPE_MIRRORS, "Failed to get chunks size while assembling memo")
  }

  logger.L(ctx).Debugf("cs %d", cs)
  msg := make([]string, cs * s.Raida.TotalServers())
	for i := 0; i < s.Raida.TotalServers(); i++ {
    if collected[i] == "" {
      logger.L(ctx).Warnf("Failed to assemble message. Chunk #%d is missing", i)
      return "", perror.New(perror.ERROR_ASSEMBLE_STRIPE_MIRRORS, "Failed to assemble memo")
    }


    str := strings.Split(collected[i], "")
    for j := 0; j < len(str); j++ {
      offset := i + j * s.Raida.TotalServers()
      msg[offset] = str[j]
    }
  }

  for i := 0; i < len(msg); i++ {
    if msg[i] == "" {
      logger.L(ctx).Warnf("idx %d is missing, len=%d",i,len(msg))
      return "", perror.New(perror.ERROR_ASSEMBLE_STRIPE_MIRRORS, "Failed to assemble final memo")
    }
  }

  message := strings.Join(msg, "")
  message = strings.Trim(message, "-")

  logger.L(ctx).Debugf("message %s", message)

	return message, nil
}

func (s *Servant) SplitMessage(message string) [][]string {
	var data [][]string

	data = make([][]string, s.Raida.TotalServers())
	pads := len(message) % s.Raida.TotalServers()
	for i := 0; i < (s.Raida.TotalServers() - pads); i++ {
		message += "-"
	}

	for i := 0; i < s.Raida.TotalServers(); i++ {
		data[i] = make([]string, 3)
		data[i][0] = ""
		data[i][1] = ""
		data[i][2] = ""
	}

	cs := strings.Split(message, "")
	for i := 0; i < len(cs); i++ {
		ridx := i % s.Raida.TotalServers()
		data[ridx][0] += cs[i]
	}

	for i := 0; i < s.Raida.TotalServers(); i++ {
		cidx0 := i + 3
		cidx1 := i + 6

		if cidx0 >= s.Raida.TotalServers() {
			cidx0 -= s.Raida.TotalServers()
		}

		if cidx1 >= s.Raida.TotalServers() {
			cidx1 -= s.Raida.TotalServers()
		}

		data[i][1] += data[cidx0][0]
		data[i][2] += data[cidx1][0]
	}

	return data
}




/*

Sort Functions

*/

func PickTopKey(ctx context.Context, counters map[string]int) (string, error) {
  pairsString := sortByCountString(counters)
  var topKey string
  if len(pairsString) > 0 {
    topKey = pairsString[0].Key
    if counters[topKey] < config.MIN_QUORUM_COUNT {
      return "", perror.New(perror.ERROR_RAIDA_QUORUM, "QUORUM is not reached")
    }
    logger.L(ctx).Debugf("Top key chosen: %s", topKey)
  } else {
    return "", perror.New(perror.ERROR_RAIDA_QUORUM, "Failed to pick key")
  }

  return topKey, nil
}


func sortByCount(totals map[int]int) PairList {
	pl := make(PairList, len(totals))
	i := 0

	for k, v := range totals {
		pl[i] = Pair{k, v}
		i++
	}

	sort.Sort(sort.Reverse(pl))
	return pl
}

func sortByCountString(totals map[string]int) PairStringList {
	pl := make(PairStringList, len(totals))
	i := 0

	for k, v := range totals {
		pl[i] = PairString{k, v}
		i++
	}

	sort.Sort(sort.Reverse(pl))
	return pl
}

type PairString struct {
	Key   string
	Value int
}

type PairStringList []PairString

func (p PairStringList) Len() int           { return len(p) }
func (p PairStringList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairStringList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type Pair struct {
	Key   int
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
