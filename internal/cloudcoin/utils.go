package cloudcoin

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
)

var bin2status = map[int]int{
  BINARY_POWN_STATUS_UNTRIED: config.RAIDA_STATUS_UNTRIED,
  BINARY_POWN_STATUS_PASS: config.RAIDA_STATUS_PASS,
  BINARY_POWN_STATUS_NETWORK: config.RAIDA_STATUS_NORESPONSE,
  BINARY_POWN_STATUS_ERROR: config.RAIDA_STATUS_ERROR,
  BINARY_POWN_STATUS_FAILED: config.RAIDA_STATUS_FAIL,
  
}


func ValidateGuid(guid string) bool {
	rex, _ := regexp.Compile(`^[0-9a-fA-F]{32}$`)

	return rex.MatchString(guid)
}

func GetIPFromSn(sn uint32) string {
  cbytes := utils.ExplodeSn(sn)

  c0 := strconv.Itoa(int(cbytes[0]))
  c1 := strconv.Itoa(int(cbytes[1]))
  c2 := strconv.Itoa(int(cbytes[2]))

  ip := "1." + c0 + "." + c1 + "." + c2

  return ip
}

func GetSNFromIP(ipaddress string) (uint32, error) {
	ipRegex, _ := regexp.Compile(`^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$`)
	s := ipRegex.FindStringSubmatch(strings.Trim(ipaddress, " "))
	if len(s) == 5 {
		o2, err := strconv.Atoi(s[2])
		if err != nil {
			return 0, errors.New("Failed to convert IP octet2")
		}

		o3, err := strconv.Atoi(s[3])
		if err != nil {
			return 0, errors.New("Failed to convert IP octet3")
		}

		o4, err := strconv.Atoi(s[4])
		if err != nil {
			return 0, errors.New("Failed to convert IP octet4")
		}

		sn := (o2 << 16) | (o3 << 8) | o4

    if sn <= 0 || sn > config.TOTAL_COINS {
    	return 0, perror.New(perror.ERROR_INVALID_IP, "Incorrect SN from the IP address")
    }

		return uint32(sn), nil
	}

	return 0, perror.New(perror.ERROR_INVALID_IP, "Incorrect IP address")
}

func GetDenomination(sn uint32) int {

  // All are ones
  return 1

	if sn < 1 {
		return 0
	}

	if sn < 2097153 {
		return 1
	}

	if sn < 4194305 {
		return 5
	}

	if sn < 6291457 {
		return 25
	}

	if sn < 14680065 {
		return 100
	}

	if sn < 16777217 {
		return 250
	}

	return 0
}

func GetChangeMethod(denomination int) int {
	method := 0
	switch denomination {
	case 250:
		method = config.CHANGE_METHOD_250F
		break
	case 100:
		method = config.CHANGE_METHOD_100E
		break
	case 25:
		method = config.CHANGE_METHOD_25B
		break
	case 5:
		method = config.CHANGE_METHOD_5A
		break
	}
	return method
}

func CoinsGetA(a []uint32, cnt int) []uint32 {
	var sns []uint32
	var i, j int

	sns = make([]uint32, cnt)

	i = 0
	j = 0
	for ; i < len(a); i++ {
		if a[i] == 0 {
			continue
		}

		sns[j] = a[i]
		a[i] = 0
		j++

		if j == cnt {
			break
		}
	}

	if j != cnt {
		return nil
	}

	return sns
}

func CoinsGet25B(sb, ss []uint32) []uint32 {
	var sns, rsns []uint32

	rsns = make([]uint32, 9)
	sns = CoinsGetA(ss, 5)
	if sns == nil {
		return nil
	}

	for i := 0; i < 5; i++ {
		rsns[i] = sns[i]
	}

	sns = CoinsGetA(sb, 4)
	if sns == nil {
		return nil
	}

	for i := 0; i < 4; i++ {
		rsns[i+5] = sns[i]
	}

	return rsns
}

func CoinsGet100E(sb, ss, sss []uint32) []uint32 {
	var sns, rsns []uint32

	rsns = make([]uint32, 12)
	sns = CoinsGetA(sb, 3)
	if sns == nil {
		return nil
	}

	for i := 0; i < 3; i++ {
		rsns[i] = sns[i]
	}

	sns = CoinsGetA(ss, 4)
	if sns == nil {
		return nil
	}

	for i := 0; i < 4; i++ {
		rsns[i+3] = sns[i]
	}

	sns = CoinsGetA(sss, 5)
	if sns == nil {
		return nil
	}

	for i := 0; i < 5; i++ {
		rsns[i+7] = sns[i]
	}

	return rsns
}

func CoinsGet250F(sb, ss, sss, ssss []uint32) []uint32 {
	var sns, rsns []uint32

	rsns = make([]uint32, 15)
	sns = CoinsGetA(sb, 1)
	if sns == nil {
		return nil
	}

	rsns[0] = sns[0]

	sns = CoinsGetA(ss, 5)
	if sns == nil {
		return nil
	}

	for i := 0; i < 5; i++ {
		rsns[i+1] = sns[i]
	}

	sns = CoinsGetA(sss, 4)
	if sns == nil {
		return nil
	}

	for i := 0; i < 4; i++ {
		rsns[i+6] = sns[i]
	}

	sns = CoinsGetA(ssss, 5)
	if sns == nil {
		return nil
	}

	for i := 0; i < 5; i++ {
		rsns[i+10] = sns[i]
	}

	return rsns
}

func GeneratePan() (string, error) {
	return utils.GenerateHex(16)
}

func GetStatusFromBinary(status int) int {
  status, ok := bin2status[status]
  if !ok {
    return config.RAIDA_STATUS_ERROR
  }

  return status
}

func GetBinaryFromStatus(status int) int {
  for k, v := range(bin2status) {
    if v == status {
      return k
    }
  }

  return BINARY_POWN_STATUS_ERROR
}

func GetANFromPG(rIdx int, sn uint32, pg string) string {
  h := strconv.Itoa(rIdx) + strconv.Itoa(int(sn)) + pg
  an := utils.GetMD5(h)

  return an
}

