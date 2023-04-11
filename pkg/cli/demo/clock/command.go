package clock

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/fuddle-io/fuddle/demos/clock/pkg/cluster"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "clock",
	Short: "run the clock service demo cluster",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	c, err := cluster.NewCluster()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Shutdown()

	fmt.Println(`
#
# Welcome to the Fuddle clock service demo!
#
# The clock service is a demo cluster that shows how Fuddle can be used to
# register members, lookup membets matching a filter, and subscribe to updates
# when the registry membership changes.
#
# Inspect the cluster with 'fuddle info', specifying '--addr' as the address
# of a Fuddle node.
#`)
	fmt.Println("")

	for _, n := range c.FuddleNodes() {
		fmt.Println("Started fuddle node:")
		fmt.Println("    ID:", n.Config.NodeID)
		fmt.Println("    RPC addr:", n.Config.RPC.JoinAdvAddr())
		fmt.Println("    Log path:", c.LogPath(n.Config.NodeID))
		fmt.Printf("    Inspect: fuddle info cluster --addr %s\n", n.Config.RPC.JoinAdvAddr())
		fmt.Println("")
	}
	for _, n := range c.ClockNodes() {
		fmt.Println("Started clock node:")
		fmt.Println("    ID:", n.ID)
		fmt.Println("    Log path:", c.LogPath(n.ID))
		fmt.Println("")
	}
	for _, n := range c.FrontendNodes() {
		fmt.Println("Started frontend node:")
		fmt.Println("    ID:", n.ID)
		fmt.Println("    Addr:", n.Addr)
		fmt.Println("    Log path:", c.LogPath(n.ID))
		fmt.Printf("    Request: curl http://%s/time\n", n.Addr)
		fmt.Println("")
	}

	<-signalCh
}
