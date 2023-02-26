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

package config

import (
	"go.uber.org/zap/zapcore"
)

// Config contains the node configuration.
type Config struct {
	// BindAddr is the bind address to listen for connections.
	BindAddr string
	// AdvAddr is the address to advertise to clients.
	AdvAddr string
}

func (c *Config) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("bind-addr", c.BindAddr)
	e.AddString("adv-addr", c.AdvAddr)
	return nil
}
