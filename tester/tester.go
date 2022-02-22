//go:generate goversioninfo -icon=./256x256.ico -64 -manifest=./cloudcoin_manager.manifest
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/freecoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/monitor"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/raida"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/showpayment"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/core"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"

	testdetect "github.com/CloudCoinConsortium/superraidaclientbackend/internal/tests/detect"
	testecho "github.com/CloudCoinConsortium/superraidaclientbackend/internal/tests/echo"
	testfix "github.com/CloudCoinConsortium/superraidaclientbackend/internal/tests/fix"
	testsync "github.com/CloudCoinConsortium/superraidaclientbackend/internal/tests/sync"
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

	if err := logger.Init(config.ROOT_PATH, false); err != nil {
		core.ExitError(err)
	}

  ctx := context.Background()
  if config.CmdCli {
    DoBack(ctx)
    return
  }

  DoBack(ctx)
}

func DoBack(ctx context.Context) {

	logger.L(ctx).Infof("Program started. Version %s", config.VERSION)
	logger.L(ctx).Info(strings.Join(os.Args, " "))
	logger.L(ctx).Debugf("Reading config %s", config.GetConfigPath())

	if _, err := config.ReadApplyConfig(); err != nil {
		core.ExitError(err)
	}


	logger.L(ctx).Debugf("Initializting Storage")
	if err := storage.Init(); err != nil {
		core.ExitError(err)
	}

	logger.L(ctx).Debugf("Initializting Guardians")
	if err := raida.InitGuardians(); err != nil {
		core.ExitError(err)
	}

	if config.CmdCli {
		logger.L(ctx).Debugf("CLI Mode")

		if flag.NArg() == 0 {
			core.ExitError(perror.New(perror.ERROR_COMMAND_LINE_ARGUMENTS, "Command is missing"))
		}

		command := flag.Arg(0)
		args := flag.Args()

		// Blocking
		monitor.DoMonitorTask(ctx)

		out, err := DoCommand(ctx, command, args[1:])
		if err != nil {
			core.ExitError(err)
		}

		fmt.Println(out)
		os.Exit(0)
	}
}

func DoCommand(ctx context.Context, command string, args []string) (string, error) {
	logger.L(ctx).Debugf("Doing command %s", command)

	var commands = map[string]interface{}{
		"showpayment": &showpayment.Caller{},
		"adminfreeid": &freecoin.Caller{},

	  "testecho": testecho.New(),
	  "testfix": testfix.New(),
	  "testdetect": testdetect.New(),
	  "testsync": testsync.New(),
	}

	caller, ok := commands[command]
	if !ok {
		return "", perror.New(perror.ERROR_COMMAND_NOT_FOUND, "Command not found")
	}

	f := reflect.ValueOf(caller).MethodByName("DoCommand")

	in := make([]reflect.Value, 2)
  in[0] = reflect.ValueOf(ctx)
	in[1] = reflect.ValueOf(args)

	res := f.Call(in)
	if len(res) != 2 {
		return "", perror.New(perror.ERROR_INTERNAL, "Invalid number of arguments returned by the function")
	}

	out := res[0].Interface()
	errIface := res[1].Interface()
	if errIface != nil {
		err := errIface.(error)
		logger.L(ctx).Debugf("Function %s returned error %s", command, err.Error())
		return "", err
	}

	jsonstr, err := json.Marshal(out)
	if err != nil {
		return "", perror.New(perror.ERROR_ENCODE_OUTPUT_JSON, "Failed to encode JSON: " + err.Error())
	}

	return string(jsonstr), nil
}
