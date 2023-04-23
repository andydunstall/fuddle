package info

import (
	"context"
	"fmt"
	"sort"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	admin "github.com/fuddle-io/fuddle/pkg/admin/client"
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
	client, err := admin.Connect(addr)
	if err != nil {
		return err
	}
	members, err := client.Members(context.Background())
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

	client, err := admin.Connect(addr)
	if err != nil {
		return err
	}
	member, err := client.Member(context.Background(), id)
	if err != nil {
		return err
	}

	fmt.Println("ID:", member.State.Id)
	fmt.Println("Status:", member.State.Status)
	fmt.Println("Service:", member.State.Service)
	fmt.Println("Locality:")
	fmt.Println("    Region:", member.State.Locality.Region)
	fmt.Println("    Availability Zone:", member.State.Locality.AvailabilityZone)
	fmt.Println("Started:", member.State.Started)
	fmt.Println("Revision:", member.State.Revision)
	fmt.Println("Metadata:")

	keys := []string{}
	for key := range member.State.Metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("    %s: %s\n", key, member.State.Metadata[key])
	}

	return nil
}

func displayMembers(members []*rpc.Member2) {
	sort.Slice(members, func(i, j int) bool {
		return members[i].State.Id < members[j].State.Id
	})

	tbl := table.New("ID", "Status", "Service", "Locality", "Created", "Revision")
	for _, member := range members {
		availabilityZone := ""
		if member.State.Locality != nil {
			availabilityZone = member.State.Locality.AvailabilityZone
		}
		tbl.AddRow(
			member.State.Id,
			member.State.Status,
			member.State.Service,
			availabilityZone,
			member.State.Started,
			formatRevision(member.State.Revision),
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
