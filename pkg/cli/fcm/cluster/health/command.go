package health

import (
	"context"
	"fmt"
	"time"

	"github.com/fuddle-io/fuddle/pkg/fcm/client"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "health",
	Short: "describe an fcm clusters health",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	client := client.NewClient(addr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	healthy, err := client.ClusterHealth(ctx, clusterID)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Healthy:", healthy)
}
