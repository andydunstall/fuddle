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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fuddle-io/fuddle/pkg/registry/cluster"
)

// Admin is a client to query the status of the cluster.
type Admin struct {
	addr   string
	client *http.Client
}

func NewAdmin(addr string) *Admin {
	return &Admin{
		addr:   addr,
		client: &http.Client{},
	}
}

func (a *Admin) Cluster(ctx context.Context) ([]*cluster.Node, error) {
	resp, err := a.get(ctx, "api/v1/cluster")
	if err != nil {
		return nil, fmt.Errorf("admin: cluster: %w", err)
	}
	defer resp.Body.Close()

	var nodes []*cluster.Node
	if err = json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return nil, fmt.Errorf("admin: cluster: %w", err)
	}
	return nodes, nil
}

func (a *Admin) Node(ctx context.Context, id string) (*cluster.Node, error) {
	resp, err := a.get(ctx, "api/v1/node/"+id)
	if err != nil {
		return nil, fmt.Errorf("admin: cluster: %w", err)
	}
	defer resp.Body.Close()

	var node *cluster.Node
	if err = json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("admin client: node, %w", err)
	}
	return node, nil
}

func (a *Admin) get(ctx context.Context, path string) (*http.Response, error) {
	url := "http://" + a.addr + "/" + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request: bad status: %d", resp.StatusCode)
	}
	return resp, nil
}
