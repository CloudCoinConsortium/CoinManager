package api

import (
	"encoding/json"
	"net/http"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/settings"
)

type SettingsResponse struct {
}


func GetSettingsReq(w http.ResponseWriter, r *http.Request) {  
  ctx := r.Context()
  logger.L(ctx).Debugf("Settings Request")

  sm := settings.New()
  response, err := sm.Load(ctx)
  if err != nil {
    logger.L(ctx).Debugf("Failed to load json: %s", err.Error())
    ErrorResponse(ctx, w, perror.New(perror.ERROR_LOAD_SETTINGS, "Failed to load settings: " + err.Error()))
    return
  }

  SuccessResponse(ctx, w, response)
}

func PostSettingsReq(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  logger.L(ctx).Debugf("Settings Update Request")

  var lsettings settings.Settings

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&lsettings)
  if err != nil {
    ErrorResponse(ctx, w, perror.New(perror.ERROR_DECODE_INPUT_JSON, "Failed to parse Input JSON"))
    return
  }

  // No Validation. TOML library will validate it
  sm := settings.New()
  err = sm.Save(ctx, &lsettings)
  if err != nil {
    logger.L(ctx).Debugf("Failed to save settings: %s", err.Error())
    ErrorResponse(ctx, w, perror.New(perror.ERROR_SAVE_SETTINGS, "Failed to save settings: " + err.Error()))
    return
  }

  SuccessResponse(ctx, w, nil)
}


