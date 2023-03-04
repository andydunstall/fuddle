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
	"fmt"
	"os"
	"os/signal"

	"github.com/andydunstall/fuddle/demos/random/pkg/frontend"
	"github.com/andydunstall/fuddle/demos/random/pkg/random"
	"github.com/andydunstall/fuddle/pkg/config"
	"github.com/andydunstall/fuddle/pkg/server"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type demoFuddleConfig struct {
	ID      string
	LogPath string
	Config  *config.Config
}

type demoFrontendConfig struct {
	Config  *frontend.Config
	LogPath string
}

type demoRandomConfig struct {
	Config  *random.Config
	LogPath string
}

var randomDemoCmd = &cobra.Command{
	Use:   "random",
	Short: "run a demo 'is even' service cluster",
	RunE:  runRandomDemo,
}

func runRandomDemo(cmd *cobra.Command, args []string) error {
	conf := &config.Config{
		BindAddr: "127.0.0.1:8220",
		AdvAddr:  "127.0.0.1:8220",

		BindAdminAddr: "127.0.0.1:8221",
		AdvAdminAddr:  "127.0.0.1:8221",
	}

	logDir, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("is even service: %w", err)
	}

	fuddleConfig := []*demoFuddleConfig{}
	for i := 0; i != 1; i++ {
		conf, err := demoFuddleNode(logDir)
		if err != nil {
			return fmt.Errorf("fuddle service: %w", err)
		}
		fuddleConfig = append(fuddleConfig, conf)
	}

	frontendConfig := []*demoFrontendConfig{}
	for i := 0; i != 3; i++ {
		conf, err := demoFrontendNode(logDir)
		if err != nil {
			return fmt.Errorf("is even service: %w", err)
		}
		frontendConfig = append(frontendConfig, conf)
	}

	randomConfig := []*demoRandomConfig{}
	for i := 0; i != 3; i++ {
		conf, err := demoRandomNode(logDir)
		if err != nil {
			return fmt.Errorf("is even service: %w", err)
		}
		randomConfig = append(randomConfig, conf)
	}

	// Catch signals so to gracefully shutdown the server.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	for _, conf := range fuddleConfig {
		server := server.NewServer(conf.Config, loggerWithPath(conf.LogPath, false))
		fmt.Println("START FUDDLE")
		if err := server.Start(); err != nil {
			return fmt.Errorf("failed to start fuddle: %w", err)
		}
		fmt.Println("START FUDDLE OK")
		defer server.GracefulStop()
	}

	for _, conf := range frontendConfig {
		service := frontend.NewService(conf.Config, loggerWithPath(conf.LogPath, false))
		fmt.Println("START FR")
		if err := service.Start(); err != nil {
			return fmt.Errorf("failed to start frontend: %w", err)
		}
		fmt.Println("START FR OK")
		defer service.GracefulStop()
	}

	for _, conf := range randomConfig {
		service := random.NewService(conf.Config, loggerWithPath(conf.LogPath, false))
		if err := service.Start(); err != nil {
			return fmt.Errorf("failed to start random: %w", err)
		}
		defer service.GracefulStop()
	}

	fmt.Printf(`
#
# Welcome to the Fuddle random number service demo!
#
# The is a toy example that uses Fuddle for cluster management.
#
# View the cluster dashboard at http://%s."
#
# Or inspect the cluster with 'fuddle status cluster'.
#
#   Nodes
#   -----
#`, conf.AdvAdminAddr)

	for _, conf := range fuddleConfig {
		fmt.Printf(`
#   fuddle: %s
#     Admin Dashboard: %s
#     Logs: %s
#`, conf.ID, "http://"+conf.Config.AdvAdminAddr, conf.LogPath)
	}

	for _, conf := range frontendConfig {
		fmt.Printf(`
#   frontend: %s
#     Endpoint: http://%s/random
#     Logs: %s
#`, conf.Config.ID, conf.Config.Addr, conf.LogPath)
	}

	for _, conf := range randomConfig {
		fmt.Printf(`
#   random: %s
#     Logs: %s
#`, conf.Config.ID, conf.LogPath)
	}
	fmt.Println("")

	<-signalCh

	return nil
}

func demoFuddleNode(logDir string) (*demoFuddleConfig, error) {
	conf := &config.Config{
		BindAddr: "127.0.0.1:8220",
		AdvAddr:  "127.0.0.1:8220",

		BindAdminAddr: "127.0.0.1:8221",
		AdvAdminAddr:  "127.0.0.1:8221",
	}

	fuddleNodeID := "fuddle-" + uuid.New().String()[:7]
	fuddleNodeLogPath := logDir + "/" + fuddleNodeID + ".log"

	return &demoFuddleConfig{
		ID:      fuddleNodeID,
		LogPath: fuddleNodeLogPath,
		Config:  conf,
	}, nil
}

func demoFrontendNode(logDir string) (*demoFrontendConfig, error) {
	id := "frontend-" + uuid.New().String()[:7]
	logPath := logDir + "/" + id + ".log"

	return &demoFrontendConfig{
		Config: &frontend.Config{
			ID:   id,
			Addr: getSystemAddress(),
		},
		LogPath: logPath,
	}, nil
}

func demoRandomNode(logDir string) (*demoRandomConfig, error) {
	id := "random-" + uuid.New().String()[:7]
	logPath := logDir + "/" + id + ".log"

	return &demoRandomConfig{
		Config: &random.Config{
			ID:   id,
			Addr: getSystemAddress(),
		},
		LogPath: logPath,
	}, nil
}
