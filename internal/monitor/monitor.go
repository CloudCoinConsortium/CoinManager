package monitor

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
)



func RunMonitorTask() {
  ctx := context.Background()
  ctx = logger.SetContext(ctx, "SysEchoMonitor")

  logger.L(ctx).Debugf("Running Monitor Task")

  for idx, url := range(raida.RaidaList.ActiveRaidaList) {
    if url == nil {
      raida.RaidaList.Mutex.Lock()
      raida.RaidaList.ActiveRaidaList[idx] = &raida.RaidaList.PrimaryRaidaList[idx]
      raida.RaidaList.Mutex.Unlock()
    }
  }

  if config.USE_LOCAL_RAIDAS {
    logger.L(ctx).Debug("Running Monitor for Local RAIDA servers")
    DoMonitorLocalTask(ctx)
  } else {
    DoMonitorTask(ctx)
  }

  for idx, url := range(raida.RaidaList.ActiveRaidaList) {
    var u string
    if url == nil {
      u = "none"
    } else {
      u = *url
    }

    logger.L(ctx).Debugf("i=%d %s", idx, u)
  }

  go func() {
    for true {
      time.Sleep(config.MONITOR_PERIOD * time.Second)
      DoMonitorTask(ctx)
    }
  }()


  logger.L(nil).Debugf("Monitor Task Launched")
}

type Pair struct {
  Key int
  Value int64
}

type PairList []Pair

func (p PairList) Len() int { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }

func UpdatePownArrayCalcDeviation(ctx context.Context, eo *raida.EchoOutput) []int {
  return UpdatePownArrayCommon(ctx, eo, true)
}

func UpdatePownArrayCommon(ctx context.Context, eo *raida.EchoOutput, calcDeviation bool) []int {
  pownArray := make([]int, config.TOTAL_RAIDA_NUMBER)
  for idx, v := range(eo.PownArray) {
    pownArray[idx] = v
  }

  var sum, sigma int64

  sum = 0
  rmap := make(map[int]int64, 0)

  for k, v := range(eo.Latencies) {
    if eo.PownArray[k] != config.RAIDA_STATUS_PASS {
      logger.L(ctx).Debugf("r%d marked unavailable", k)
      continue
    }

    logger.L(ctx).Debugf("r%d latency %d ms", k, v)
    rmap[k] = v
    sum += v
  }

  if len(rmap) == 0 {
    logger.L(ctx).Debugf("No RAIDA servers responded")
    return pownArray
  }

  mean := sum / int64(len(rmap))
  for _, v := range(rmap) {
    sigma += int64(math.Pow(math.Abs(float64(v - mean)), 2))
  }

  sigma /= int64(len(rmap))
  sigma = int64(math.Sqrt(float64(sigma)))

  dv := int64(math.Abs(float64(mean + 2 * sigma)))
  logger.L(ctx).Debugf("%d RAIDA servers replied to Echo. Mean is %dms, sigma is %dms, latency + 2nd deviation is %dms", len(rmap), mean, sigma, dv)

  if calcDeviation {
    logger.L(ctx).Debugf("Second timeout is set to %dms", dv)
    config.GLOBAL_NEXT_2ND_DEVIATION_TIMEOUT_MS = int(dv)
  }


  pl := make(PairList, len(rmap))
  i := 0

  for k, v := range(rmap) {
    pl[i] = Pair{k, v}
    i++
  }

  sort.Sort(pl)

  var pval int64

  pval = 0
  fidx := -1

  for idx, p := range(pl) {
    ridx := p.Key
    latency := p.Value

    if pval == 0 {
      pval = latency
      logger.L(ctx).Debugf("r%d latency %d ms", ridx, latency)
      continue
    }

    diff := latency - pval
    pval = latency

    if diff > dv {
      logger.L(ctx).Debugf("The difference (%dms) between responses exceeded next timeout %dms. Marking the rest of the raida servers as unavailable", diff, dv)
      fidx = idx
      break
    }

    logger.L(ctx).Debugf("r%d latency %dms, diff %dms", ridx, latency, diff)
  }


  if fidx != -1 {
    for _, p := range(pl[fidx:]) {
      ridx := p.Key
      logger.L(ctx).Debugf("Marking r%d as unavailable", ridx)

      pownArray[ridx] = config.RAIDA_STATUS_NORESPONSE
    }
  }




  return pownArray
}

func DoMonitorTask(ctx context.Context) {
  logger.L(ctx).Debugf("Running Monitor Echo")

  e := raida.NewEcho(nil)

  logger.L(ctx).Debugf("Pinging Primary Servers")
  e.SetPrivateActiveRaidaList(&raida.RaidaList.PrimaryRaidaList)
  poutput, err := e.Echo(ctx)
  if err != nil {
    logger.L(ctx).Debugf("Monitor Task: Echo Failed. %s", err.Error())
    return
  }

  pPownArray := UpdatePownArrayCalcDeviation(ctx, poutput)

  logger.L(ctx).Debugf("Pinging Backup Servers")
  e.SetPrivateActiveRaidaList(&raida.RaidaList.BackupRaidaList)
  boutput, err := e.Echo(ctx)
  if err != nil {
    logger.L(ctx).Debugf("Monitor Task: Echo Failed. %s", err.Error())
    return
  }

  bPownArray := UpdatePownArrayCommon(ctx, boutput, false)

  raida.RaidaList.Mutex.Lock()
  for idx := 0; idx < len(raida.RaidaList.ActiveRaidaList); idx++ {
    if raida.RaidaList.ActiveRaidaList[idx] == &raida.RaidaList.PrimaryRaidaList[idx] {
      logger.L(ctx).Debugf("r%d was set to primary %s", idx, *raida.RaidaList.ActiveRaidaList[idx])

      // Check primary array
      presult := pPownArray[idx]
      if presult != config.RAIDA_STATUS_PASS {
        logger.L(ctx).Debugf("Switching r%d to backup", idx)
        if raida.RaidaList.BackupRaidaList[idx] == "" {
          logger.L(ctx).Debugf("No backup for Raida%d defined on the Guardians. Discarding this Raida", idx)
          raida.RaidaList.ActiveRaidaList[idx] = nil
        } else {
          logger.L(ctx).Debugf("Switching Raida%d to backup %s", idx, raida.RaidaList.BackupRaidaList[idx])

          bresult := bPownArray[idx]
          if bresult != config.RAIDA_STATUS_PASS {
            logger.L(ctx).Debugf("Backup URL for r%d is unavailable as well! Giving up for this raida", idx)
            raida.RaidaList.ActiveRaidaList[idx] = nil
          } else {
            logger.L(ctx).Debugf("Set backup url %s for raida %d", raida.RaidaList.BackupRaidaList[idx], idx)
            raida.RaidaList.ActiveRaidaList[idx] = &raida.RaidaList.BackupRaidaList[idx]
          }
        }
      }
    } else if raida.RaidaList.ActiveRaidaList[idx] == &raida.RaidaList.BackupRaidaList[idx] {
      logger.L(ctx).Debugf("r%d was set to backup %s", idx, *raida.RaidaList.ActiveRaidaList[idx])
  
      // First check if Primary RAIDA is back again
      presult := pPownArray[idx]
      if presult == config.RAIDA_STATUS_PASS {
        logger.L(ctx).Debugf("Primary RAIDA is up again. Switching r%d to primary", idx)
        raida.RaidaList.ActiveRaidaList[idx] = &raida.RaidaList.PrimaryRaidaList[idx]
      } else {
        // Check backup raida again
        bresult := bPownArray[idx]
        if bresult != config.RAIDA_STATUS_PASS {
          logger.L(ctx).Debugf("Backup URL for r%d is unavailable. Discarding this raida")
          raida.RaidaList.ActiveRaidaList[idx] = nil
        } 
        // Doing nothing. Preserve the backup server for this raida
      }
    } else if raida.RaidaList.ActiveRaidaList[idx] == nil {
      logger.L(ctx).Debugf("Raida%d was discarded. See if we can bring it back", idx)
      // First check if Primary RAIDA is back again
      presult := pPownArray[idx]
      if presult == config.RAIDA_STATUS_PASS {
        logger.L(ctx).Debugf("Primary RAIDA is up. Switching r%d to primary", idx)
        raida.RaidaList.ActiveRaidaList[idx] = &raida.RaidaList.PrimaryRaidaList[idx]
      } else {
        // Check backup raida 
        bresult := bPownArray[idx]
        if bresult == config.RAIDA_STATUS_PASS {
          logger.L(ctx).Debugf("Backup URL for r%d is up. Setting backu")
          raida.RaidaList.ActiveRaidaList[idx] = &raida.RaidaList.BackupRaidaList[idx]
        }
      }
    } else {
      logger.L(ctx).Debugf("Unknown active raida list r%d: setting to nil", idx)
      raida.RaidaList.ActiveRaidaList[idx] = nil
    }

  }
  raida.RaidaList.Mutex.Unlock()

  logger.L(ctx).Debugf("Finished Monitor Echo")

  ShowRaidaList(ctx)
}

func ShowRaidaList(ctx context.Context, ) {
  statusString := ""

  raida.RaidaList.Mutex.Lock()
  for idx := 0; idx < len(raida.RaidaList.ActiveRaidaList); idx++ {
    if raida.RaidaList.ActiveRaidaList[idx] == &raida.RaidaList.PrimaryRaidaList[idx] {
      statusString += "p"
    } else if raida.RaidaList.ActiveRaidaList[idx] == &raida.RaidaList.BackupRaidaList[idx] {
      statusString += "b"
    } else {
      statusString += "-"
    }
  }
  raida.RaidaList.Mutex.Unlock()

  logger.L(ctx).Debug("RaidaChosen after Monitor: 'p' - primary, 'b' - backup, '-' - discarded")
  logger.L(ctx).Debugf("%s", statusString)
}

func DoMonitorLocalTask(ctx context.Context, ) {
  logger.L(ctx).Debugf("Running Monitor of Local Raida servers Echo")

  e := raida.NewEcho(nil)

  logger.L(ctx).Debugf("Pinging Primary Servers")
  e.SetPrivateActiveRaidaList(&raida.RaidaList.PrimaryRaidaList)
  poutput, err := e.Echo(ctx)
  if err != nil {
    logger.L(ctx).Debugf("Monitor Task: Echo Failed. %s", err.Error())
    return
  }

  pPownArray := UpdatePownArrayCalcDeviation(ctx, poutput)

  raida.RaidaList.Mutex.Lock()
  for idx := 0; idx < len(raida.RaidaList.ActiveRaidaList); idx++ {
    // Check primary array
    presult := pPownArray[idx]
    if presult != config.RAIDA_STATUS_PASS {
      logger.L(ctx).Debugf("Raida failed r%d ", idx)
      raida.RaidaList.ActiveRaidaList[idx] = nil
    } else {
      logger.L(ctx).Debugf("Raida available r%d ", idx)
      raida.RaidaList.ActiveRaidaList[idx] = &raida.RaidaList.PrimaryRaidaList[idx]
    }
  }

  raida.RaidaList.Mutex.Unlock()

  logger.L(ctx).Debugf("Finished Monitor Echo")

  ShowRaidaList(ctx)
}

