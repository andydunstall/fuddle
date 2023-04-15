package info

import (
	"context"
	"fmt"
	"time"

	"github.com/fuddle-io/fuddle/pkg/fcm/client"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "info",
	Short: "describe an fcm cluster",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	client := client.NewClient(addr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	clusterInfo, err := client.ClusterInfo(ctx, clusterID)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("")
	fmt.Println("  ID:", clusterInfo.ID)
	fmt.Println("")

	fmt.Println("  Fuddle Nodes:")
	for _, n := range clusterInfo.FuddleNodes {
		fmt.Println("      ID:", n.ID)
		fmt.Println("      RPC Addr:", n.RPCAddr)
		fmt.Println("      Admin Addr:", n.AdminAddr)
		fmt.Println("      Log Path:", n.LogPath)
		fmt.Println("")
	}

	fmt.Println("  Client Nodes:")
	for _, n := range clusterInfo.ClientNodes {
		fmt.Println("      ID:", n.ID)
		fmt.Println("      Log Path:", n.LogPath)
		fmt.Println("")
	}
}
