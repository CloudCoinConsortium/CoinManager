package legacyraida

import (
	"context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"

	//	"fmt"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	//	"config"
	//	"time"
)

type LegacyRAIDA struct {
	TotalNumber     int
	SideSize        int
	DetectionAgents []DetectionAgent
  PrivateActiveRaidaList *[]string
  BaseTimeouts map[uint16]int
}

func New(progressChannel chan interface{}) *LegacyRAIDA {
	DetectionAgents := make([]DetectionAgent, config.TOTAL_RAIDA_NUMBER)
	for idx, _ := range DetectionAgents {
		DetectionAgents[idx] = *NewDetectionAgent(idx, progressChannel)
	}

	return &LegacyRAIDA{
		TotalNumber:     config.TOTAL_RAIDA_NUMBER,
		DetectionAgents: DetectionAgents,
	}
}

func (r *LegacyRAIDA) FormErrorResults(code int, message string) []Result {
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

func (r *LegacyRAIDA) SendRequest(ctx context.Context, command string, data []string) []Result {
	done := make(chan Result)
	for idx, _ := range r.DetectionAgents {
		go func(agent *DetectionAgent, data string) {
      agent.SendRequest(ctx, command, data, done)
		}(&r.DetectionAgents[idx], data[idx])
	}

	results := make([]Result, r.TotalNumber)
	chanResults := make([]Result, r.TotalNumber)
	for i := 0; i < r.TotalNumber; i++ {
		chanResults[i] = <-done
	}

	logger.L(ctx).Debugf("Done request")
	for _, result := range chanResults {
    logger.L(ctx).Debugf("setting result %v rIdx %d", result, result.Index)
		results[result.Index] = result
	}

	return results
}
