package ui

import (
	"os"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/core"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/polevpn/webview"
)

func DoUI(uipath string) {
	debug := true
  
  logger.L(nil).Debugf("UI launched")

	w := webview.New(1280, 960, false, debug)

  logger.L(nil).Debugf("UI Webview created")

	defer w.Destroy()
	w.SetTitle("CloudCoin Manager " + config.VERSION)
	w.SetSize(1280, 960, webview.HintNone)
 // w.SetIcon(iconbytes)
  w.SetIcon([]byte("256x256.ico"))


  logger.L(nil).Debugf("UI Ready to Navigate")
  w.Navigate("file://" + uipath)

  core.SetUIHandle(w.Window())
  logger.L(nil).Debugf("UI Navigated")
	w.Run()


  logger.L(nil).Debugf("Window closed")
  os.Exit(0)
}
