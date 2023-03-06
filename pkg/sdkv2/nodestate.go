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

package sdk

import (
	"time"
)

// NodeState represents the state of a node in the cluster.
//
// It includes both fixed attributes of the node, and mutable application
// defined state.
type NodeState struct {
	// ID is a unique identifier for the node in the cluster.
	ID string

	// Service is the name of the service running on the node.
	Service string

	// Locality is the location of the node in the cluster.
	Locality string

	// Created is the time the node was created.
	Created time.Time

	// Revision identifies the version of the service running on the node.
	Revision string

	// State contains application defined key-value pairs.
	State map[string]string
}
