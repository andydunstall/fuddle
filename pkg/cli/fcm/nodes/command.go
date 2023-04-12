package nodes

import (
	"github.com/fuddle-io/fuddle/pkg/cli/fcm/nodes/add"
	"github.com/fuddle-io/fuddle/pkg/cli/fcm/nodes/remove"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "nodes",
	Short: "add and remove nodes from a cluster",
}

func init() {
	Command.AddCommand(
		add.Command,
		remove.Command,
	)
}
