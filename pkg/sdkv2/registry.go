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

// Registry manages the nodes entry into the cluster registry.
type Registry struct {
	unexp int
}

// Register registers the given node with the cluster registry.
//
// Once registered the nodes state will be propagated to the other nodes in
// the cluster. It will also stream the existing cluster state and any future
// updates to maintain a local eventually consistent view of the cluster.
//
// The given addresses are a set of seed addresses for Fuddle nodes.
func Register(addrs []string, node NodeState, opts ...Option) (*Registry, error) {
	return nil, nil
}

// Nodes returns the set of nodes in the cluster.
func (r *Registry) Nodes(opts ...NodesOption) []NodeState {
	return nil
}

// Update will update the state of this node, which will be propagated to the
// other nodes in the cluster.
func (r *Registry) Update(key string, value string) {
}

// Unregister unregisters the node from the cluster registry.
//
// Note nodes must unregister themselves before shutting down. Otherwise
// Fuddle will think the node failed rather than left.
func (r *Registry) Unregister() {
}
