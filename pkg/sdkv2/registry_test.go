// Copyright (C) 2023 Andrew Dunstall
//
// Registry is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Registry is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package sdk

import (
	"fmt"
	"time"
)

// Registers an 'orders' service node in 'us-east-1-b'.
func Example_registerOrdersServiceNode() {
	registry, err := Register(
		// Seed addresses of Fuddle servers.
		[]string{"192.168.1.1:8220", "192.168.1.2:8220", "192.168.1.3:8220"},
		NodeState{
			ID:       "orders-32eaba4e",
			Service:  "orders",
			Locality: "aws.us-east-1-b",
			Created:  time.Now(),
			Revision: "v5.1.0-812ebbc",
			State: map[string]string{
				"status":           "booting",
				"addr.rpc.ip":      "192.168.2.1",
				"addr.rpc.port":    "5562",
				"addr.admin.ip":    "192.168.2.1",
				"addr.admin.port":  "7723",
				"protocol.version": "3",
				"instance":         "i-0bc636e38d6c537a7",
			},
		},
	)
	if err != nil {
		// handle err ...
	}
	defer registry.Unregister()

	// ...

	// Once ready update the nodes status to 'active'. This update will be
	// propagated to the other nodes in the cluster.
	registry.Update("status", "active")
}

// Registers a 'frontend' service and queries the set of active order service
// nodes in us-east-1.
func Example_lookupOrdersServiceNodes() {
	registry, err := Register(
		// Seed addresses of Fuddle servers.
		[]string{"192.168.1.1:8220", "192.168.1.2:8220", "192.168.1.3:8220"},
		NodeState{
			ID:       "frontend-9fe2a841",
			Service:  "frontend",
			Locality: "aws.us-east-1-c",
			Created:  time.Now(),
			Revision: "v2.0.1-217cbf1",
			State: map[string]string{
				"status":           "active",
				"addr.http.ip":     "192.168.3.155",
				"addr.http.port":   "8080",
				"protocol.version": "3",
				"instance":         "i-0b78b6770d3068dea",
			},
		},
	)
	if err != nil {
		// handle err ...
	}
	defer registry.Unregister()

	// Filter to only include order service nodes in us-east-1 whose status
	// is active and protocol version is either 2 or 3.
	orderNodes := registry.Nodes(WithFilter(Filter{
		"order": {
			Locality: []string{"aws.us-east-1-*"},
			State: StateFilter{
				"status":           []string{"active"},
				"protocol.version": []string{"2", "3"},
			},
		},
	}))
	addrs := []string{}
	for _, node := range orderNodes {
		addr := node.State["addr.rpc.ip"] + ":" + node.State["addr.rpc.port"]
		addrs = append(addrs, addr)
	}

	// ...

	fmt.Println("order service:", addrs)
}
