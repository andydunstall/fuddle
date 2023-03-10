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

	"github.com/fuddle-io/fuddle/pkg/build"
	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/server"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// bindRegistryAddr is the bind address to listen for registry clients.
	bindRegistryAddr string
	// advRegistryAddr is the address to advertise to registry clients.
	advRegistryAddr string

	// bindAdminAddr is the bind address to listen for admin clients.
	bindAdminAddr string
	// advAdminAddr is the address to advertise to admin clients.
	advAdminAddr string

	// locality is the location of the node in the cluster.
	locality string

	// startVerbose indicates whether debug logs should be enabled on the node.
	startVerbose bool
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
		&bindRegistryAddr,
		"registry-addr", "",
		"0.0.0.0:8220",
		"the bind address to listen for registry clients",
	)
	startCmd.Flags().StringVarP(
		&advRegistryAddr,
		"adv-registry-addr", "",
		"",
		"the address to advertise to registry clients (defaults to the bind address)",
	)

	startCmd.Flags().StringVarP(
		&bindAdminAddr,
		"admin-addr", "",
		"0.0.0.0:8221",
		"the bind address to listen for admin connections",
	)
	startCmd.Flags().StringVarP(
		&advAdminAddr,
		"adv-admin-addr", "",
		"",
		"the address to advertise to admin clients (defaults to the bind address)",
	)

	startCmd.Flags().StringVarP(
		&locality,
		"locality", "l",
		"",
		"the location of the node in the cluster",
	)

	startCmd.PersistentFlags().BoolVarP(
		&startVerbose,
		"verbose", "v",
		false,
		"if set enabled debug logs on the node",
	)
}

func runStart(cmd *cobra.Command, args []string) {
	loggerConf := zap.NewProductionConfig()
	if startVerbose {
		loggerConf.Level.SetLevel(zapcore.DebugLevel)
	}
	logger := zap.Must(loggerConf.Build())

	conf := &config.Config{
		ID: "fuddle-" + uuid.New().String()[:7],

		BindRegistryAddr: bindRegistryAddr,
		AdvRegistryAddr:  bindRegistryAddr,

		BindAdminAddr: bindAdminAddr,
		AdvAdminAddr:  bindAdminAddr,

		Locality: locality,
		Revision: build.Revision,
	}
	if advRegistryAddr != "" {
		conf.AdvRegistryAddr = bindRegistryAddr
	}
	if advAdminAddr != "" {
		conf.AdvAdminAddr = bindAdminAddr
	}

	server := server.NewServer(conf, server.WithLogger(logger))

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
