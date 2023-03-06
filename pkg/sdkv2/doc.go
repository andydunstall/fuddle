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

/*
Package sdk provides an SDK for nodes to register, unregister and query the
state of the cluster.

# Registering

To enter the Fuddle registry nodes must register themselves.

The registered node state contains a fixed set of attributes (ID, service,
locality, ...) and a mutable application defined state. The fixed set of
attributes cannot change, though the application state may be updated after
registering, where updates will be propagated around the cluster.

Once registered, the registry client streams the existing cluster state (which
includes the set of nodes in the cluster and the state of those nodes), and
all further updates. Therefore the client maintains an eventually consistent
view of the cluster, which the node can query without having to make RPCs back
to the Fuddle server.

# Unregistering

When a node is shutdown, it must first unregister from the Fuddle registry.
Otherwise Fuddle will view the node as failed rather than having left the
cluster.

# Lookup

Since the client maintains its own local cluster view, which can be queried by
the node.

Instead of receiving all nodes, clients can optionally filter the nodes based on
service, locality and state.
*/
package sdk
