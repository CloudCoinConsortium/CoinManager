package utils

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"math"
	mrand "math/rand"
	"os"
	"strconv"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
)

// how many digit's groups to process
const groupsNumber int = 4

var _smallNumbers = []string{
	"zero", "one", "two", "three", "four",
	"five", "six", "seven", "eight", "nine",
	"ten", "eleven", "twelve", "thirteen", "fourteen",
	"fifteen", "sixteen", "seventeen", "eighteen", "nineteen",
}
var _tens = []string{
	"", "", "twenty", "thirty", "forty", "fifty",
	"sixty", "seventy", "eighty", "ninety",
}
var _scaleNumbers = []string{
	"", "thousand", "million", "billion",
}

type digitGroup int

// Convert converts number into the words representation.
func Convert(number int) string {
	return convert(number, false)
}

// ConvertAnd converts number into the words representation
// with " and " added between number groups.
func ConvertAnd(number int) string {
	return convert(number, true)
}

func convert(number int, useAnd bool) string {
	// Zero rule
	if number == 0 {
		return _smallNumbers[0]
	}

	// Divide into three-digits group
	var groups [groupsNumber]digitGroup
	positive := math.Abs(float64(number))

	// Form three-digit groups
	for i := 0; i < groupsNumber; i++ {
		groups[i] = digitGroup(math.Mod(positive, 1000))
		positive /= 1000
	}

	var textGroup [groupsNumber]string
	for i := 0; i < groupsNumber; i++ {
		textGroup[i] = digitGroup2Text(groups[i], useAnd)
	}
	combined := textGroup[0]
	and := useAnd && (groups[0] > 0 && groups[0] < 100)

	for i := 1; i < groupsNumber; i++ {
		if groups[i] != 0 {
			prefix := textGroup[i] + " " + _scaleNumbers[i]

			if len(combined) != 0 {
				prefix += separator(and)
			}

			and = false

			combined = prefix + combined
		}
	}

	if number < 0 {
		combined = "minus " + combined
	}

	return combined
}

func intMod(x, y int) int {
	return int(math.Mod(float64(x), float64(y)))
}

func digitGroup2Text(group digitGroup, useAnd bool) (ret string) {
	hundreds := group / 100
	tensUnits := intMod(int(group), 100)

	if hundreds != 0 {
		ret += _smallNumbers[hundreds] + " hundred"

		if tensUnits != 0 {
			ret += separator(useAnd)
		}
	}

	tens := tensUnits / 10
	units := intMod(tensUnits, 10)

	if tens >= 2 {
		ret += _tens[tens]

		if units != 0 {
			ret += "-" + _smallNumbers[units]
		}
	} else if tensUnits != 0 {
		ret += _smallNumbers[tensUnits]
	}

	return
}

// separator returns proper separator string between
// number groups.
func separator(useAnd bool) string {
	if useAnd {
		return " and "
	}
	return " "
}

func CalcCrc32(data []byte) uint32 {
	crc32q := crc32.MakeTable(0xedb88320)

	return crc32.Checksum([]byte(data), crc32q)
}

func BasicPNGChecks(ctx context.Context, bytes []byte) int {
  if len(bytes) < 16 {
    logger.L(ctx).Debugf("Too small length for PNG header")
    return -1
  }

	if bytes[0] != 0x89 && bytes[1] != 0x50 && bytes[2] != 0x4e && bytes[3] != 0x45 && bytes[4] != 0x0d && bytes[5] != 0x0a && bytes[6] != 0x1a && bytes[7] != 0x0a {
		logger.L(ctx).Debugf("Invalid header")
		return -1
	}

	chunkLength := binary.BigEndian.Uint32(bytes[8:])
	headerSig := binary.BigEndian.Uint32(bytes[12:])
	if headerSig != 0x49484452 {
		logger.L(ctx).Debugf("Invalid signature")
		return -1
	}

	idx := int(16 + chunkLength)
	crcOffset := 12 + int(4+chunkLength)
	crcSig := binary.BigEndian.Uint32(bytes[idx:])
	calcCrc := CalcCrc32(bytes[12:crcOffset])
	if crcSig != calcCrc {
		logger.L(ctx).Debugf("Invalid PNG Crc32 checksum")
		return -1
	}

	return idx
}

func GetHash(v string) string {
  if v == "" {
    return ""
  }

  bs := []byte(v)

  h := sha1.New()
  h.Write(bs)

  sum := h.Sum(nil)

  return hex.EncodeToString(sum)
}

func GenerateReceiptID() (string, error) {
  return GenerateHex(16)
}

func GeneratePG() (string, error) {
  return GenerateHex(16)
}

func GenerateHex(length int) (string, error) {
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
    return "", perror.New(perror.ERROR_GENERATE_RANDOM, "Failed to generate random string " + err.Error())
	}

	return hex.EncodeToString(bytes), nil
}

func GenerateStrNumber(length int) string {
  seedObj := mrand.NewSource(time.Now().UnixNano())
  randObj := mrand.New(seedObj)

  s := ""
  for i := 0; i < length; i++ {
    s += strconv.Itoa(randObj.Intn(10))
  }

  return s
}

func ReverseString(s string) string {
  runes := []rune(s)
  for i, j := 0, len(runes) - 1; i < j; i, j = i + 1, j - 1 {
    runes[i], runes[j] = runes[j], runes[i]
  }

  return string(runes)
}


func GetUint24(v []byte) uint32 {
  b0 := uint32(v[0])
  b1 := uint32(v[1])
  b2 := uint32(v[2])

  return (b0 << 16) | (b1 << 8) | b2
}

func ExplodeSn(sn uint32) []byte {
  cbyte := make([]byte, 3)
  cbyte[0] = byte((sn >> 16) & 0xff)
  cbyte[1] = byte((sn >> 8) & 0xff)
  cbyte[2] = byte(sn & 0xff)

  return cbyte
}

func GetCurrentTsBytes() []byte {
  t := time.Now()

  year, month, day := t.UTC().Date()
  hour, min, sec := t.UTC().Clock()
  
  bts := make([]byte, 6)
  bts[0] = byte(year - config.DEFAULT_YEAR)
  bts[1] = byte(month)
  bts[2] = byte(day)
  bts[3] = byte(hour)
  bts[4] = byte(min)
  bts[5] = byte(sec)

  /*
  ts := uint32(time.Now().Unix())

  bts := make([]byte, 6)
  bts[0] = 0x0
  bts[1] = 0x0

  binary.BigEndian.PutUint32(bts[2:], ts)
  */

  return bts
}

func GetTimeFromBytes(bytes []byte) time.Time {
  if len(bytes) != 6 {
    return time.Time{}
  }

  year := int(bytes[0]) + config.DEFAULT_YEAR
  month := time.Month(bytes[1])
  day := int(bytes[2])
  hour := int(bytes[3])
  min := int(bytes[4])
  sec := int(bytes[5])

  t := time.Date(year, month, day, hour, min, sec, 0, time.UTC)

  return t
}

func GetTotalIterations(coinCount int, strideSize int) int {
  batches := int(coinCount / strideSize)

  if int(coinCount % strideSize) != 0 {
    batches++
  }

  return batches * config.TOTAL_RAIDA_NUMBER
}

func CalcTotalIterations(coinCount int) int {
  batches := int(coinCount / config.MAX_NOTES_TO_SEND)

  if int(coinCount % config.MAX_NOTES_TO_SEND) != 0 {
    batches++
  }

  return batches * config.TOTAL_RAIDA_NUMBER
}


func GetMD5(s string) string {
  data := []byte(s)

  v := md5.Sum(data)
  sv := fmt.Sprintf("%x", v)

  return sv
}

func GetMD5File(file *os.File) (string, error) {
  hash := md5.New()

  _, err := io.Copy(hash, file)
  if err != nil {
    return "", err
  }

  sv := fmt.Sprintf("%x", hash.Sum(nil))

  return sv, nil
}

func GetUDPPacketCount(length int) int {
  packets := int(length / config.MAX_RAIDA_DATAGRAM_SIZE)
  if (length % config.MAX_RAIDA_DATAGRAM_SIZE) != 0 {
    packets++
  }
  return packets
}

