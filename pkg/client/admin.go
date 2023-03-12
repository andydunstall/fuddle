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

	"github.com/fuddle-io/fuddle/pkg/registry"
)

type Admin struct {
	addr string
}

func NewAdmin(addr string) *Admin {
	return &Admin{
		addr: addr,
	}
}

func (a *Admin) Nodes(ctx context.Context) ([]*registry.Node, error) {
	url := fmt.Sprintf("http://%s/api/v1/cluster", a.addr)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("admin client: nodes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("admin client: nodes: bad status code: %d", resp.StatusCode)
	}

	var nodes []*registry.Node
	if err = json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return nil, fmt.Errorf("admin client: nodes, %w", err)
	}
	return nodes, nil
}

func (a *Admin) Node(ctx context.Context, id string) (*registry.Node, error) {
	url := fmt.Sprintf("http://%s/api/v1/node/%s", a.addr, id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("admin client: node: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("admin client: node: bad status code: %d", resp.StatusCode)
	}

	var node *registry.Node
	if err = json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("admin client: node, %w", err)
	}
	return node, nil
}
