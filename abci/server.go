package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	cmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"

	"github.com/tendermint/tendermint/abci/types"

	abciApp "abci-example/abci/app"
	"abci-example/utils"
)

type loggerWriter struct{}

const (
	fileDatetimeFormat = "02-01-2006_15-04-05"
	logTargetConsole   = "console"
	logTargetFile      = "file"
)

func init() {
	// Set default logrus
	var logLevel = utils.GetEnv("ABCI_LOG_LEVEL", "debug")
	var logTarget = utils.GetEnv("ABCI_LOG_TARGET", logTargetConsole)
	currentTime := time.Now()
	currentTimeStr := currentTime.Format(fileDatetimeFormat)
	var logFilePath = utils.GetEnv("ABCI_LOG_FILE_PATH", "./abci-"+strconv.Itoa(os.Getpid())+"-"+currentTimeStr+".log")
	if logTarget == logTargetConsole {
		logrus.SetOutput(os.Stdout)
	} else if logTarget == logTargetFile {
		logFile, _ := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		logrus.SetOutput(logFile)
	} else {
		panic(fmt.Errorf("Unknown log target: \"%s\". Only \"console\" and \"file\" are allowed", logTarget))
	}
	switch logLevel {
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	default:
		logrus.SetLevel(logrus.DebugLevel)
	}
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
}

// Ref: github.com/tendermint/tendermint/cmd/tendermint/main.go
func main() {
	rootCmd := cmd.RootCmd
	rootCmd.AddCommand(
		cmd.GenValidatorCmd,
		cmd.InitFilesCmd,
		cmd.ProbeUpnpCmd,
		cmd.LiteCmd,
		cmd.ReplayCmd,
		cmd.ReplayConsoleCmd,
		cmd.ResetAllCmd,
		cmd.ResetPrivValidatorCmd,
		cmd.ShowValidatorCmd,
		cmd.TestnetFilesCmd,
		cmd.ShowNodeIDCmd,
		cmd.GenNodeKeyCmd,
		cmd.VersionCmd)
	// NOTE:
	// Users wishing to:
	//	* Use an external signer for their validators
	//	* Supply an in-proc abci app
	//	* Supply a genesis doc file from another source
	//	* Provide their own DB implementation
	// can copy this file and use something other than the
	// DefaultNewNode function
	nodeFunc := newNode
	// Create & start node
	rootCmd.AddCommand(cmd.NewRunNodeCmd(nodeFunc))
	cmd := cli.PrepareBaseCmd(rootCmd, "TM", os.ExpandEnv(filepath.Join("$HOME", cfg.DefaultTendermintDir)))
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

// Ref: github.com/tendermint/tendermint/node/node.go (func DefaultNewNode)
func newNode(config *cfg.Config, logger log.Logger) (*nm.Node, error) {
	var app types.Application
	app = abciApp.NewApplicationInterface()
	// Generate node PrivKey
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, err
	}
	// Convert old PrivValidator if it exists.
	oldPrivVal := config.OldPrivValidatorFile()
	newPrivValKey := config.PrivValidatorKeyFile()
	newPrivValState := config.PrivValidatorStateFile()
	if _, err := os.Stat(oldPrivVal); !os.IsNotExist(err) {
		oldPV, err := privval.LoadOldFilePV(oldPrivVal)
		if err != nil {
			return nil, fmt.Errorf("Error reading OldPrivValidator from %v: %v\n", oldPrivVal, err)
		}
		logger.Info("Upgrading PrivValidator file",
			"old", oldPrivVal,
			"newKey", newPrivValKey,
			"newState", newPrivValState,
		)
		oldPV.Upgrade(newPrivValKey, newPrivValState)
	}
	return nm.NewNode(config,
		privval.LoadOrGenFilePV(newPrivValKey, newPrivValState),
		nodeKey,
		proxy.NewLocalClientCreator(app),
		nm.DefaultGenesisDocProviderFunc(config),
		nm.DefaultDBProvider,
		nm.DefaultMetricsProvider(config.Instrumentation),
		logger.With("module", "node"),
	)
}
