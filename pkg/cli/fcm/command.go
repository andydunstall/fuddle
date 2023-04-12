package fcm

import (
	"github.com/fuddle-io/fuddle/pkg/cli/fcm/cluster"
	"github.com/fuddle-io/fuddle/pkg/cli/fcm/nodes"
	"github.com/fuddle-io/fuddle/pkg/cli/fcm/start"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "fcm",
	Short: "fcm is a tool for spinning up local fuddle clusters",
}

func init() {
	Command.AddCommand(
		start.Command,
		cluster.Command,
		nodes.Command,
	)
}
