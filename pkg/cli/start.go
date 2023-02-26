// Copyright (C) 2023 Andrew Dunstall
//
// Anity is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Anity is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package cli

import (
	"github.com/spf13/cobra"
)

// startCmd starts a fuddle node.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start a fuddle node",
	Long:  "start a fuddle node",
	RunE:  runStart,
}

func runStart(cmd *cobra.Command, args []string) error {
	return nil
}
