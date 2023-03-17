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
	"github.com/fuddle-io/fuddle/pkg/build"
	"github.com/google/uuid"
	"go.uber.org/zap/zapcore"
)

// Config contains the node configuration.
type Config struct {
	// ID is a unique identifier for the fuddle node.
	ID string

	// BindRegistryAddr is the bind address to listen for registry clients.
	BindRegistryAddr string
	// AdvRegistryAddr is the address to advertise to registry clients.
	AdvRegistryAddr string

	// Locality is the location of the node in the cluster.
	Locality string

	// Revision is the build commit.
	Revision string
}

func (c *Config) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("id", c.ID)

	e.AddString("bind-registry-addr", c.BindRegistryAddr)
	e.AddString("adv-registry-addr", c.AdvRegistryAddr)

	e.AddString("locality", c.Locality)

	e.AddString("revision", c.Revision)

	return nil
}

func DefaultConfig() *Config {
	return &Config{
		ID:               "fuddle-" + uuid.New().String()[:8],
		BindRegistryAddr: "0.0.0.0:8220",
		AdvRegistryAddr:  "0.0.0.0:8220",
		Locality:         "unknown",
		Revision:         build.Revision,
	}
}
