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

package start

var (
	// bindRegistryAddr is the bind address to listen for registry clients.
	bindRegistryAddr string
	// advRegistryAddr is the address to advertise to registry clients.
	advRegistryAddr string

	// locality is the location of the node in the cluster.
	locality string
)

func init() {
	Command.Flags().StringVarP(
		&bindRegistryAddr,
		"registry-addr", "",
		"0.0.0.0:8220",
		"the bind address to listen for registry clients",
	)
	Command.Flags().StringVarP(
		&advRegistryAddr,
		"adv-registry-addr", "",
		"",
		"the address to advertise to registry clients (defaults to the bind address)",
	)

	Command.Flags().StringVarP(
		&locality,
		"locality", "l",
		"",
		"the location of the node in the cluster",
	)
}
