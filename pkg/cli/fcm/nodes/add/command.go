package add

import (
	"context"
	"fmt"
	"time"

	"github.com/fuddle-io/fuddle/pkg/fcm/client"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "add",
	Short: "add nodes to a cluster",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	client := client.NewClient(addr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	nodesInfo, err := client.AddNodes(ctx, clusterID, fuddleNodes, clientNodes)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("  Fuddle Nodes:")
	for _, n := range nodesInfo.FuddleNodes {
		fmt.Println("      ID:", n.ID)
		fmt.Println("      RPC Addr:", n.RPCAddr)
		fmt.Println("      Admin Addr:", n.AdminAddr)
		fmt.Println("      Log Path:", n.LogPath)
		fmt.Println("")
	}

	fmt.Println("  Client Nodes:")
	for _, n := range nodesInfo.ClientNodes {
		fmt.Println("      ID:", n.ID)
		fmt.Println("      Log Path:", n.LogPath)
		fmt.Println("")
	}
}
