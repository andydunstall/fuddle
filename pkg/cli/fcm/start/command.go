package start

import (
	"os"
	"os/signal"

	"github.com/fuddle-io/fuddle/pkg/fcm"
	"github.com/fuddle-io/fuddle/pkg/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var Command = &cobra.Command{
	Use:   "start",
	Short: "start the fcm server",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	loggerConf := zap.NewProductionConfig()
	loggerConf.Level.SetLevel(logger.StringToLevel(logLevel))
	logger := zap.Must(loggerConf.Build())

	// Catch signals so to gracefully shutdown the server.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	server, err := fcm.NewFCM(
		addr,
		port,
		fcm.WithDefaultCluster(cluster),
		fcm.WithLogger(logger),
	)
	if err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
	defer server.Shutdown()

	sig := <-signalCh
	logger.Info("received exit signal", zap.String("signal", sig.String()))
}
