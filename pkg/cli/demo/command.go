package demo

import (
	"github.com/fuddle-io/fuddle/pkg/cli/demo/clock"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "demo",
	Short: "run a demo cluster",
}

func init() {
	Command.AddCommand(
		clock.Command,
	)
}
