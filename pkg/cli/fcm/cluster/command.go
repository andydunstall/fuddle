package cluster

import (
	"github.com/fuddle-io/fuddle/pkg/cli/fcm/cluster/create"
	"github.com/fuddle-io/fuddle/pkg/cli/fcm/cluster/health"
	"github.com/fuddle-io/fuddle/pkg/cli/fcm/cluster/info"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "cluster",
	Short: "create, update and delete clusters",
}

func init() {
	Command.AddCommand(
		create.Command,
		info.Command,
		health.Command,
	)
}
