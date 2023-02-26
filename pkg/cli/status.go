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
	"sort"

	"github.com/andydunstall/fuddle/pkg/client"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var (
	statusAdminAddr string
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "inspect the status of the cluster",
}

var statusClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "inspect the status of the cluster",
	Long: `
Inspect the status of the cluster.

Displays an overview of the cluster status and a list of nodes in the cluster.
`,
	RunE: runClusterStatus,
}

func init() {
	statusCmd.AddCommand(
		statusClusterCmd,
	)

	statusCmd.PersistentFlags().StringVarP(
		&statusAdminAddr,
		"addr", "a",
		"localhost:8221",
		"address of the admin server to query",
	)
}

func runClusterStatus(cmd *cobra.Command, args []string) error {
	client := client.NewAdmin(statusAdminAddr)
	nodeIDs, err := client.Nodes(context.Background())
	if err != nil {
		return err
	}

	displayNodes(nodeIDs)

	return nil
}

func displayNodes(nodeIDs []string) {
	sort.Strings(nodeIDs)

	tbl := table.New("ID")
	for _, id := range nodeIDs {
		tbl.AddRow(id)
	}

	tbl.Print()
}
