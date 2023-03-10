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

package demo

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/andydunstall/fuddle/pkg/build"
	"github.com/andydunstall/fuddle/pkg/config"
	"github.com/andydunstall/fuddle/pkg/server"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// verbose indicates whether debug logs should be enabled.
	verbose bool
)

var CounterCmd = &cobra.Command{
	Use:   "counter",
	Short: "run a demo counter service cluster",
	Long: `Run the counter service demo.

This demo shows how Fuddle can be used to observe the nodes in a cluster,
discover nodes, and route request using application specific routing, including
consistent hashing and load balancing with a custom policy.
`,
	RunE: runCounterService,
}

func init() {
	CounterCmd.PersistentFlags().BoolVarP(
		&verbose,
		"verbose", "v",
		false,
		"if set enabled debug logs on the node",
	)
}

func runCounterService(cmd *cobra.Command, args []string) error {
	fmt.Println(`
#
# Welcome to the Fuddle counter service demo!
#
# This demo shows how Fuddle can be used to observe the nodes in a cluster,
# discover nodes, and route request using application specific routing, including
# consistent hashing and load balancing with a custom policy.
#
# View the cluster dashboard at http://127.0.0.1:8221."
#
# Or inspect the cluster with 'fuddle status cluster'.
#`)

	logDir, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("counter service: create log dir: %w", err)
	}

	fuddleConf := fuddleNodeConfig()

	fuddleNode := server.NewServer(
		fuddleConf,
		demoLogger(logDir, fuddleConf.ID),
	)
	if err := fuddleNode.Start(); err != nil {
		return fmt.Errorf("counter service: fuddle node: %w", err)
	}
	defer fuddleNode.GracefulStop()

	// Catch signals to gracefully shutdown the server.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	fmt.Println(`#   Nodes
#   -----
#`)

	fmt.Println(`#     Fuddle
#     -----
#`)

	fmt.Printf(`#     %s
#       Admin Dashboard: http://%s
#       Locality: %s
#       Logs: %s
#`, fuddleConf.ID, fuddleConf.AdvAdminAddr, fuddleConf.Locality, demoLogPath(logDir, fuddleConf.ID))
	fmt.Println("")

	fmt.Println(`#     Frontend
#     --------
#
#     frontend-9cd2c9e
#       Endpoint: http://127.0.0.1:61564
#       Locality: us-east-1-a
#       Logs: /var/folders/_z/p6j4xhdd1kn1xwct176qj3bc0000gp/T/2129500756/frontend-9cd2c9e.log
#
#     frontend: frontend-4ffda82
#       Endpoint: http://127.0.0.1:61565
#       Locality: us-east-1-b
#       Logs: /var/folders/_z/p6j4xhdd1kn1xwct176qj3bc0000gp/T/2129500756/frontend-4ffda82.log
#
#     frontend: frontend-4eb03b7
#       Endpoint: http://127.0.0.1:61566
#       Locality: us-east-1-c
#       Logs: /var/folders/_z/p6j4xhdd1kn1xwct176qj3bc0000gp/T/2129500756/frontend-4eb03b7.log
#
#     Counter
#     -------
#
#     counter: counter-eeb9488
#       Locality: us-east-1-a
#       Logs: /var/folders/_z/p6j4xhdd1kn1xwct176qj3bc0000gp/T/2129500756/counter-eeb9488.log
#
#     counter: counter-57cbaef
#       Locality: us-east-1-b
#       Logs: /var/folders/_z/p6j4xhdd1kn1xwct176qj3bc0000gp/T/2129500756/counter-57cbaef.log
#
#     counter: counter-ce4f2fa
#       Locality: us-east-1-c
#       Logs: /var/folders/_z/p6j4xhdd1kn1xwct176qj3bc0000gp/T/2129500756/counter-ce4f2fa.log
#
#     Clock
#     -----
#
#     clock: clock-eeb9488
#       Locality: us-east-1-a
#       Logs: /var/folders/_z/p6j4xhdd1kn1xwct176qj3bc0000gp/T/2129500756/clock-eeb9488.log
#
#     clock: clock-57cbaef
#       Locality: us-east-1-b
#       Logs: /var/folders/_z/p6j4xhdd1kn1xwct176qj3bc0000gp/T/2129500756/clock-57cbaef.log
#
#     clock: clock-ce4f2fa
#       Locality: us-east-1-c
#       Logs: /var/folders/_z/p6j4xhdd1kn1xwct176qj3bc0000gp/T/2129500756/clock-ce4f2fa.log
#`)

	<-signalCh

	return nil
}

func fuddleNodeConfig() *config.Config {
	// Hardcode the fuddle addresses so we can document the dashboard URL.
	return &config.Config{
		ID:            "fuddle-" + uuid.New().String()[:8],
		BindAddr:      "127.0.0.1:8220",
		AdvAddr:       "127.0.0.1:8220",
		BindAdminAddr: "127.0.0.1:8221",
		AdvAdminAddr:  "127.0.0.1:8221",
		Locality:      "us-east-1-a",
		Revision:      build.Revision,
	}
}

func demoLogger(dir string, id string) *zap.Logger {
	path := dir + "/" + id + ".log"
	loggerConf := zap.NewProductionConfig()
	if verbose {
		loggerConf.Level.SetLevel(zapcore.DebugLevel)
	}
	loggerConf.OutputPaths = []string{path}
	return zap.Must(loggerConf.Build())
}

func demoLogPath(dir string, id string) string {
	return dir + "/" + id + ".log"
}
