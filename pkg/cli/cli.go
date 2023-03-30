// Copyright (C) 2023 Andrew Dunstall
//
// Fuddle is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Fuddle is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
		demoCmd,
		versionCmd,
	)
}

// Start starts the CLI.
func Start() error {
	return fuddleCmd.Execute()
}
