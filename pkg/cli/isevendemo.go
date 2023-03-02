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

	"github.com/andydunstall/fuddle/pkg/config"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type demoFuddleConfig struct {
	ID      string
	LogPath string
	Config  *config.Config
}

type demoFrontendConfig struct {
	ID      string
	Addr    string
	LogPath string
}

type demoIsEvenConfig struct {
	ID      string
	Addr    string
	LogPath string
}

var isEvenDemoCmd = &cobra.Command{
	Use:   "iseven",
	Short: "run a demo 'is even' service cluster",
	RunE:  runIsEvenDemo,
}

func runIsEvenDemo(cmd *cobra.Command, args []string) error {
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

	fuddleConfig, err := demoFuddleNode(logDir)
	if err != nil {
		return fmt.Errorf("is even service: %w", err)
	}

	frontendConfig := []*demoFrontendConfig{}
	for i := 0; i != 3; i++ {
		conf, err := demoFrontendNode(logDir)
		if err != nil {
			return fmt.Errorf("is even service: %w", err)
		}
		frontendConfig = append(frontendConfig, conf)
	}

	isEvenConfig := []*demoIsEvenConfig{}
	for i := 0; i != 3; i++ {
		conf, err := demoIsEvenNode(logDir)
		if err != nil {
			return fmt.Errorf("is even service: %w", err)
		}
		isEvenConfig = append(isEvenConfig, conf)
	}

	fmt.Printf(`
#
# Welcome to the Fuddle ‘Is-Even’ service demo!
#
# 'Is-Even' is a toy example that uses Fuddle for cluster management.
#
# View the cluster dashboard at http://%s."
#
# Or inspect the cluster with 'fuddle status cluster'.
#
#   Nodes
#   -----
#`, conf.AdvAdminAddr)

	fmt.Printf(`
#   fuddle: %s
#     Admin Dashboard: %s
#     Logs: %s
#`, fuddleConfig.ID, "http://"+fuddleConfig.Config.AdvAdminAddr, fuddleConfig.LogPath)

	for _, conf := range frontendConfig {
		fmt.Printf(`
#   frontend: %s
#     Endpoint: http://%s/iseven?n=10
#     Logs: %s
#`, conf.ID, conf.Addr, conf.LogPath)
	}

	for _, conf := range isEvenConfig {
		fmt.Printf(`
#   iseven: %s
#     Endpoint: http://%s/iseven?n=10
#     Logs: %s
#`, conf.ID, conf.Addr, conf.LogPath)
	}

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
		ID:      id,
		Addr:    getSystemAddress(),
		LogPath: logPath,
	}, nil
}

func demoIsEvenNode(logDir string) (*demoIsEvenConfig, error) {
	id := "iseven-" + uuid.New().String()[:7]
	logPath := logDir + "/" + id + ".log"

	return &demoIsEvenConfig{
		ID:      id,
		Addr:    getSystemAddress(),
		LogPath: logPath,
	}, nil
}
