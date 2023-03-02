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
	"github.com/andydunstall/fuddle/pkg/rpc"
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
	nodes, err := client.Nodes(context.Background())
	if err != nil {
		return err
	}

	displayNodes(nodes)

	return nil
}

func displayNodes(nodes []*rpc.NodeState) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Id < nodes[j].Id
	})

	tbl := table.New("ID", "Service", "Revision")
	for _, node := range nodes {
		tbl.AddRow(node.Id, node.Service, formatRevision(node.Revision))
	}

	tbl.Print()
}

func formatRevision(revision string) string {
	if len(revision) > 7 {
		return revision[:7] + "..."
	}
	return revision
}