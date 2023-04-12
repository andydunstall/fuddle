package fcm

import (
	"github.com/fuddle-io/fuddle/pkg/cli/fcm/start"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use: "fcm",
}

func init() {
	Command.AddCommand(
		start.Command,
	)
}
