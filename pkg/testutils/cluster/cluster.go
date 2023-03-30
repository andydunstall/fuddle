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

	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/fuddle"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"go.uber.org/zap"
)

type Cluster struct {
	node    *fuddle.Fuddle
	proxies []*Proxy

	logger *zap.Logger
}

func NewCluster(opts ...Option) (*Cluster, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	if options.nodes < 1 {
		return nil, fmt.Errorf("cluster: invalid number of nodes")
	}

	c := &Cluster{
		logger: testutils.Logger(),
	}

	c.logger.Info("starting cluster", zap.Int("nodes", options.nodes))

	node, nodeAddr, err := c.runFuddleNode()
	if err != nil {
		return nil, fmt.Errorf("cluster: %w", err)
	}

	// Since Fuddle doesn't yet support clustering, add a 'cluster' from the
	// clients perspective by adding a layer of proxies in front of the Fuddle
	// node.
	var proxies []*Proxy
	for i := 0; i != options.nodes; i++ {
		proxy, err := NewProxy(nodeAddr)
		if err != nil {
			return nil, fmt.Errorf("cluster: %w", err)
		}
		proxies = append(proxies, proxy)
	}

	c.node = node
	c.proxies = proxies
	return c, nil
}

func (c *Cluster) Addrs() []string {
	var addrs []string
	for _, p := range c.proxies {
		addrs = append(addrs, p.Addr())
	}
	return addrs
}

// CloseIfActive closes nodes that have active connections.
func (c *Cluster) CloseIfActive() {
	for _, p := range c.proxies {
		addr := p.Addr()
		if p.CloseIfActive() {
			c.logger.Info("closed active proxy", zap.String("addr", addr))
		}
	}
}

func (c *Cluster) BlockActiveConns() {
	for _, p := range c.proxies {
		p.BlockActiveConns()
	}
	c.logger.Info("blocked active conns")
}

// CloseWithMostConns closes node with the most active connections.
func (c *Cluster) CloseWithMostConns() {
	maxConns := 0
	var maxProxy *Proxy
	for _, p := range c.proxies {
		c := p.NumConns()
		if c > maxConns {
			maxConns = c
			maxProxy = p
		}
	}
	if maxProxy != nil {
		maxProxy.Close()
	}
}

func (c *Cluster) Close() {
	for _, p := range c.proxies {
		p.Close()
	}
	c.node.Stop()
}

func (c *Cluster) runFuddleNode() (*fuddle.Fuddle, string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", fmt.Errorf("fuddle: listen: %w", err)
	}

	server := fuddle.New(
		config.DefaultConfig(),
		fuddle.WithListener(ln),
		fuddle.WithLogger(c.logger),
	)
	if err := server.Start(); err != nil {
		return nil, "", fmt.Errorf("fuddle: run: %w", err)
	}
	return server, ln.Addr().String(), nil
}
