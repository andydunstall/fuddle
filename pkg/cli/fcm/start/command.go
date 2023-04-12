package start

import (
	"os"
	"os/signal"

	"github.com/fuddle-io/fuddle/pkg/fcm"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Command = &cobra.Command{
	Use: "start",
	Run: run,
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

	// Catch signals so to gracefully shutdown the server.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	server, err := fcm.NewFCM(addr, port, fcm.WithLogger(logger))
	if err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
	defer server.Shutdown()

	sig := <-signalCh
	logger.Info("received exit signal", zap.String("signal", sig.String()))
}
