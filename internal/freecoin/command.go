package freecoin

import (
	"strconv"
	"strings"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
)

type Caller struct {
}

type FCCommandResponse struct {
  SavedCoins int `json:"saved_coins"`
  FailedCoins int `json:"failed_coins"`
}

func (v *Caller) DoCommand(args []string) (interface{}, error) {

  logger.L(nil).Debugf("GET ADMIN FREE COIN")

  if len(args) != 3 {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "RAIDA ID, SN Range and password are required for FreeID: ./cloudcoin_manager -cli adminfreeid 14 14000:14999 def57ecd47565605552a2d26b221f140")
  }

  raidaID := args[0]
  snrange := args[1]
  password := args[2]

  iRaidaID, err := strconv.Atoi(raidaID)
  if err != nil {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "Invalid RAIDA ID")
  }

  if iRaidaID < 0 || iRaidaID >= config.TOTAL_RAIDA_NUMBER {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "Invalid RAIDA ID")
  }

  arraySNs := strings.Split(snrange, ":")
  if len(arraySNs) != 2 {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "Invalid SN Range must be in MINSN:MAXSN format")
  }

  isnmin, err := strconv.Atoi(arraySNs[0])
  if err != nil {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "Invalid Min Serial Number")
  }

  isnmax, err := strconv.Atoi(arraySNs[1])
  if err != nil {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "Invalid Max Serial Number")
  }

  usnmin := uint32(isnmin)
  usnmax := uint32(isnmax)

  // 1,000 SN for each RAIDA server
  minAllowedSN := uint32(1000 * iRaidaID)
  maxAllowedSN := minAllowedSN + 999

  if usnmin < minAllowedSN || usnmax > maxAllowedSN {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "This SN Range is not allowed for this RAIDA ID")
  }

  if usnmin > usnmax {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "This SN Range is incorrect")
  }

  logger.L(nil).Debugf("pas %s", password)
  RaidaPasswords := map[int]string{
    0: "e6cdb059a844ba675cc4176fc8fedaf7", 
    1: "3aa4521c9e69fb1861985ec185dca1b2", 
    2: "9863e2c6361a140ded59a490a1fb93c2", 
    3: "3a6dba9748ed1a92f63a38acc25e7659", 
    4: "69e981c7b72d370243e30c1c4d130b39", 
    5: "64d542ffe3ccaa5bb58645a43691f69f", 
    6: "a8fddf0b90ae87b322f29679b2dbe632", 
    7: "fb5e958720a418b858e42682912d120f", 
    8: "b22cbc465d783002e339333f2a1c8d86", 
    9: "39fa97850b18d3819ee3be117afbd0e5", 
    10: "dc32a4403144ff22f0ef0a9b951fe1dd", 
    11: "fd002909680e81b99c8798f37ad71d35", 
    12: "4ef35759ea82a0dba889f0a3a241523a", 
    13: "d22d1b57b8845bb9a36925d952128bef", 
    14: "df257ecd4756560fff2a2d26b221fbb0", 
    15: "ab08be5e762c6e86a959d336b763a6d5", 
    16: "723aa478b6b9477b033ac2b9c70c82c5", 
    17: "db231969a491cb058e246b0156220f96", 
    18: "a3351c9a1d6a176e280b435447fdf8cc", 
    19: "af2f130f52f851a3384dbc7132ce3a27", 
    20: "ddd4526565175ad285fb0d100033b0a4", 
    21: "3a70f3ca9225239d0e2e818e972fe781", 
    22: "eaab9454eb5c9fd8de5fd6dfd498173b", 
    23: "aa95b1685e99164bfdb6b1c099dbe565", 
    24: "5792e684c4204f070b632a9361f4099d", 
  }

  rp := RaidaPasswords[iRaidaID]
  if rp != password {
    return "", perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "Incorrect Password")
  }

  fcr := &FCCommandResponse{}
  fcr.SavedCoins = 0
  fcr.FailedCoins = 0

  fc, _ := New(nil)
  for i := usnmin; i <= usnmax; i++ {
    logger.L(nil).Debugf("Getting sn %d", i)
    response, err := fc.GetSpecificCoin(nil, i)
    if err != nil {
      logger.L(nil).Debugf("Failed to get coin %d: %s", i, err.Error())
      fcr.FailedCoins++
      continue
    }

    cc := cloudcoin.NewFromData(response.Sn)
    cc.SetAns(response.Ans)

    err = storage.GetDriver().SaveCoin(nil, cc, config.DEFAULT_WALLET_NAME, config.DIR_BANK)
    if err != nil {
      logger.L(nil).Debugf("Failed to save coin %d", i, err.Error())
      fcr.FailedCoins++
      continue
    }

    fcr.SavedCoins++
  }


/*
  response, err := fc.GetSpecificCoin(usn)
  if err != nil {
    return "", err
  }

  logger.L(nil).Debugf("resp: %v", response.Sn, response.Ans)
*/
  /*

  sp := New(nil)
  out, err := sp.Show(guid)
  if err != nil {
    return "", err
  }
  */
  // return out, nil
  return fcr, nil
}
