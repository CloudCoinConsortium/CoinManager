//go:generate goversioninfo -icon=./256x256.ico -64 -manifest=./cloudcoin_manager.manifest
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/api"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/freecoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/monitor"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/showpayment"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/tasks"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/ui"
	"github.com/gorilla/mux"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/core"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	//	testecho "github.com/CloudCoinConsortium/superraidaclientbackend/internal/tests/echo"
	//	testfix "github.com/CloudCoinConsortium/superraidaclientbackend/internal/tests/fix"
	//	testdetect "github.com/CloudCoinConsortium/superraidaclientbackend/internal/tests/detect"
)


var raida_tester string

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s [-debug] <operation> <args>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s [-help]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s [-version]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "<args> arguments for operation\n\n")
	flag.PrintDefaults()
}

func Version() {
	fmt.Printf("%s\n", config.VERSION)
}


var router *mux.Router

func main() {

  home := ""

	//flag.StringVar(&config.CmdCommand, "", "", "Operation")
	flag.BoolVar(&config.CmdDebug, "debug", false, "Display Debug Information")
	flag.BoolVar(&config.CmdCli, "cli", false, "RUN in CLI mode")
	flag.StringVar(&home, "home", "", "Home Directory")
	flag.BoolVar(&config.CmdHelp, "help", false, "Show Usage")
	flag.StringVar(&config.CmdUIPath, "uipath", "", "UI Path")
	flag.BoolVar(&config.CmdVersion, "version", false, "Display version")
	flag.Usage = Usage
	flag.Parse()


  config.CmdDebug = true

	if config.CmdVersion {
		Version()
		os.Exit(0)
	}

	if config.CmdHelp {
		Usage()
		os.Exit(0)
	}

	if err := config.SetRootPath(home); err != nil {
		core.ExitError(err)
	}

	if err := core.CreateRootFolder(); err != nil {
		core.ExitError(err)
	}

	if err := logger.Init(config.ROOT_PATH, true); err != nil {
		core.ExitError(err)
	}

	if err := core.UpdateAssets(); err != nil {
		core.ExitError(err)
	}


  if config.CmdCli {
    DoBack()
    return
  }

  uipath, err := core.GetUIPath(config.CmdUIPath)
  if err != nil {
    core.ExitError(err)
  }

  logger.L(nil).Debugf("UIPath %s", uipath)

  go DoBack()
  //DoBack()

	ui.DoUI(uipath)
}

func DoBack() {

	logger.L(nil).Infof("Program started. Version %s", config.VERSION)
	logger.L(nil).Info(strings.Join(os.Args, " "))
	logger.L(nil).Debugf("Reading config %s", config.GetConfigPath())

	if _, err := config.ReadApplyConfig(); err != nil {
		core.ExitError(err)
	}


	logger.L(nil).Debugf("Initializting Tasks")
	tasks.InitTasks()

	logger.L(nil).Debugf("Initializting Storage")
	if err := storage.Init(); err != nil {
		core.ExitError(err)
	}

	logger.L(nil).Debugf("Initializting Guardians or LocalRAIDAs")
  if config.USE_LOCAL_RAIDAS {
    logger.L(nil).Debugf("We will use Local RAIDA hints file")
    if len(config.LocalRaidas) != config.TOTAL_RAIDA_NUMBER {
      core.ExitError(errors.New("Invalid number of LocalRAIDA servers in the config file"))
    }
    raida.InitRAIDAList()
  } else {
    logger.L(nil).Debugf("We will query Guadians to find out RAIDA IP addresses")
  	if err := raida.InitGuardians(); err != nil {
	  	core.ExitError(err)
  	}
  }

  // DIsable temporarily
	if config.CmdCli && 2 == 3 {
		logger.L(nil).Debugf("CLI Mode")

		if flag.NArg() == 0 {
			core.ExitError(perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "Command is missing"))
		}

		command := flag.Arg(0)
		args := flag.Args()

		// Blocking
		monitor.DoMonitorTask(nil)

		out, err := DoCommand(command, args[1:])
		if err != nil {
			core.ExitError(err)
		}

		fmt.Println(out)
		os.Exit(0)
	}

	logger.L(nil).Debugf("Initializting RAIDA Monitor")
	monitor.RunMonitorTask()


	logger.L(nil).Debugf("Initializting Web Server")
	handleRequests()

}


func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))

	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))

	return
}

func setContext(h http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    var matchInfo mux.RouteMatch
    b := router.Match(r, &matchInfo)
    if b {
      if matchInfo.MatchErr == nil {
        name := matchInfo.Route.GetName()
        if name == "" {
          name = r.RequestURI
        }

        ctx := r.Context()
        ctx = logger.SetContext(ctx, name)

        r = r.WithContext(ctx)
      }
    }

    //logger.L(nil).Debugf("rrrrr %s m=%v n=%s", r.RequestURI, b, matchInfo.Route.GetName())
    h.ServeHTTP(w, r)
  })
}

/* Catch-All Function */
func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return 
			}
		}

    ctx := r.Context()
		logger.L(ctx).Debugf("Request %s %s", r.Method, r.URL.String())
		w.Header().Set("Content-Type", "application/json")

		h.ServeHTTP(w, r)
	})
}

func defReqOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	api.SuccessResponse(nil, w, nil)
}


func defReq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	logger.L(nil).Debugf("Invalid URL requested %s", r.URL.String())
	api.ErrorResponse(nil, w, perror.New(perror.ERROR_SERVICE_NOT_FOUND, "Service Not Found. Invalid URL"))
}










func handleRequests() {

	router = mux.NewRouter()

	router.HandleFunc("/", defReqOptions)
	router.HandleFunc("/api/v1/echo", api.EchoReq).Methods("GET").Name("echo")
	router.HandleFunc("/api/v1/detect", api.DetectReq).Methods("POST").Name("detect")
	router.HandleFunc("/api/v1/pown", api.PownReq).Methods("POST").Name("pown")
	router.HandleFunc("/api/v1/wallets", api.WalletsReq).Methods("GET").Name("getwallets")
	router.HandleFunc("/api/v1/wallets", api.CreateWalletReq).Methods("POST").Name("createwallet")
	router.HandleFunc("/api/v1/wallets/{name}", api.WalletReq).Methods("GET").Name("getwallet")
	router.HandleFunc("/api/v1/wallets/{name}", api.DeleteWalletReq).Methods("DELETE").Name("deletewallet")
	router.HandleFunc("/api/v1/wallets/{name}", api.UpdateWalletReq).Methods("PUT").Name("updatewallet")
	router.HandleFunc("/api/v1/unpack", api.UnpackReq).Methods("POST").Name("unpack")
	router.HandleFunc("/api/v1/import", api.ImportReq).Methods("POST").Name("import")
	router.HandleFunc("/api/v1/transfer", api.TransferReq).Methods("POST").Name("transfer")
	router.HandleFunc("/api/v1/export", api.ExportReq).Methods("POST").Name("export")
	router.HandleFunc("/api/v1/deposit", api.DepositReq).Methods("POST").Name("deposit")
	router.HandleFunc("/api/v1/tasks/{id}", api.TaskReq).Methods("GET").Name("gettask")
	router.HandleFunc("/api/v1/info", api.InfoReq).Methods("GET").Name("info")
	router.HandleFunc("/api/v1/news", api.NewsReq).Methods("GET").Name("news")
	router.HandleFunc("/api/v1/freecoin", api.FreeCoinReq).Methods("GET").Name("freecoin")
	router.HandleFunc("/api/v1/skywallets", api.SkyWalletsReq).Methods("GET").Name("getskyvaults")
	router.HandleFunc("/api/v1/skywallets", api.CreateSkyWalletReq).Methods("POST").Name("createsky")
	router.HandleFunc("/api/v1/skywallets/{name}", api.SkyWalletReq).Methods("GET").Name("getskyvault")
	router.HandleFunc("/api/v1/skywallets/{name}", api.DeleteSkyWalletReq).Methods("DELETE").Name("deletesky")
	router.HandleFunc("/api/v1/skywallets/{name}", api.UpdateSkyWalletReq).Methods("POST").Name("updatesky")
	router.HandleFunc("/api/v1/skytransfer", api.SkyTransferReq).Methods("POST").Name("transfersky")
	router.HandleFunc("/api/v1/settings", api.PostSettingsReq).Methods("POST").Name("updatesettings")
	router.HandleFunc("/api/v1/settings", api.GetSettingsReq).Methods("GET").Name("getsettings")
	router.HandleFunc("/api/v1/fix", api.FixReq).Methods("POST").Name("fix")
	router.HandleFunc("/api/v1/fix", api.FixAllReq).Methods("PUT").Name("fixall")
	router.HandleFunc("/api/v1/withdraw", api.WithdrawReq).Methods("POST").Name("withdraw")
	router.HandleFunc("/api/v1/version", api.VersionReq).Methods("GET").Name("version")
	router.HandleFunc("/api/v1/showpayment", api.ShowPaymentReq).Methods("POST").Name("showpayment")
	router.HandleFunc("/api/v1/health", api.HealthReq).Methods("POST").Name("health")
	router.HandleFunc("/api/v1/skyhealth", api.SkyHealthReq).Methods("POST").Name("skyhealth")
	router.HandleFunc("/api/v1/statements/{name}", api.DeleteStatementsReq).Methods("DELETE").Name("deletestatements")
	router.HandleFunc("/api/v1/sync", api.SyncReq).Methods("POST").Name("sync")
	router.HandleFunc("/api/v1/skydetect", api.SkyDetectReq).Methods("POST").Name("skydetect")
	router.HandleFunc("/api/v1/skyfix", api.SkyFixReq).Methods("POST").Name("skyfix")
	router.HandleFunc("/api/v1/backup", api.BackupReq).Methods("POST").Name("backup")
	router.HandleFunc("/api/v1/skybackup", api.SkyBackupReq).Methods("POST").Name("skybackup")
	router.HandleFunc("/api/v1/transactions/{name}/{guid}", api.GetTransactionsReq).Methods("GET").Name("transactions")
	router.HandleFunc("/api/v1/transactions/{name}", api.DeleteTransactionsReq).Methods("DELETE").Name("deletetransactions")
	router.HandleFunc("/api/v1/convert", api.ConvertReq).Methods("POST").Name("convert")
	router.HandleFunc("/api/v1/filepicker", api.FilePickerReq).Methods("GET").Name("filepicker")
	router.HandleFunc("/api/v1/senderhistory", api.SenderHistoryReq).Methods("GET").Name("history")
	router.HandleFunc("/api/v1/wallets/{name}/leftovers", api.LeftOverReq).Methods("GET").Name("leftovers")
	router.HandleFunc("/api/v1/skytransactions/{name}/{guid}", api.GetSkyTransactionsReq).Methods("GET").Name("skytransactions")

	router.NotFoundHandler = http.HandlerFunc(defReq)

	logger.L(nil).Debug("HTTP Router Configured")

	//http.ListenAndServe(":" + strconv.Itoa(config.LISTEN_PORT), nil)
	err := http.ListenAndServe(":" + strconv.Itoa(config.LISTEN_PORT), setContext(allowCORS(router)))
	if err != nil {
		logger.L(nil).Errorf("Failed to listen: %s", err.Error())
		core.ExitError(perror.New(perror.ERROR_LISTEN, "Failed to listen: " + err.Error()))
	}

	logger.L(nil).Debug("HTTP Server started")
}

func DoCommand(command string, args []string) (string, error) {
	logger.L(nil).Debugf("Doing command %s", command)

	var commands = map[string]interface{}{
		"showpayment": &showpayment.Caller{},
		"adminfreeid": &freecoin.Caller{},

//	  "testecho": testecho.New(),
//	  "testfix": testfix.New(),
//	  "testdetect": testdetect.New(),
	}

	caller, ok := commands[command]
	if !ok {
		return "", perror.New(perror.ERROR_COMMAND_NOT_FOUND, "Command not found")
	}

	f := reflect.ValueOf(caller).MethodByName("DoCommand")

	in := make([]reflect.Value, 1)
	in[0] = reflect.ValueOf(args)

	res := f.Call(in)
	if len(res) != 2 {
		return "", perror.New(perror.ERROR_INTERNAL, "Invalid number of arguments returned by the function")
	}

	out := res[0].Interface()
	errIface := res[1].Interface()
	if errIface != nil {
		err := errIface.(error)
		logger.L(nil).Debugf("Function %s returned error %s", command, err.Error())
		return "", err
	}

	jsonstr, err := json.Marshal(out)
	if err != nil {
		return "", perror.New(perror.ERROR_ENCODE_OUTPUT_JSON, "Failed to encode JSON: " + err.Error())
	}

	return string(jsonstr), nil
}
