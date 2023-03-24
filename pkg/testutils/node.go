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

package testutils

import (
	"math/rand"
	"time"

	fuddle "github.com/fuddle-io/fuddle-go"
	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/google/uuid"
)

func WaitForNode(client *fuddle.Fuddle, id string) (fuddle.Node, error) {
	var node fuddle.Node
	found := false
	recvCh := make(chan interface{})
	unsubscribe := client.Subscribe(func(nodes []fuddle.Node) {
		if found {
			return
		}

		for _, n := range nodes {
			if n.ID == id {
				found = true
				node = n
				close(recvCh)
				return
			}
		}
	})
	defer unsubscribe()

	if err := WaitWithTimeout(recvCh, time.Millisecond*500); err != nil {
		return node, err
	}
	return node, nil
}

func WaitForCount(client *fuddle.Fuddle, count int) error {
	found := false
	recvCh := make(chan interface{})
	unsubscribe := client.Subscribe(func(nodes []fuddle.Node) {
		if found {
			return
		}

		if len(nodes) == count {
			found = true
			close(recvCh)
			return
		}
	})
	defer unsubscribe()

	if err := WaitWithTimeout(recvCh, time.Millisecond*500); err != nil {
		return err
	}
	return nil
}

// RandomNode returns a node with random attributes and state.
func RandomNode() fuddle.Node {
	return fuddle.Node{
		ID:       uuid.New().String(),
		Service:  uuid.New().String(),
		Locality: uuid.New().String(),
		Created:  rand.Int63(),
		Revision: uuid.New().String(),
		Metadata: RandomMetadata(),
	}
}

// RandomSDKNode returns a node with random attributes and state.
func RandomSDKNode() fuddle.Node {
	return fuddle.Node{
		ID:       uuid.New().String(),
		Service:  uuid.New().String(),
		Locality: uuid.New().String(),
		Created:  rand.Int63(),
		Revision: uuid.New().String(),
		Metadata: RandomMetadata(),
	}
}

// RandomRPCNode returns a node with random attributes and state.
func RandomRPCNode() *rpc.Node {
	metadata := RandomVersionedMetadata()
	return &rpc.Node{
		Id: uuid.New().String(),
		Attributes: &rpc.NodeAttributes{
			Service:  uuid.New().String(),
			Locality: uuid.New().String(),
			Created:  rand.Int63(),
			Revision: uuid.New().String(),
		},
		Metadata: metadata,
	}
}

func RandomMetadata() map[string]string {
	return map[string]string{
		uuid.New().String(): uuid.New().String(),
		uuid.New().String(): uuid.New().String(),
		uuid.New().String(): uuid.New().String(),
		uuid.New().String(): uuid.New().String(),
		uuid.New().String(): uuid.New().String(),
	}
}

func RandomVersionedMetadata() map[string]*rpc.VersionedValue {
	metadata := make(map[string]*rpc.VersionedValue)
	for i := 0; i != 5; i++ {
		version := rand.Uint64()
		metadata[uuid.New().String()] = &rpc.VersionedValue{
			Value:   uuid.New().String(),
			Version: version,
		}
	}
	return metadata
}

func RPCNodeToSDKNode(n *rpc.Node) fuddle.Node {
	metadata := make(map[string]string)
	for k, vv := range n.Metadata {
		metadata[k] = vv.Value
	}
	return fuddle.Node{
		ID:       n.Id,
		Service:  n.Attributes.Service,
		Locality: n.Attributes.Locality,
		Created:  n.Attributes.Created,
		Revision: n.Attributes.Revision,
		Metadata: metadata,
	}
}
