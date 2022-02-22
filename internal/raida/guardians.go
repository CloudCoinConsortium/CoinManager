package raida

import (
	"bufio"
	"crypto/tls"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)



type Guardian struct {
  Name string
  Success bool
  PrimaryRaidas []string
  BackupRaidas []string
  Hash string
}

var RaidaList ServerList

var Guardians []Guardian

func NewGuardian(name string) *Guardian {
  primaryRaidas := make([]string, config.TOTAL_RAIDA_NUMBER)
  backupRaidas := make([]string, config.TOTAL_RAIDA_NUMBER)

  return &Guardian{
    Name: name,
    Success: false,
    PrimaryRaidas: primaryRaidas,
    BackupRaidas: backupRaidas,
  }
}

func InitRAIDAList() {
  RaidaList.PrimaryRaidaList = config.LocalRaidas
  RaidaList.BackupRaidaList = config.LocalRaidas
  RaidaList.ActiveRaidaList = make([]*string, config.TOTAL_RAIDA_NUMBER)
  RaidaList.Mutex = &sync.Mutex{}
}

func InitGuardians() error {
  Guardians = make([]Guardian, 0)

  var c chan string = make(chan string)
  for _, gname := range(config.Guardians) {
    guardian := NewGuardian(gname)

    go InitGuardian(guardian, c)
    Guardians = append(Guardians, *guardian)
  }

  oks := 0
  for i := 0; i < len(config.Guardians); i++ {
    name := <- c
    if name == "" {
      continue
    }

    for idx, _ := range(Guardians) {
      if Guardians[idx].Name == name {
        Guardians[idx].MarkSuccess()
        oks++
        logger.L(nil).Debugf("Guardian %s Added", name)
        break
      }
    }
  }

  if (oks < 3) {
    logger.L(nil).Errorf("Too few Guardians initialized. It is not enough to continue")
    return perror.New(perror.ERROR_GUARDIANS, "Too few Guardians initialized")
  }


  hashes := make(map[string]int, 0)
  for _, guardian := range(Guardians) {
    if !guardian.IsSuccess() {
      continue
    }

    hashes[guardian.Hash]++
    if (hashes[guardian.Hash] >= 3) {
      logger.L(nil).Debugf("At least three valid Guardians Found")
      logger.L(nil).Debugf("Initialized %d Guardians", len(Guardians))

      RaidaList.PrimaryRaidaList = guardian.PrimaryRaidas
      RaidaList.BackupRaidaList = guardian.BackupRaidas
      RaidaList.ActiveRaidaList = make([]*string, config.TOTAL_RAIDA_NUMBER)
      RaidaList.Mutex = &sync.Mutex{}
      return nil
    }
  }

  return perror.New(perror.ERROR_GUARDIANS, "Failed to find at least three same config host.txt files on the Guardians")
}

func InitGuardian(g *Guardian, c chan string) {
  err := g.Init()
  if err != nil {
    c <- ""
    return
  }

  c <- g.Name
}

func (v *Guardian) IsSuccess() bool {
  return v.Success
}

func (v *Guardian) MarkSuccess() {
  v.Success = true
}

func (v *Guardian) Init() error {
  logger.L(nil).Debugf("Initializing Guardian %s", v.Name)


  body, err := v.ReadURL("/host.txt")
  if err != nil {
    return err
  }

  err = v.SetRaidas(body)
  if err != nil {
    logger.L(nil).Debugf("Failed to parse body for %s: %s", v.Name, err.Error())
    return perror.New(perror.ERROR_GUARDIAN_PARSE, "Faield to parse guardian config: " + v.Name)
  }

  return nil
}

func (v *Guardian) ReadURL(url string) (string, error) {
  tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
  }

	client := &http.Client{
    Transport: tr,
    Timeout: config.GUARDIAN_HTTP_TIMEOUT * time.Second,
  }

  hostsURL := "https://" + v.Name + url
  logger.L(nil).Debugf("Requesting %s", hostsURL)
  resp, err := client.Get(hostsURL)
  if err != nil {
    logger.L(nil).Debugf("Request to url %s failed: %s", hostsURL, err.Error())
    return "", perror.New(perror.ERROR_HTTP_REQUEST, "Request to Guardian failed " + hostsURL + ": " + err.Error())
  }

	statusCode := resp.StatusCode
  if statusCode != 200 {
    logger.L(nil).Debugf("Invalid status %d for %s", statusCode, v.Name)
    return "", perror.New(perror.ERROR_HTTP_REQUEST, "Invalid Status Code for " + v.Name + " : " + strconv.Itoa(statusCode))
  }
 
  defer resp.Body.Close()

  body, err := io.ReadAll(resp.Body)
  if err != nil {
    logger.L(nil).Debugf("Failed to read body for %s:%s", v.Name, err.Error())
    return "", perror.New(perror.ERROR_HTTP_REQUEST, "Faield to read body for " + v.Name + ": " + err.Error())
  }

  return string(body), nil
}


func (v *Guardian) GetNews() error {
  logger.L(nil).Debugf("Getting news for Guardian %s", v.Name)

  return nil
}

func (v *Guardian) SetRaidas(body string) error {
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
    line := strings.TrimSpace(scanner.Text())

    parts := strings.Split(line, " ")
    if len(parts) != 2 {
      continue
    }

    urlItem := strings.Split(parts[0], ":")
    if len(urlItem) != 2 {
      continue
    }

    ip := urlItem[0]
    port, err := strconv.Atoi(urlItem[1])
    if err != nil {
      continue
    }

    err = validation.Validate(ip, validation.Required, is.IP)
    if err != nil {
      continue
    }

    err = validation.Validate(urlItem[1], validation.Required, is.Port)
    if err != nil {
      continue
    }

    trailer := parts[1]
    if len(trailer) < 2 {
      continue
    }

    s := string(trailer[0])
    rnum, err := strconv.Atoi(trailer[1:])
    if err != nil {
      continue
    }

    if rnum < 0 || rnum >= config.TOTAL_RAIDA_NUMBER {
      continue
    }

    if s != "m" && s != "r" {
      continue
    }

    if s == "r" {
      v.PrimaryRaidas[rnum] = ip + ":" + strconv.Itoa(port)
    } else if s == "m" {
      v.BackupRaidas[rnum] = ip + ":" + strconv.Itoa(port)
    }
	}

  hobj := ""
  for idx, item := range(v.PrimaryRaidas) {
    if item == "" {
      logger.L(nil).Debugf("Primary raida %d not found", idx)
      return perror.New(perror.ERROR_GUARDIAN_PARSE, "Primary Raida " + strconv.Itoa(idx) + " not found")
    }

    hobj += item
  }

  for _, item := range(v.BackupRaidas) {
    hobj += item
  }

  hash := utils.GetHash(hobj)
  v.Hash = hash

  return nil
}
