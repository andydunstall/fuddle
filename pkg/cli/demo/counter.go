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
	"net"
	"os"
	"os/signal"

	"github.com/fuddle-io/fuddle/demos/counter/pkg/service/counter"
	"github.com/fuddle-io/fuddle/demos/counter/pkg/service/frontend"
	"github.com/fuddle-io/fuddle/pkg/build"
	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/server"
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
	fmt.Println(`#
# Welcome to the Fuddle counter service demo!
#
# This demo shows how Fuddle can be used to observe the nodes in a cluster,
# discover nodes, and route request using application specific routing, including
# consistent hashing and load balancing with a custom policy.
#
# View the cluster dashboard at http://127.0.0.1:8221.
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
		server.WithLogger(demoLogger(logDir, fuddleConf.ID)),
	)
	if err := fuddleNode.Start(); err != nil {
		return fmt.Errorf("counter service: fuddle node: %w", err)
	}
	defer fuddleNode.GracefulStop()

	var frontendNodes []*frontend.Config
	frontendNodes = append(frontendNodes, frontendNodeConfig("us-east-1-a"))
	frontendNodes = append(frontendNodes, frontendNodeConfig("us-east-1-b"))
	frontendNodes = append(frontendNodes, frontendNodeConfig("us-east-1-c"))

	for _, conf := range frontendNodes {
		node := frontend.NewService(
			conf,
			frontend.WithLogger(demoLogger(logDir, fuddleConf.ID)),
		)
		if err := node.Start(); err != nil {
			return fmt.Errorf("counter service: frontend node: %w", err)
		}
		defer node.GracefulStop()
	}

	var counterNodes []*counter.Config
	counterNodes = append(counterNodes, counterNodeConfig("us-east-1-a"))
	counterNodes = append(counterNodes, counterNodeConfig("us-east-1-b"))
	counterNodes = append(counterNodes, counterNodeConfig("us-east-1-c"))

	for _, conf := range counterNodes {
		node := counter.NewService(
			conf, counter.WithLogger(demoLogger(logDir, fuddleConf.ID)),
		)
		if err := node.Start(); err != nil {
			return fmt.Errorf("counter service: counter node: %w", err)
		}
		defer node.GracefulStop()
	}

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
#     -------
#`)

	for _, conf := range frontendNodes {
		fmt.Printf(`#     %s
#       Endpoint: ws://%s/{id}
#       Locality: %s
#       Logs: %s
#`, conf.ID, conf.WSAddr, conf.Locality, demoLogPath(logDir, conf.ID))
		fmt.Println("")
	}

	fmt.Println(`#     Counter
#     -------
#`)

	for _, conf := range counterNodes {
		fmt.Printf(`#     %s
#       Locality: %s
#       Logs: %s
#`, conf.ID, conf.Locality, demoLogPath(logDir, conf.ID))
		fmt.Println("")
	}

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

func frontendNodeConfig(locality string) *frontend.Config {
	return &frontend.Config{
		ID:          "frontend-" + uuid.New().String()[:8],
		WSAddr:      GetSystemAddress(),
		FuddleAddrs: []string{"127.0.0.1:8220"},
		Locality:    locality,
		Revision:    build.Revision,
	}
}

func counterNodeConfig(locality string) *counter.Config {
	return &counter.Config{
		ID:          "counter-" + uuid.New().String()[:8],
		RPCAddr:     GetSystemAddress(),
		FuddleAddrs: []string{"127.0.0.1:8220"},
		Locality:    locality,
		Revision:    build.Revision,
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

func GetSystemAddress() string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	return ln.Addr().String()
}
