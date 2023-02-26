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
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/andydunstall/fuddle/pkg/client"
	"github.com/spf13/cobra"
)

var (
	demoNodeAddr string
)

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "register demo nodes with the cluster",
	RunE:  runDemo,
}

func init() {
	demoCmd.PersistentFlags().StringVarP(
		&demoNodeAddr,
		"addr", "a",
		"localhost:8220",
		"address of the fuddle server to register with",
	)
}

func runDemo(cmd *cobra.Command, args []string) error {
	registry, err := client.ConnectRegistry(demoNodeAddr)
	if err != nil {
		return err
	}

	// Catch signals so to gracefully shutdown.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	for i := 0; i != 10; i++ {
		nodeID := fmt.Sprintf("node-%d", i)
		if err := registry.Register(context.Background(), nodeID); err != nil {
			return err
		}
		defer func() {
			if err := registry.Unregister(context.Background(), nodeID); err != nil {
				fmt.Println(fmt.Errorf("demo command: %w", err))
			}
		}()
	}

	<-signalCh

	return nil
}
