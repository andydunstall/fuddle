package cluster

import (
	"github.com/fuddle-io/fuddle/pkg/cli/fcm/cluster/create"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "cluster",
	Short: "create, update and delete clusters",
}

func init() {
	Command.AddCommand(
		create.Command,
	)
}
