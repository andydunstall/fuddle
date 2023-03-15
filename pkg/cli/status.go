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
	"sort"

	"github.com/fuddle-io/fuddle/pkg/client"
	"github.com/fuddle-io/fuddle/pkg/registry/cluster"
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

var statusNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "inspect the status of a node",
	Long: `
Inspect the status of a node.
`,
	RunE: runNodeStatus,
}

func init() {
	statusCmd.AddCommand(
		statusClusterCmd,
		statusNodeCmd,
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

func runNodeStatus(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing node ID")
	}

	id := args[0]

	client := client.NewAdmin(statusAdminAddr)
	node, err := client.Node(context.Background(), id)
	if err != nil {
		return err
	}

	fmt.Println("ID:", node.ID)
	fmt.Println("Service:", node.Service)
	fmt.Println("Locality:", node.Locality)
	fmt.Println("Created:", node.Created)
	fmt.Println("Revision:", node.Revision)
	fmt.Println("Metadata:")

	keys := []string{}
	for key := range node.Metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("    %s: %s\n", key, node.Metadata[key])
	}

	return nil
}

func displayNodes(nodes []*cluster.Node) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})

	tbl := table.New("ID", "Service", "Locality", "Created", "Revision")
	for _, node := range nodes {
		tbl.AddRow(
			node.ID,
			node.Service,
			node.Locality,
			node.Created,
			formatRevision(node.Revision),
		)
	}

	tbl.Print()
}

func formatRevision(revision string) string {
	if len(revision) > 10 {
		return revision[:10] + "..."
	}
	return revision
}
