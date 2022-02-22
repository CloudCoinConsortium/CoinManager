package cloudcoin

import (

	//	"os"
	//	"io/ioutil"
	"context"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
)

const (
  BINARY_POWN_STATUS_UNTRIED = 0x0
  BINARY_POWN_STATUS_PASS = 0xA
  BINARY_POWN_STATUS_NETWORK = 0xD
  BINARY_POWN_STATUS_ERROR = 0xE
  BINARY_POWN_STATUS_FAILED = 0xF
)

type CloudCoin struct {
	Sn       uint32 `json:"sn"`
	Ans      []string    `json:"ans"`
	Pans     []string    `json:"-"`
	Statuses []int       `json:"-"`
  PownString string    `json:"pownstring,omitempty"`

  formatType uint8
  cloudId uint8
  coinID1 uint8
  coinID2 uint8
  splitNumber uint8
  flags uint8
  encryptionType uint8
  receiptId string
  passwordHash string

  passed, failed int
  gradeStatus int
  locationStatus int

  skyName string
}

type CloudCoinStack struct {
	Stack []CloudCoin `json:"cloudcoin"`
}

func NewFromData(sn uint32) *CloudCoin {
	var cc CloudCoin

  cc.Sn = sn
	cc.Ans = make([]string, config.TOTAL_RAIDA_NUMBER)
	cc.Pans = make([]string, config.TOTAL_RAIDA_NUMBER)
  for i := 0; i < len(cc.Ans); i++ {
    cc.Ans[i] = "00000000000000000000000000000000" 
    cc.Pans[i] = "00000000000000000000000000000000" 
  }

  cc.formatType = config.COIN_FORMAT_TYPE_STANDARD
  cc.cloudId = 0
  cc.coinID1 = config.COIN_ID1_MAIN
  cc.coinID2 = config.COIN_ID2_CLOUDCOIN
  cc.flags = 0
  cc.encryptionType = config.COIN_ENCRYPTION_TYPE_NONE
  cc.receiptId = "00000000000000000000000000000000"
  cc.passwordHash = "000000000000000000"
  cc.splitNumber = 0

	cc.Statuses = make([]int, config.TOTAL_RAIDA_NUMBER)
	for idx := 0; idx < config.TOTAL_RAIDA_NUMBER; idx++ {
		cc.Statuses[idx] = config.RAIDA_STATUS_UNTRIED
	}

  cc.gradeStatus = config.COIN_STATUS_UNKNOWN
  cc.locationStatus = config.COIN_LOCATION_STATUS_UNKNOWN

	return &cc
}

func NewFromBinarySingle(ctx context.Context, alldata[] byte) (*CloudCoin, error) {
  if len(alldata) <= config.CC_BINARY_HEADER_SIZE + 1 {
    return nil, perror.New(perror.ERROR_INVALID_HEADER_SIZE, "Wrong size of the body/header")
  }
  return NewFromBinary(ctx, alldata[:config.CC_BINARY_HEADER_SIZE], alldata[config.CC_BINARY_HEADER_SIZE:])
}

func NewFromBinary(ctx context.Context, header []byte, data[] byte) (*CloudCoin, error) {
	var cc CloudCoin

  if (len(header) != config.CC_BINARY_HEADER_SIZE) {
    return nil, perror.New(perror.ERROR_INVALID_HEADER_SIZE, "Header is not " + strconv.Itoa(config.CC_BINARY_HEADER_SIZE) + " bytes")
  }

  withPans := false
  if (len(data) != (config.TOTAL_RAIDA_NUMBER * 16) + 16) {
    if (len(data) != (config.TOTAL_RAIDA_NUMBER * 16 * 2) + 16) {
     return nil, perror.New(perror.ERROR_INVALID_COIN_BODY_SIZE, "Coin body must be 416 or 816 bytes")
    }

    withPans = true
  }

  logger.L(ctx).Debugf("Got body size %d (header size %d) bytes with pans %v", len(data), len(header), withPans)

  ftype := uint8(header[0])
  //if ftype != config.COIN_FORMAT_TYPE_STANDARD && ftype != config.COIN_FORMAT_TYPE_PANG && ftype != config.COIN_FORMAT_TYPE_STORE_IN_MIND {
  if ftype != config.COIN_FORMAT_TYPE_STANDARD {
    return nil, perror.New(perror.ERROR_UNSUPPORTED_COIN_FORMAT_TYPE, "Unsupported format type")
  }

  cloudId := uint8(header[1])
  if cloudId != 0 {
    return nil, perror.New(perror.ERROR_UNSUPPORTED_COIN_CLOUDID, "Unsupported CloudID")
  }

  coinId1 := uint8(header[2])
  if coinId1 != config.COIN_ID1_MAIN {
    return nil, perror.New(perror.ERROR_UNSUPPORTED_COIN_ID1, "Unsupported CoinId1")
  }

  coinId2 := uint8(header[3])
  if coinId2 != config.COIN_ID2_CLOUDCOIN && coinId2 != config.COIN_ID2_ID {
    return nil, perror.New(perror.ERROR_UNSUPPORTED_COIN_ID1, "Unsupported CoinId2")
  }

  splitId := uint8(header[4])

  encryptionType := uint8(header[5])
  if encryptionType != config.COIN_ENCRYPTION_TYPE_NONE {
    return nil, perror.New(perror.ERROR_UNSUPPORTED_ENCRYPTION_TYPE, "Unsupported Encryption Type")
  }

  passwordHash := hex.EncodeToString(header[6:15])
  flags := uint8(header[15])
  receiptId := hex.EncodeToString(header[16:])

  var sn uint32

  sn = (uint32(data[0]) << 16) | (uint32(data[1]) << 8) | uint32(data[2])
  if (sn == 0 || sn > config.TOTAL_COINS) {
    return nil, perror.New(perror.ERROR_INVALID_SERIAL_NUMBER, "Invalid serial Number")
  }

  cc.Sn = sn
	cc.Ans = make([]string, config.TOTAL_RAIDA_NUMBER)
	cc.Pans = make([]string, config.TOTAL_RAIDA_NUMBER)
	cc.Statuses = make([]int, config.TOTAL_RAIDA_NUMBER)
	for idx := 0; idx < config.TOTAL_RAIDA_NUMBER; idx++ {
		cc.Statuses[idx] = config.RAIDA_STATUS_UNTRIED
	}

  // Set Pown Statues
  for v := 3; v < 16; v++ {
    ps := uint32(data[v])

    rfIdx := (v - 3) * 2
    rsIdx := rfIdx + 1

    rfStatus := ps >> 4
    rsStatus := ps & 0xF

    lfStatus := GetStatusFromBinary(int(rfStatus))
    lsStatus := GetStatusFromBinary(int(rsStatus))

    if rfIdx < config.TOTAL_RAIDA_NUMBER {
      cc.SetDetectStatus(rfIdx, int(lfStatus))
    }

    if rsIdx < config.TOTAL_RAIDA_NUMBER {
      cc.SetDetectStatus(rsIdx, int(lsStatus))
    }
  }

  for i := 0; i < config.TOTAL_RAIDA_NUMBER; i++ {
    start := 16 + i * 16
    end := start + 16
    an := data[start:end]
    cc.Ans[i] = hex.EncodeToString(an)

    if withPans {
      pstart := 416 + i * 16
      pend := pstart + 16
      pan := data[pstart:pend]
      cc.Pans[i] = hex.EncodeToString(pan)
    }
  }

  cc.cloudId = cloudId
  cc.coinID1 = coinId1
  cc.coinID2 = coinId2
  cc.encryptionType = encryptionType
  cc.flags = flags
  cc.formatType = ftype
  cc.passwordHash = passwordHash
  cc.receiptId = receiptId
  cc.splitNumber = splitId


  cc.SetPownStringFromDetectStatus()
  cc.Grade()

  logger.L(ctx).Debugf("Got coin %d: %s, %s", cc.Sn, cc.GetPownString(), cc.GetGradeStatusString())

  //cc.gradeStatus = config.COIN_STATUS_UNKNOWN
  cc.locationStatus = config.COIN_LOCATION_STATUS_UNKNOWN

	return &cc, nil
}

func (cc *CloudCoin) SetAns(ans []string) {
  cc.Ans = ans
}

func (cc *CloudCoin) SetSkyName(name string) {
  cc.skyName = name
  cc.coinID2 = config.COIN_ID2_ID
}

func (cc *CloudCoin) GetSkyName() string {
  return cc.skyName
}

func (cc *CloudCoin) SetCoinID(id uint16) {
  cc.coinID1 = uint8(id >> 8)
  cc.coinID2 = uint8(id & 0xff)
}

func (cc *CloudCoin) GetCoinID() uint16 {
  return uint16((cc.coinID1 << 8) | cc.coinID2)
}

func (cc *CloudCoin) IsIDCoin() bool {
  return cc.coinID2 == config.COIN_ID2_ID
}

func (cc *CloudCoin) GetDenomination() int {
	return GetDenomination(cc.Sn)
}

func (cc *CloudCoin) SetStatusesFromPownString() error {
  if (cc.PownString == "") {
    return perror.New(perror.ERROR_INVALID_POWNSTRING, "Invalid Pownstring")
  }

  if (len(cc.PownString) != config.TOTAL_RAIDA_NUMBER) {
    return perror.New(perror.ERROR_INVALID_POWNSTRING, "Invalid Pownstring")
  }

  for idx, c := range(cc.PownString) {
    status := config.RAIDA_STATUS_UNTRIED
    switch c {
    case 'p':
      status = config.RAIDA_STATUS_PASS
    case 'e':
      status = config.RAIDA_STATUS_ERROR
    case 'u':
      status = config.RAIDA_STATUS_UNTRIED
    case 'n':
      status = config.RAIDA_STATUS_NORESPONSE
    case 'f':
      status = config.RAIDA_STATUS_FAIL
    }

    cc.SetDetectStatus(idx, status)
  }

  return nil
}

func (cc *CloudCoin) GetName() string {
	s := fmt.Sprintf("%d.CloudCoin.%d.%d" + config.CC_FILE_BINARY_EXTENSION, cc.GetDenomination(), cc.coinID2, cc.Sn)

	return s
}

func (cc *CloudCoin) GetIntName() string {
	s := fmt.Sprintf("%d.CloudCoin.%d.%d.%s" + config.CC_FILE_BINARY_EXTENSION, cc.GetDenomination(), cc.coinID2, cc.Sn, cc.GetPownString())

	return s
}

func (cc *CloudCoin) SetLocationStatus(status int) {
  cc.locationStatus = status
}

func (cc *CloudCoin) GetLocationStatus() int {
  return cc.locationStatus
}

func (cc *CloudCoin) SetDetectStatus(idx int, status int) {
	cc.Statuses[idx] = status
}

func (cc *CloudCoin) Grade() {
  cc.passed = 0
  cc.failed = 0
  for _, status := range(cc.Statuses) {
    if status == config.RAIDA_STATUS_PASS {
      cc.passed++
    } else if status == config.RAIDA_STATUS_FAIL {
      cc.failed++
    }
  }

  isAuthentic := cc.passed >= config.MIN_PASSED_NUM_TO_BE_AUTHENTIC
  isCounterfeit := cc.failed >= config.MAX_FAILED_NUM_TO_BE_COUNTERFEIT

  if (isAuthentic) {
    if (cc.failed > 0) {
      cc.gradeStatus = config.COIN_STATUS_FRACKED
    } else {
      cc.gradeStatus = config.COIN_STATUS_AUTHENTIC
    }
  } else {
    if (isCounterfeit) {
      cc.gradeStatus = config.COIN_STATUS_COUNTERFEIT
    } else {
      cc.gradeStatus = config.COIN_STATUS_LIMBO
    }
  }
}

func (cc *CloudCoin) GetGradeStatus() int {
  return cc.gradeStatus
}

func (cc *CloudCoin) GetGradeStatusString() string {
  switch (cc.gradeStatus) {
  case config.COIN_STATUS_UNKNOWN:
    return "Unknown"
  case config.COIN_STATUS_AUTHENTIC:
    return "Authentic"
  case config.COIN_STATUS_FRACKED:
    return "Fracked"
  case config.COIN_STATUS_COUNTERFEIT:
    return "Counterfeit"
  case config.COIN_STATUS_LIMBO:
    return "Limbo"
  default:
    return "?"
  }
}

func (cc *CloudCoin) IsAuthentic() (bool, bool, bool) {
  passed := 0
  failed := 0
  for _, status := range(cc.Statuses) {
    if status == config.RAIDA_STATUS_PASS {
      passed++
    } else if status == config.RAIDA_STATUS_FAIL {
      failed++
    }
  }

  isAuthentic := passed >= config.MIN_PASSED_NUM_TO_BE_AUTHENTIC
  hasFailed := failed > 0
  isCounterfeit := failed >= config.MAX_FAILED_NUM_TO_BE_COUNTERFEIT

  return isAuthentic, hasFailed, isCounterfeit
}

func (cc *CloudCoin) GetPownString() string {
	pownString := ""
	for idx, _ := range cc.Statuses {
		switch cc.Statuses[idx] {
		case config.RAIDA_STATUS_UNTRIED:
			pownString += "u"
		case config.RAIDA_STATUS_FAIL:
			pownString += "f"
		case config.RAIDA_STATUS_PASS:
			pownString += "p"
		case config.RAIDA_STATUS_ERROR:
			pownString += "e"
		case config.RAIDA_STATUS_NORESPONSE:
			pownString += "n"
    default:
      pownString += "?"
		}
	}

	return pownString
}

func (cc *CloudCoin) SetPownStringFromDetectStatus() {
  cc.PownString = cc.GetPownString()
}

func (cc *CloudCoin) SetPownString(pownstring string) error {
  cc.PownString = pownstring

  return cc.SetStatusesFromPownString()
}

func (cc *CloudCoin) SetAn(idx int, an string) {
	cc.Ans[idx] = an
}

func (cc *CloudCoin) SetPan(idx int, pan string) {
	cc.Pans[idx] = pan
}

func (cc *CloudCoin) GenerateMyPans() {
	for idx := 0; idx < config.TOTAL_RAIDA_NUMBER; idx++ {
		cc.Pans[idx], _ = GeneratePan()
	}
}

func (cc *CloudCoin) SetAnsToPansIfPassed() {
	for idx := 0; idx < config.TOTAL_RAIDA_NUMBER; idx++ {
    if (cc.Statuses[idx] != config.RAIDA_STATUS_PASS) {
      continue
    }

		cc.Ans[idx] = cc.Pans[idx]
	}
}


func (cc *CloudCoin) GetHeader() ([]byte, error) {
  storage := make([]byte, 32)

  storage[0] = byte(cc.formatType)
  storage[1] = byte(cc.cloudId)
  storage[2] = byte(cc.coinID1)
  storage[3] = byte(cc.coinID2)
  storage[4] = byte(cc.encryptionType)
  storage[5] = byte(cc.splitNumber)

  passwordHash, err := hex.DecodeString(cc.passwordHash)
  if err != nil {
    return nil, err
  }

  copy(storage[6:], passwordHash)

  storage[15] = 0
  
  receiptId, err := hex.DecodeString(cc.receiptId)
  if err != nil {
    return nil, err
  }

  copy(storage[16:], receiptId)

  return storage, nil
}

func (cc *CloudCoin) GetData() ([]byte, error) {
  storage, err := cc.GetHeader()
  if err != nil {
    return nil, err
  }

  contents, err := cc.GetContentData()
  if err != nil {
    return nil, err
  }

  storage = append(storage, contents...)

  return storage, nil
}

func (cc *CloudCoin) GetContentData() ([]byte, error) {
  storage := make([]byte, 0)

  b0 := byte((cc.Sn >> 16) & 0xff)
  b1 := byte((cc.Sn >> 8) & 0xff)
  b2 := byte(cc.Sn & 0xff)

  storage = append(storage, b0, b1, b2)

  // PownStatuses
  data := cc.EncodeBinaryPowns()

  storage = append(storage, data...)

  for idx := 0; idx < config.TOTAL_RAIDA_NUMBER; idx++ {
    data, err := hex.DecodeString(cc.Ans[idx])
    if err != nil {
      return nil, err
    }

    storage = append(storage, data...)
  }

  return storage, nil
}

func (cc *CloudCoin) GetContentDataWithPans() ([]byte, error) {
  storage := make([]byte, 0)

  b0 := byte((cc.Sn >> 16) & 0xff)
  b1 := byte((cc.Sn >> 8) & 0xff)
  b2 := byte(cc.Sn & 0xff)

  storage = append(storage, b0, b1, b2)

  // PownStatuses
  data := cc.EncodeBinaryPowns()

  storage = append(storage, data...)

  for idx := 0; idx < config.TOTAL_RAIDA_NUMBER; idx++ {
    data, err := hex.DecodeString(cc.Ans[idx])
    if err != nil {
      return nil, err
    }

    storage = append(storage, data...)
  }

  for idx := 0; idx < config.TOTAL_RAIDA_NUMBER; idx++ {
    data, err := hex.DecodeString(cc.Pans[idx])
    if err != nil {
      return nil, err
    }

    storage = append(storage, data...)
  }

  return storage, nil
}

func (cc *CloudCoin) GetDataWithPans() ([]byte, error) {

  storage, err := cc.GetHeader()
  if err != nil {
    return nil, err
  }

  contents, err := cc.GetContentDataWithPans()
  if err != nil {
    return nil, err
  }

  storage = append(storage, contents...)

  return storage, nil
}

func (cc *CloudCoin) EncodeBinaryPowns() []byte {
  bs := make([]byte, 13)

  for i := 0; i < len(bs); i++ {
    rfIdx := i * 2
    rsIdx := rfIdx + 1
    sf := GetBinaryFromStatus(cc.Statuses[rfIdx])

    var ss int
    if rsIdx < config.TOTAL_RAIDA_NUMBER {
      ss = GetBinaryFromStatus(cc.Statuses[rsIdx])
    } else {
      ss = 0
    }

    bs[i] = byte((sf << 4) | ss)
  }

  return bs
}

