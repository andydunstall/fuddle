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

package info

import (
	"context"
	"fmt"
	"sort"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/client"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "info",
	Short: "inspect the status of the cluster",
}

var clusterCommand = &cobra.Command{
	Use:   "cluster",
	Short: "inspect the status of the cluster",
	Long: `
Inspect the status of the cluster.

Displays an overview of the cluster status and a list of members in the cluster.
`,
	RunE: runClusterStatus,
}

var memberCommand = &cobra.Command{
	Use:   "member",
	Short: "inspect the status of a member",
	Long: `
Inspect the status of a member.
`,
	RunE: runMemberStatus,
}

func init() {
	Command.AddCommand(
		clusterCommand,
		memberCommand,
	)
}

func runClusterStatus(cmd *cobra.Command, args []string) error {
	client, err := client.NewAdmin(addr)
	if err != nil {
		return err
	}
	members, err := client.Cluster(context.Background())
	if err != nil {
		return err
	}

	displayMembers(members)

	return nil
}

func runMemberStatus(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing member ID")
	}

	id := args[0]

	client, err := client.NewAdmin(addr)
	if err != nil {
		return err
	}
	member, err := client.Member(context.Background(), id)
	if err != nil {
		return err
	}

	fmt.Println("ID:", member.Id)
	fmt.Println("Service:", member.Service)
	fmt.Println("Locality:", member.Locality)
	fmt.Println("Created:", member.Created)
	fmt.Println("Revision:", member.Revision)
	fmt.Println("Metadata:")

	keys := []string{}
	for key := range member.Metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("    %s: %s\n", key, member.Metadata[key])
	}

	return nil
}

func displayMembers(members []*rpc.Member) {
	sort.Slice(members, func(i, j int) bool {
		return members[i].Id < members[j].Id
	})

	tbl := table.New("ID", "Status", "Service", "Locality", "Created", "Revision")
	for _, member := range members {
		tbl.AddRow(
			member.Id,
			member.Status,
			member.Service,
			member.Locality,
			member.Created,
			formatRevision(member.Revision),
		)
	}

	tbl.Print()
}

func formatRevision(revision string) string {
	if len(revision) > 25 {
		return revision[:25] + "..."
	}
	return revision
}
