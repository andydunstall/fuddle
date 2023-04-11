package start

import (
	"os"
	"os/signal"
	"strings"

	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/fuddle"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Command starts a fuddle node.
var Command = &cobra.Command{
	Use:   "start",
	Short: "start a fuddle node",
	Long:  "start a fuddle node",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	loggerConf := zap.NewProductionConfig()
	switch logLevel {
	case "debug":
		loggerConf.Level.SetLevel(zapcore.DebugLevel)
	case "info":
		loggerConf.Level.SetLevel(zapcore.InfoLevel)
	case "warn":
		loggerConf.Level.SetLevel(zapcore.WarnLevel)
	case "error":
		loggerConf.Level.SetLevel(zapcore.ErrorLevel)
	default:
		// If the level is invalid or not specified, use info.
		loggerConf.Level.SetLevel(zapcore.InfoLevel)
	}
	logger := zap.Must(loggerConf.Build())

	conf := config.DefaultConfig()

	conf.Gossip.BindAddr = gossipBindAddr
	conf.Gossip.BindPort = gossipBindPort
	if gossipAdvAddr != "" {
		conf.Gossip.AdvAddr = gossipAdvAddr
	} else {
		conf.Gossip.AdvAddr = gossipBindAddr
	}
	if gossipAdvPort != 0 {
		conf.Gossip.AdvPort = gossipAdvPort
	} else {
		conf.Gossip.AdvPort = gossipBindPort
	}

	if gossipSeeds != "" {
		conf.Gossip.Seeds = strings.Split(gossipSeeds, ",")
	}

	conf.RPC.BindAddr = rpcBindAddr
	conf.RPC.BindPort = rpcBindPort
	if rpcAdvAddr != "" {
		conf.RPC.AdvAddr = rpcAdvAddr
	} else {
		conf.RPC.AdvAddr = rpcBindAddr
	}
	if rpcAdvPort != 0 {
		conf.RPC.AdvPort = rpcAdvPort
	} else {
		conf.RPC.AdvPort = rpcBindPort
	}

	conf.Admin.BindAddr = adminBindAddr
	conf.Admin.BindPort = adminBindPort
	if adminAdvAddr != "" {
		conf.Admin.AdvAddr = adminAdvAddr
	} else {
		conf.Admin.AdvAddr = adminBindAddr
	}
	if adminAdvPort != 0 {
		conf.Admin.AdvPort = adminAdvPort
	} else {
		conf.Admin.AdvPort = adminBindPort
	}

	// Catch signals so to gracefully shutdown the server.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	server, err := fuddle.NewFuddle(conf, fuddle.WithLogger(logger))
	if err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
	defer server.Shutdown()

	sig := <-signalCh
	logger.Info("received exit signal", zap.String("signal", sig.String()))
}
