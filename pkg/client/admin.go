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
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Admin is a client to query the status of the cluster.
type Admin struct {
	addr   string
	conn   *grpc.ClientConn
	client rpc.RegistryClient
}

func NewAdmin(addr string, opts ...Option) (*Admin, error) {
	options := options{
		connectTimeout: time.Second,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	conn, client, err := connect(addr, options.connectTimeout)
	if err != nil {
		return nil, fmt.Errorf("admin: %w", err)
	}
	return &Admin{
		addr:   addr,
		conn:   conn,
		client: client,
	}, nil
}

func (a *Admin) Cluster(ctx context.Context) ([]*rpc.Node, error) {
	resp, err := a.client.Nodes(ctx, &rpc.NodesRequest{
		IncludeMetadata: false,
	})
	if err != nil {
		return nil, fmt.Errorf("admin: cluster: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf(
			"admin: cluster: %s: %s", resp.Error.Status, resp.Error.Description,
		)
	}
	return resp.Nodes, nil
}

func (a *Admin) Node(ctx context.Context, id string) (*rpc.Node, error) {
	resp, err := a.client.Node(ctx, &rpc.NodeRequest{
		NodeId: id,
	})
	if err != nil {
		return nil, fmt.Errorf("admin: cluster: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf(
			"admin: cluster: %s: %s", resp.Error.Status, resp.Error.Description,
		)
	}
	return resp.Node, nil
}

func (a *Admin) Close() {
	a.conn.Close()
}

func connect(addr string, timeout time.Duration) (*grpc.ClientConn, rpc.RegistryClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// Block until connected so we know the address is ok.
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("connect: %w", err)
	}

	client := rpc.NewRegistryClient(conn)
	return conn, client, nil
}
