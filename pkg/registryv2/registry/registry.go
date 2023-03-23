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

package registry

import (
	"fmt"
	"sync"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
)

var (
	ErrAlreadyRegistered = fmt.Errorf("node already registered")
	ErrNotFound          = fmt.Errorf("not found")
	ErrInvalidUpdate     = fmt.Errorf("invalid update")
)

type Registry struct {
	nodes map[string]*rpc.Node

	// mu protects the fields above.
	mu sync.Mutex
}

func NewRegistry() *Registry {
	return &Registry{
		nodes: make(map[string]*rpc.Node),
	}
}

func (r *Registry) Node(id string) (*rpc.Node, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if n, ok := r.nodes[id]; ok {
		return n, nil
	}
	return nil, ErrNotFound
}

func (r *Registry) Nodes(includeMetadata bool) []*rpc.Node {
	r.mu.Lock()
	defer r.mu.Unlock()

	var nodes []*rpc.Node
	for _, node := range r.nodes {
		// Copy to modify metadata.
		node := CopyNode(node)
		if !includeMetadata {
			node.Metadata = nil
		}

		nodes = append(nodes, node)
	}

	return nodes
}

func (r *Registry) Register(node *rpc.Node) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if node.Id == "" || node.Attributes == nil || node.Metadata == nil {
		return ErrInvalidUpdate
	}

	if _, ok := r.nodes[node.Id]; ok {
		return ErrAlreadyRegistered
	}

	r.nodes[node.Id] = node

	return nil
}

func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// If the ID is not found do nothing.
	delete(r.nodes, id)

	return nil
}

func (r *Registry) UpdateNode(id string, metadata map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	n, ok := r.nodes[id]
	if !ok {
		return ErrNotFound
	}

	if metadata == nil {
		return ErrInvalidUpdate
	}

	for k, v := range metadata {
		n.Metadata[k] = &rpc.VersionedValue{
			Value: v,
		}
	}

	return nil
}

func CopyNode(n *rpc.Node) *rpc.Node {
	metadata := make(map[string]*rpc.VersionedValue)
	for k, v := range n.Metadata {
		metadata[k] = &rpc.VersionedValue{
			Value:   v.Value,
			Version: v.Version,
		}
	}

	return &rpc.Node{
		Id: n.Id,
		Attributes: &rpc.NodeAttributes{
			Service:  n.Attributes.Service,
			Locality: n.Attributes.Locality,
			Created:  n.Attributes.Created,
			Revision: n.Attributes.Revision,
		},
		Metadata: metadata,
	}
}
