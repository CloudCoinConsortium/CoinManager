package settings

import (
	"context"
	"errors"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
)



type SettingsManager struct {

}

type Settings struct {
  DefaultTimeoutMult int `json:"default_timeout_mult"`
  EchoTimeoutMult int `json:"echo_timeout_mult"`
  MaxNotes int `json:"max_notes"`
  ChangeServerSN int `json:"change_server_sn"`
  UseLocalRaidas bool `json:"use_local_raidas"`
  LocalRaidas []string `json:"local_raidas"`
}

func New() *SettingsManager {
  return &SettingsManager{}
}

func (v *SettingsManager) Load(ctx context.Context) (*Settings, error) {
  logger.L(ctx).Debugf("Loading settings")

  rconf, err := config.ReadApplyConfig()
  if err != nil {
    logger.L(ctx).Errorf("Failed to read config: %s", err.Error())
    return nil, err
  }

  s := &Settings{}
  s.DefaultTimeoutMult = rconf.Main.DefaultTimeoutMult
  s.EchoTimeoutMult = rconf.Main.EchoTimeoutMult
  s.MaxNotes = rconf.Main.MaxNotesToSend
  s.ChangeServerSN = rconf.Main.ChangeServerSN
  s.UseLocalRaidas = rconf.Main.UseLocalRaidas
  s.LocalRaidas = rconf.Main.LocalRaidas

  return s, nil
}

func (v *SettingsManager) Save(ctx context.Context, s *Settings) error {
  logger.L(ctx).Debugf("Saving settings %v", s)

  rconf, err := config.ReadApplyConfig()
  if err != nil {
    logger.L(ctx).Errorf("Failed to read config: %s", err.Error())
    return err
  }

  rconf.Main.DefaultTimeoutMult = s.DefaultTimeoutMult
  rconf.Main.EchoTimeoutMult = s.EchoTimeoutMult
  rconf.Main.MaxNotesToSend = s.MaxNotes
  rconf.Main.ChangeServerSN = s.ChangeServerSN
  rconf.Main.UseLocalRaidas = s.UseLocalRaidas
  rconf.Main.LocalRaidas = s.LocalRaidas

  if s.UseLocalRaidas && len(s.LocalRaidas) != config.TOTAL_RAIDA_NUMBER {
    return errors.New("Invalid number of LocalRaida servers. Must be equal to " + strconv.Itoa(config.TOTAL_RAIDA_NUMBER))
  }


  err = config.ApplySave(rconf)
  if err != nil {
    logger.L(ctx).Errorf("Failed to save config: %s", err.Error())
    return err
  }

  return nil

}


