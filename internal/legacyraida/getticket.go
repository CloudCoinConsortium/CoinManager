package legacyraida

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
)


type GetTicket struct {
	Servant
}

type GetTicketOutput struct {
  // Array of batches of 25 tickets
  Tmap []map[string][]cloudcoin.CloudCoin
  Details [][]string `json:"details,omitempty"`
}

type GetTicketResult struct {
  Status string `json:"status"`
  Ticket string `json:"ticket"`
  Message string `json:"message"`
}

func NewGetTicket(progressChannel chan interface{}) *GetTicket {
	return &GetTicket{
		*NewServant(progressChannel),
	}
}

func (v *GetTicket) GetTicket(ctx context.Context, coins []cloudcoin.CloudCoin) (*GetTicketOutput, error) {
	logger.L(ctx).Debugf("Getting ticketis %d notes", len(coins))

  finalTmap := make([]map[string][]cloudcoin.CloudCoin, config.TOTAL_RAIDA_NUMBER)
  for i := 0; i < config.TOTAL_RAIDA_NUMBER; i++ {
    finalTmap[i] = make(map[string][]cloudcoin.CloudCoin, 0)
  }
  stride := config.MAX_NOTES_TO_SEND
  var to = &GetTicketOutput{}
  for i := 0; i < len(coins); i += stride {
    max := i + stride
    if max > len(coins) {
      max = len(coins)
    }

    logger.L(ctx).Debugf("coins slice %d:%d, Processing %d notes", i, max, (max - i))
    tmaps, err := v.ProcessGetTicket(ctx, coins[i:max])
    if err != nil {
      logger.L(ctx).Errorf("Failed to process batch: %s", err.Error())
      perr := err.(*perror.ProgramError)
      to.Details = append(to.Details, perr.Details)
      return nil, err
    }

    for i := 0; i < config.TOTAL_RAIDA_NUMBER; i++ {
      tmap := tmaps[i]
      for ticket, _ := range(tmap) {
        finalTmap[i][ticket] = tmap[ticket]
      }
    }
  }

  for idx, _ := range(coins) {
    coins[idx].Grade()
    logger.L(ctx).Debugf("Coin %d: %s: %s", coins[idx].Sn, coins[idx].GetPownString(), coins[idx].GetGradeStatusString())
  }

  to.Tmap = finalTmap

  return to, nil
}

func (v *GetTicket) ProcessGetTicket(ctx context.Context, coins []cloudcoin.CloudCoin) ([]map[string][]cloudcoin.CloudCoin, error) {
	params := make([]string, config.TOTAL_RAIDA_NUMBER)
  command := "multi_detect"

  tmap := make([]map[string][]cloudcoin.CloudCoin, config.TOTAL_RAIDA_NUMBER)
  for i := 0; i < config.TOTAL_RAIDA_NUMBER; i++ {
    tmap[i] = make(map[string][]cloudcoin.CloudCoin, 0)
  }
  logger.L(ctx).Debugf("vvv %v", coins)
  // Iterate over raida servers
	for idx, _ := range params {
  	params[idx] = "b=t"
    for _, cc := range(coins) {
      params[idx] += "&nns[]=1"
      params[idx] += "&sns[]=" + strconv.Itoa(int(cc.Sn))
      params[idx] += "&denomination[]=" + strconv.Itoa(cc.GetDenomination())
      params[idx] += "&ans[]=" + cc.Ans[idx]
      params[idx] += "&pans[]=" + cc.Ans[idx]
    }
	}

  logger.L(ctx).Debugf("params %v", params)


	results := v.Raida.SendRequest(ctx, command, params)

  for idx, result := range(results) {
    if result.ErrCode != config.REMOTE_RESULT_ERROR_NONE {
      logger.L(ctx).Errorf("Error from oldraida %d", idx)
      if result.ErrCode == config.REMOTE_RESULT_ERROR_COMMON {
        for cidx, _ := range(coins) {
          coins[cidx].SetDetectStatus(idx, config.RAIDA_STATUS_ERROR)
        }
      } else if result.ErrCode == config.REMOTE_RESULT_ERROR_TIMEOUT {
        for cidx, _ := range(coins) {
          coins[cidx].SetDetectStatus(idx, config.RAIDA_STATUS_ERROR)
        }
      } else if result.ErrCode == config.REMOTE_RESULT_ERROR_SKIPPED {
        for cidx, _ := range(coins) {
          coins[cidx].SetDetectStatus(idx, config.RAIDA_STATUS_UNTRIED)
        }
      }
      continue
    }

    var gtr GetTicketResult
    err := json.Unmarshal([]byte(result.Data), &gtr)
    if err != nil {
      logger.L(ctx).Errorf("Json Parse Error from oldraida %d", idx)
      for cidx, _ := range(coins) {
        coins[cidx].SetDetectStatus(idx, config.RAIDA_STATUS_ERROR)
      }
      continue
    }

    if gtr.Status == "allpass" {
      ticket := gtr.Ticket
      logger.L(ctx).Debugf("All coins are authentic. Ticket %s", ticket)
      tmap[idx][ticket] = make([]cloudcoin.CloudCoin, 0)
      for cidx, _ := range(coins) {
        coins[cidx].SetDetectStatus(idx, config.RAIDA_STATUS_PASS)
        tmap[idx][ticket] = append(tmap[idx][ticket], coins[cidx])
      }
    } else if gtr.Status == "allfail" {
      logger.L(ctx).Debugf("All coins are counterfeit")
      for cidx, _ := range(coins) {
        coins[cidx].SetDetectStatus(idx, config.RAIDA_STATUS_FAIL)
      }
    } else if gtr.Status == "mixed" {
      logger.L(ctx).Debugf("Mixed response")
      ticket := gtr.Ticket
      rss := strings.Split(gtr.Message, ",")
      if len(rss) != len(coins) {
        logger.L(ctx).Errorf("Length mismatch. mixed length: %d, coins %d", len(rss), len(coins))
        for cidx, _ := range(coins) {
          coins[cidx].SetDetectStatus(idx, config.RAIDA_STATUS_ERROR)
        }
        continue
      }

      tmap[idx][ticket] = make([]cloudcoin.CloudCoin, 0)
      for i := 0; i < len(rss); i++ {
        rstatus := rss[i]
        if rstatus == "pass" {
          coins[i].SetDetectStatus(idx, config.RAIDA_STATUS_PASS)
          tmap[idx][ticket] = append(tmap[idx][ticket], coins[i])
        } else if rstatus == "fail" {
          logger.L(ctx).Debugf("coin %d is counterfeit on r%d", coins[i].Sn, idx)
          coins[i].SetDetectStatus(idx, config.RAIDA_STATUS_FAIL)
        } else {
          logger.L(ctx).Debugf("coin %d is weird on r%d status %s", coins[i].Sn, idx, rstatus)
          coins[i].SetDetectStatus(idx, config.RAIDA_STATUS_ERROR)
        }
      }


    } else {
      logger.L(ctx).Errorf("Invalid status from r%d: %s", idx, gtr.Status)
      for cidx, _ := range(coins) {
        coins[cidx].SetDetectStatus(idx, config.RAIDA_STATUS_ERROR)
      }
    }
  }

  logger.L(ctx).Debugf("vvv %v", results)

  return tmap, nil
}

