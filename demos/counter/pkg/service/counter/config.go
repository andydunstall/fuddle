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

package counter

import (
	"strings"

	"go.uber.org/zap/zapcore"
)

// Config contains the node configuration.
type Config struct {
	// ID is a unique identifier for the node.
	ID string

	// RPCAddr is the address to listen for RPC connections.
	RPCAddr string

	// FuddleAddrs contains fuddle registry seed nodes.
	FuddleAddrs []string

	// Locality is the location of the node in the cluster.
	Locality string

	// Revision is the build commit.
	Revision string
}

func (c *Config) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("id", c.ID)
	e.AddString("rpc-addr", c.RPCAddr)
	e.AddString("fuddle-addrs", strings.Join(c.FuddleAddrs, ","))
	e.AddString("locality", c.Locality)
	e.AddString("revision", c.Revision)
	return nil
}
