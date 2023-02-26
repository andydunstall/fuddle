// Copyright (C) 2023 Andrew Dunstall
//
// Fuddle is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Fuddle is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package cli

import (
	"os"
	"os/signal"

	"github.com/andydunstall/fuddle/pkg/config"
	"github.com/andydunstall/fuddle/pkg/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	// bindAddr is the address the server should bind to.
	bindAddr string
	// advAddr is the address the server should advertise to clients.
	advAddr string
)

// startCmd starts a fuddle node.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start a fuddle node",
	Long:  "start a fuddle node",
	Run:   runStart,
}

func init() {
	startCmd.Flags().StringVarP(
		&bindAddr,
		"addr", "a",
		"0.0.0.0:8220",
		"the bind address to listen for connections",
	)
	startCmd.Flags().StringVarP(
		&advAddr,
		"adv-addr", "",
		"",
		"the address to advertise to clients (defaults to the bind address)",
	)
}

func runStart(cmd *cobra.Command, args []string) {
	logger, _ := zap.NewProduction()
	conf := &config.Config{
		BindAddr: bindAddr,
		AdvAddr:  bindAddr,
	}
	if advAddr != "" {
		conf.AdvAddr = bindAddr
	}

	server := server.NewServer(conf, logger)

	// Catch signals so to gracefully shutdown the server.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	if err := server.Start(); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
	defer server.GracefulShutdown()

	sig := <-signalCh
	logger.Info("received exit signal", zap.String("signal", sig.String()))
}
