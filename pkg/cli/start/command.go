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

package start

import (
	"os"
	"os/signal"

	"github.com/fuddle-io/fuddle/pkg/build"
	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/fuddle"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
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
	logger := zap.Must(loggerConf.Build())

	conf := &config.Config{
		ID: "fuddle-" + uuid.New().String()[:7],

		BindRegistryAddr: bindRegistryAddr,
		AdvRegistryAddr:  bindRegistryAddr,

		Locality: locality,
		Revision: build.Revision,
	}
	if advRegistryAddr != "" {
		conf.AdvRegistryAddr = bindRegistryAddr
	}

	server := fuddle.New(conf, fuddle.WithLogger(logger))

	// Catch signals so to gracefully shutdown the server.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	if err := server.Start(); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
	defer server.GracefulStop()

	sig := <-signalCh
	logger.Info("received exit signal", zap.String("signal", sig.String()))
}
