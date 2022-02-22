package dnsservice

import (
	"context"
	"encoding/json"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/httpclient"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
)

type DNSResponse struct {
  Status string `json:"status"`
  Message string `json:"message"`
}

type DNSService struct {
  progressChannel chan interface{}
}


func New(progressChannel chan interface{}) *DNSService {
  return &DNSService{
    progressChannel: progressChannel,
  }
}

func (v *DNSService) GetSN(ctx context.Context, name string) (uint32, error) {
  ips, err := net.LookupIP(name)
  if err != nil {
    switch err.(type) {
    case *net.DNSError:
      derr := err.(*net.DNSError)
      if derr.IsNotFound {
        logger.L(ctx).Debugf("host %s not found", name)
        return 0, nil
      }
    }
    logger.L(ctx).Debugf("Failed to resolve %s: %s", name, err.Error())
    return 0, err
  }
  
  if len(ips) == 0 {
    logger.L(ctx).Debugf("No IP set for %s", name)
    return 0, nil
  }

  if len(ips) > 1 {
    logger.L(ctx).Warnf("Too many IPs for %s: %v. We can't continue", name, ips)
    return 0, perror.New(perror.ERROR_TOO_MANY_IPS, "Too many IP addresses for this skywallet")
  }

  ip := ips[0]

  logger.L(ctx).Debugf("Name %s resolved to %s", name, ip)

  sn, err := cloudcoin.GetSNFromIP(ip.String())
  if err != nil {
    return 0, err
  }

  return sn, nil
}

func (v *DNSService) RegisterName(ctx context.Context, cc *cloudcoin.CloudCoin) error {
  name := cc.GetSkyName()
  if name == "" {
    return perror.New(perror.ERROR_NO_SKYWALLET_NAME, "SkyWallet Name is not defined")
  }

  ip := cloudcoin.GetIPFromSn(cc.Sn)
  logger.L(ctx).Debugf("Registering Name %s for sn %d ip %s", name, cc.Sn, ip)

  existingSn, err := v.GetSN(ctx, name)
  if err != nil {
    return err
  }

  if existingSn != 0 {
    logger.L(ctx).Debugf("DNS name already exists for %s: %d", name, existingSn)
    return perror.New(perror.ERROR_DNS_RECORD_EXISTS, "This name already has DNS name")
  }


  rr, err := v.GetRandomRaida(ctx)
  if err != nil {
    logger.L(ctx).Errorf("Failed to pick random raida")
    return err
  }

  logger.L(ctx).Debugf("Requesting ticket from raida %d", rr)

  t := raida.NewGetTicket(v.progressChannel)
  response, err := t.GetSingleTicketSky(ctx, cc, rr)
  if err != nil {
    logger.L(ctx).Errorf("Failed to get ticket")
    return perror.New(perror.ERROR_GET_TICKET, "Failed to get ticket")
  }

  c := httpclient.New(config.DNSSERVICE_URL, 0)

  params := make(map[string]string, 0)
  params["sn"] = strconv.Itoa(int(cc.Sn))
  params["ticket"] = response
  params["username"] = name
  params["raidanumber"] = strconv.Itoa(rr)

  body, herr := c.Send(ctx, "/ddns_nv.php", params, nil, false)
  if herr != nil {
    return perror.New(perror.ERROR_DNS_SERVICE_CONNECT, "Failed to contact DDNS Server code: #" + strconv.Itoa(herr.Code) + ": " + herr.Message)
  }

  logger.L(ctx).Debugf("Response %s", body)

  dr := &DNSResponse{}
  err = json.Unmarshal([]byte(body), dr)
  if err != nil {
    logger.L(ctx).Errorf("Failed to parse DNS server response: %s", err.Error())
    return perror.New(perror.ERROR_PARSE_DNS_RESPONSE, "Failed to parse DNS server response")
  }

  if dr.Status !=  "success" {
    logger.L(ctx).Errorf("Request Failed. DNS returned status: %s", dr.Status)
    return perror.New(perror.ERROR_DNS_SERVICE_RESPONSE, "Incorrect Status from the DNS server")
  }

  logger.L(ctx).Debugf("Record for %s set successfully", name)
  return nil
}

func (v *DNSService) DeleteName(ctx context.Context, cc *cloudcoin.CloudCoin) error {
  ip := cloudcoin.GetIPFromSn(cc.Sn)
  name := cc.GetSkyName()
  if name == "" {
    return perror.New(perror.ERROR_NO_SKYWALLET_NAME, "SkyWallet Name is not defined")
  }

  logger.L(ctx).Debugf("Deleting Name %s for sn %d ip %s", name, cc.Sn, ip)

  rr, err := v.GetRandomRaida(ctx)
  if err != nil {
    logger.L(ctx).Errorf("Failed to pick random raida")
    return err
  }

  logger.L(ctx).Debugf("Requesting ticket from raida %d", rr)

  t := raida.NewGetTicket(v.progressChannel)
  response, err := t.GetSingleTicketSky(ctx, cc, rr)
  if err != nil {
    logger.L(ctx).Errorf("Failed to get ticket")
    return perror.New(perror.ERROR_GET_TICKET, "Failed to get ticket")
  }

  c := httpclient.New(config.DNSSERVICE_URL, 0)

  params := make(map[string]string, 0)
  params["sn"] = strconv.Itoa(int(cc.Sn))
  params["ticket"] = response
  params["username"] = name
  params["raidanumber"] = strconv.Itoa(rr)

  body, herr := c.Send(ctx, "/ddns_delete_nv.php", params, nil, false)
  if herr != nil {
    return perror.New(perror.ERROR_DNS_SERVICE_CONNECT, "Failed to contact DDNS Server code: #" + strconv.Itoa(herr.Code) + ": " + herr.Message)
  }

  logger.L(ctx).Debugf("Response %s", body)

  dr := &DNSResponse{}
  err = json.Unmarshal([]byte(body), dr)
  if err != nil {
    logger.L(ctx).Errorf("Failed to parse DNS server response: %s", err.Error())
    return perror.New(perror.ERROR_PARSE_DNS_RESPONSE, "Failed to parse DNS server response")
  }

  if dr.Status !=  "success" {
    logger.L(ctx).Errorf("Request Failed. DNS returned status: %s", dr.Status)
    return perror.New(perror.ERROR_DNS_SERVICE_RESPONSE, "Incorrect Status from the DNS server")
  }

  logger.L(ctx).Debugf("Record for %s set successfully")
  return nil
}

func (v *DNSService) GetRandomRaida(ctx context.Context) (int, error) {
  rand.Seed(time.Now().UnixNano())

  var rn int

  rn = -1
  for i := 0; i < config.MAX_ATTEMPTS_TO_PICK_RANDOM_RAIDA; i++ {
    rn = rand.Intn(config.TOTAL_RAIDA_NUMBER)
    logger.L(ctx).Debugf("Generated Random Raida for Ticket %d", rn)

    raida.RaidaList.Mutex.Lock()
    rl := raida.RaidaList.ActiveRaidaList[rn]
    raida.RaidaList.Mutex.Unlock()

    if rl == nil {
      logger.L(ctx).Debugf("RAIDA %d is unavaliable. Trying next one...", rn)
      continue
    }

    break
  }

  if rn == -1 {
    return 0, perror.New(perror.ERROR_PICKING_RANDOM_RAIDA, "Failed to pick random RAIDA. Attempts exceeded")
  }

  return rn, nil
}


