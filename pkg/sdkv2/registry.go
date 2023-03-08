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
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/andydunstall/fuddle/pkg/rpc"
	multierror "github.com/hashicorp/go-multierror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Registry manages the nodes entry into the cluster registry.
type Registry struct {
	// nodeID is the ID of this registered node.
	nodeID string

	cluster *cluster

	// conn is the underlying gRPC connection to the Fuddle server.
	conn *grpc.ClientConn
	// stream is the registry connection to the Fuddle server.
	stream rpc.Registry_RegisterClient

	wg sync.WaitGroup
}

// Register registers the given node with the cluster registry.
//
// Once registered the nodes state will be propagated to the other nodes in
// the cluster. It will also stream the existing cluster state and any future
// updates to maintain a local eventually consistent view of the cluster.
//
// The given addresses are a set of seed addresses for Fuddle nodes.
func Register(addrs []string, node NodeState) (*Registry, error) {
	conn, stream, err := connect(addrs)
	if err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}

	r := &Registry{
		nodeID:  node.ID,
		cluster: newCluster(node),
		conn:    conn,
		stream:  stream,
		wg:      sync.WaitGroup{},
	}
	if err = r.sendJoinRPC(node); err != nil {
		r.conn.Close()
		return nil, fmt.Errorf("registry: %w", err)
	}

	r.wg.Add(1)
	go r.sync()

	return r, nil
}

// Nodes returns the set of nodes in the cluster.
func (r *Registry) Nodes(opts ...NodesOption) []NodeState {
	return r.cluster.Nodes(opts...)
}

// Subscribe registers the given callback to fire when the registry state
// changes.
//
// Note the callback is called synchronously with the registry mutex held,
// therefore it must NOT block or callback to the registry (or it will
// deadlock).
func (r *Registry) Subscribe(cb func()) func() {
	return r.cluster.Subscribe(cb)
}

// UpdateLocalState will update the state of this node, which will be propagated
// to the other nodes in the cluster.
func (r *Registry) UpdateLocalState(update map[string]string) error {
	if err := r.cluster.UpdateLocalState(update); err != nil {
		return fmt.Errorf("registry: %w", err)
	}
	if err := r.sendUpdateRPC(update); err != nil {
		return fmt.Errorf("registry: %w", err)
	}
	return nil
}

// Unregister unregisters the node from the cluster registry.
//
// Note nodes must unregister themselves before shutting down. Otherwise
// Fuddle will think the node failed rather than left.
func (r *Registry) Unregister() error {
	err := r.sendLeaveRPC()

	r.conn.Close()
	r.wg.Wait()

	if err != nil {
		return fmt.Errorf("registry: %w", err)
	}
	return nil
}

func (r *Registry) sync() {
	defer r.wg.Done()

	for {
		update, err := r.stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			return
		}

		if err := r.applyUpdate(update); err != nil {
			return
		}
	}
}

func (r *Registry) sendJoinRPC(node NodeState) error {
	rpcUpdate := &rpc.NodeUpdate{
		NodeId:     node.ID,
		UpdateType: rpc.NodeUpdateType_JOIN,
		Attributes: &rpc.Attributes{
			Service:  node.Service,
			Locality: node.Locality,
			Created:  node.Created,
			Revision: node.Revision,
		},
		State: node.State,
	}
	if err := r.stream.Send(rpcUpdate); err != nil {
		return fmt.Errorf("send join rpc: %w", err)
	}
	return nil
}

func (r *Registry) sendUpdateRPC(update map[string]string) error {
	rpcUpdate := &rpc.NodeUpdate{
		NodeId:     r.nodeID,
		UpdateType: rpc.NodeUpdateType_STATE,
		State:      update,
	}
	if err := r.stream.Send(rpcUpdate); err != nil {
		return fmt.Errorf("send join rpc: %w", err)
	}
	return nil
}

func (r *Registry) sendLeaveRPC() error {
	rpcUpdate := &rpc.NodeUpdate{
		NodeId:     r.nodeID,
		UpdateType: rpc.NodeUpdateType_LEAVE,
	}
	if err := r.stream.Send(rpcUpdate); err != nil {
		return fmt.Errorf("send leave rpc: %w", err)
	}
	return nil
}

func (r *Registry) applyUpdate(update *rpc.NodeUpdate) error {
	switch update.UpdateType {
	case rpc.NodeUpdateType_JOIN:
		if err := r.applyJoinUpdateLocked(update); err != nil {
			return err
		}
	case rpc.NodeUpdateType_LEAVE:
		r.applyLeaveUpdateLocked(update)
	case rpc.NodeUpdateType_STATE:
		if err := r.applyStateUpdateLocked(update); err != nil {
			return err
		}
	default:
		return fmt.Errorf("cluster: unknown update type: %s", update.UpdateType)
	}

	return nil
}

func (r *Registry) applyJoinUpdateLocked(update *rpc.NodeUpdate) error {
	if update.NodeId == "" {
		return fmt.Errorf("cluster: join update: missing id")
	}

	if update.Attributes == nil {
		return fmt.Errorf("cluster: join update: missing attributes")
	}

	node := NodeState{
		ID:       update.NodeId,
		Service:  update.Attributes.Service,
		Locality: update.Attributes.Locality,
		Revision: update.Attributes.Revision,
		Created:  update.Attributes.Created,
		// Copy the state to avoid modifying the update. If update.State is
		// nil this returns an empty map.
		State: CopyState(update.State),
	}
	return r.cluster.AddNode(node)
}

func (r *Registry) applyLeaveUpdateLocked(update *rpc.NodeUpdate) {
	r.cluster.RemoveNode(update.NodeId)
}

func (r *Registry) applyStateUpdateLocked(update *rpc.NodeUpdate) error {
	// If the update is missing state must ignore it.
	if update.State == nil {
		return nil
	}
	return r.cluster.UpdateState(update.NodeId, update.State)
}

func connect(addrs []string) (*grpc.ClientConn, rpc.Registry_RegisterClient, error) {
	var result error
	for _, addr := range addrs {
		conn, err := grpc.Dial(
			addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		client := rpc.NewRegistryClient(conn)
		stream, err := client.Register(context.Background())
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		return conn, stream, nil
	}

	return nil, nil, fmt.Errorf("connect: %w", result)
}
