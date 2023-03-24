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

type subscriber struct {
	Callback func(update *rpc.NodeUpdate)
}

type Registry struct {
	nodes map[string]*rpc.Node

	subscribers map[*subscriber]interface{}

	// mu protects the fields above.
	mu sync.Mutex
}

func NewRegistry() *Registry {
	return &Registry{
		nodes:       make(map[string]*rpc.Node),
		subscribers: make(map[*subscriber]interface{}),
	}
}

func (r *Registry) Node(id string) (*rpc.Node, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if n, ok := r.nodes[id]; ok {
		return CopyNode(n), nil
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

func (r *Registry) Subscribe(cb func(update *rpc.NodeUpdate)) func() {
	r.mu.Lock()
	defer r.mu.Unlock()

	sub := &subscriber{
		Callback: cb,
	}
	r.subscribers[sub] = struct{}{}

	// Bootstrap by sending register updates for nodes in the registry.
	for _, node := range r.nodes {
		update := &rpc.NodeUpdate{
			NodeId:     node.Id,
			UpdateType: rpc.NodeUpdateType_REGISTER,
			Attributes: CopyAttributes(node.Attributes),
			Metadata:   CopyMetadata(node.Metadata),
		}
		sub.Callback(update)
	}

	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		delete(r.subscribers, sub)
	}
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

	update := &rpc.NodeUpdate{
		NodeId:     node.Id,
		UpdateType: rpc.NodeUpdateType_REGISTER,
		Attributes: CopyAttributes(node.Attributes),
		Metadata:   CopyMetadata(node.Metadata),
	}
	// Note call subscribers with mutex locked to guarantee order.
	for sub := range r.subscribers {
		sub.Callback(update)
	}

	return nil
}

func (r *Registry) Unregister(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.nodes[id]
	if !ok {
		// If the ID is not found do nothing.
		return false
	}

	delete(r.nodes, id)

	update := &rpc.NodeUpdate{
		NodeId:     id,
		UpdateType: rpc.NodeUpdateType_UNREGISTER,
	}
	// Note call subscribers with mutex locked to guarantee order.
	for sub := range r.subscribers {
		sub.Callback(update)
	}

	return true
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

	versionedMetadataUpdate := make(map[string]*rpc.VersionedValue)
	for k, v := range metadata {
		n.Metadata[k] = &rpc.VersionedValue{
			Value: v,
		}
		versionedMetadataUpdate[k] = &rpc.VersionedValue{
			Value: v,
		}
	}

	update := &rpc.NodeUpdate{
		NodeId:     id,
		UpdateType: rpc.NodeUpdateType_METADATA,
		Metadata:   versionedMetadataUpdate,
	}
	// Note call subscribers with mutex locked to guarantee order.
	for sub := range r.subscribers {
		sub.Callback(update)
	}

	return nil
}

func CopyNode(n *rpc.Node) *rpc.Node {
	return &rpc.Node{
		Id:         n.Id,
		Attributes: CopyAttributes(n.Attributes),
		Metadata:   CopyMetadata(n.Metadata),
	}
}

func CopyAttributes(attrs *rpc.NodeAttributes) *rpc.NodeAttributes {
	return &rpc.NodeAttributes{
		Service:  attrs.Service,
		Locality: attrs.Locality,
		Created:  attrs.Created,
		Revision: attrs.Revision,
	}
}

func CopyMetadata(metadata map[string]*rpc.VersionedValue) map[string]*rpc.VersionedValue {
	cp := make(map[string]*rpc.VersionedValue)
	for k, v := range metadata {
		cp[k] = &rpc.VersionedValue{
			Value:   v.Value,
			Version: v.Version,
		}
	}
	return cp
}
