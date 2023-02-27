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

package client

import (
	"context"
	"fmt"

	"github.com/andydunstall/fuddle/pkg/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Registry struct {
	conn   *grpc.ClientConn
	client rpc.RegistryClient
}

func ConnectRegistry(addr string) (*Registry, error) {
	conn, err := grpc.Dial(
		addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("registry client: %w", err)
	}

	return &Registry{
		conn:   conn,
		client: rpc.NewRegistryClient(conn),
	}, nil
}

func (r *Registry) Register(ctx context.Context, id string) error {
	_, err := r.client.Register(context.Background(), &rpc.RegisterRequest{
		NodeId: id,
	})
	if err != nil {
		return fmt.Errorf("registry client: register: %w", err)
	}
	return nil
}

func (r *Registry) Unregister(ctx context.Context, id string) error {
	_, err := r.client.Unregister(context.Background(), &rpc.UnregisterRequest{
		NodeId: id,
	})
	if err != nil {
		return fmt.Errorf("registry client: unregister: %w", err)
	}
	return nil
}

func (r *Registry) Nodes(ctx context.Context) ([]string, error) {
	resp, err := r.client.Nodes(context.Background(), &rpc.NodesRequest{})
	if err != nil {
		return nil, fmt.Errorf("registry client: nodes: %w", err)
	}
	return resp.Ids, nil
}

func (r *Registry) Close() {
	r.conn.Close()
}
