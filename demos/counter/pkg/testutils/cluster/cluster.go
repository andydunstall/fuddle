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

package cluster

import (
	"fmt"
	"net"

	"github.com/fuddle-io/fuddle/demos/counter/pkg/service/counter"
	"github.com/fuddle-io/fuddle/demos/counter/pkg/service/frontend"
	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/fuddle"
	"github.com/google/uuid"
)

type Service interface {
	Start() error
	GracefulStop()
	Stop()
}

type node struct {
	ID      string
	Service Service
}

type Cluster struct {
	nodes         []node
	fuddleSeeds   []string
	counterNodes  map[string]string
	counterAddrs  []string
	frontendAddrs []string
}

func NewCluster(opts ...Option) (*Cluster, error) {
	c := &Cluster{
		counterNodes: make(map[string]string),
	}
	if err := c.AddNodes(opts...); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Cluster) AddNodes(opts ...Option) error {
	options := options{
		fuddleNodes:   0,
		counterNodes:  0,
		frontendNodes: 0,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	for i := 0; i != options.fuddleNodes; i++ {
		if err := c.addFuddleNode(); err != nil {
			return fmt.Errorf("cluster: %w", err)
		}
	}
	for i := 0; i != options.counterNodes; i++ {
		if err := c.addCounterNode(); err != nil {
			return fmt.Errorf("cluster: %w", err)
		}
	}
	for i := 0; i != options.frontendNodes; i++ {
		if err := c.addFrontendNode(); err != nil {
			return fmt.Errorf("cluster: %w", err)
		}
	}

	for _, s := range c.nodes {
		if err := s.Service.Start(); err != nil {
			return fmt.Errorf("cluster: %w", err)
		}
	}

	return nil
}

func (c *Cluster) CounterNodeIDs() []string {
	var ids []string
	for id := range c.counterNodes {
		ids = append(ids, id)
	}
	return ids
}

func (c *Cluster) CounterNodes() map[string]string {
	nodes := make(map[string]string)
	for id, addr := range c.counterNodes {
		nodes[id] = addr
	}
	return nodes
}

func (c *Cluster) CounterAddrs() []string {
	return c.counterAddrs
}

func (c *Cluster) FrontendAddrs() []string {
	return c.frontendAddrs
}

func (c *Cluster) RemoveNode(id string) {
	for idx, node := range c.nodes {
		if id == node.ID {
			node.Service.Stop()
			c.nodes = append(c.nodes[:idx], c.nodes[idx+1:]...)
			return
		}
	}
}

func (c *Cluster) Shutdown() {
	// Shutdown nodes in reverse.
	for i := len(c.nodes) - 1; i >= 0; i-- {
		c.nodes[i].Service.GracefulStop()
	}
}

func (c *Cluster) addFuddleNode() error {
	// Add a listeners to bind to a system assigned port.
	rpcLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("add fuddle node: %s", err)
	}
	adminLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("add fuddle node: %s", err)
	}

	conf := config.DefaultConfig()
	conf.BindRegistryAddr = rpcLn.Addr().String()
	conf.AdvRegistryAddr = rpcLn.Addr().String()
	conf.Locality = "us-east-1-a"

	s := fuddle.New(
		conf,
		fuddle.WithRPCListener(rpcLn),
		fuddle.WithAdminListener(adminLn),
	)
	c.nodes = append(c.nodes, node{
		ID:      conf.ID,
		Service: s,
	})
	c.fuddleSeeds = append(c.fuddleSeeds, conf.AdvRegistryAddr)

	return nil
}

func (c *Cluster) addCounterNode() error {
	// Add a listeners to bind to a system assigned port.
	rpcLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("add counter node: %s", err)
	}

	conf := &counter.Config{
		ID:          "cluster-" + uuid.New().String()[:8],
		RPCAddr:     rpcLn.Addr().String(),
		FuddleAddrs: c.fuddleSeeds,
		Locality:    "us-east-1-a",
		Revision:    "unknown",
	}

	s := counter.NewService(
		conf,
		counter.WithRPCListener(rpcLn),
	)
	c.nodes = append(c.nodes, node{
		ID:      conf.ID,
		Service: s,
	})
	c.counterAddrs = append(c.counterAddrs, conf.RPCAddr)
	c.counterNodes[conf.ID] = conf.RPCAddr

	return nil
}

func (c *Cluster) addFrontendNode() error {
	// Add a listeners to bind to a system assigned port.
	wsLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("add frontend node: %s", err)
	}

	conf := &frontend.Config{
		ID:          "frontend-" + uuid.New().String()[:8],
		WSAddr:      wsLn.Addr().String(),
		FuddleAddrs: c.fuddleSeeds,
		Locality:    "us-east-1-a",
		Revision:    "unknown",
	}

	s := frontend.NewService(
		conf,
		frontend.WithWSListener(wsLn),
	)
	c.nodes = append(c.nodes, node{
		ID:      conf.ID,
		Service: s,
	})
	c.frontendAddrs = append(c.frontendAddrs, conf.WSAddr)

	return nil
}
