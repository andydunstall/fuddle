package cli

import (
	"github.com/fuddle-io/fuddle/pkg/cli/info"
	"github.com/fuddle-io/fuddle/pkg/cli/start"
	"github.com/spf13/cobra"
)

// fuddleCmd is the root command to run fuddle.
var fuddleCmd = &cobra.Command{
	Use:          "fuddle [command] (flags)",
	Short:        "fuddle cli and server",
	Long:         "fuddle cli and server",
	SilenceUsage: true,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func init() {
	cobra.EnableCommandSorting = false

	fuddleCmd.AddCommand(
		start.Command,
		info.Command,
	)
}

// Start starts the CLI.
func Start() error {
	return fuddleCmd.Execute()
}
