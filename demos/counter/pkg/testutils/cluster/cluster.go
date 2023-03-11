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

	"github.com/andydunstall/fuddle/demos/counter/pkg/service/counter"
	"github.com/andydunstall/fuddle/pkg/config"
	"github.com/andydunstall/fuddle/pkg/server"
	"github.com/google/uuid"
)

type Service interface {
	Start() error
	GracefulStop()
}

type Cluster struct {
	services    []Service
	fuddleSeeds []string
}

func NewCluster(opts ...Option) (*Cluster, error) {
	options := options{
		fuddleNodes:  0,
		counterNodes: 0,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	c := &Cluster{}
	for i := 0; i != options.fuddleNodes; i++ {
		if err := c.addFuddleNode(); err != nil {
			return nil, fmt.Errorf("cluster: %w", err)
		}
	}
	for i := 0; i != options.counterNodes; i++ {
		if err := c.addCounterNode(); err != nil {
			return nil, fmt.Errorf("cluster: %w", err)
		}
	}

	for _, s := range c.services {
		if err := s.Start(); err != nil {
			return nil, fmt.Errorf("cluster: %w", err)
		}
	}

	return c, nil
}

func (c *Cluster) Shutdown() {
	// Shutdown services in reverse.
	for i := len(c.services) - 1; i >= 0; i-- {
		c.services[i].GracefulStop()
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
	conf.BindAddr = rpcLn.Addr().String()
	conf.AdvAddr = rpcLn.Addr().String()
	conf.BindAdminAddr = adminLn.Addr().String()
	conf.AdvAdminAddr = adminLn.Addr().String()
	conf.Locality = "us-east-1-a"

	s := server.NewServer(
		conf,
		server.WithRPCListener(rpcLn),
		server.WithAdminListener(adminLn),
	)
	c.services = append(c.services, s)
	c.fuddleSeeds = append(c.fuddleSeeds, conf.AdvAddr)

	return nil
}

func (c *Cluster) addCounterNode() error {
	// Add a listeners to bind to a system assigned port.
	rpcLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("add fuddle node: %s", err)
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
	c.services = append(c.services, s)

	return nil
}
