package cluster

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/fuddle"
	"github.com/fuddle-io/fuddle/pkg/testutils"
)

type Node struct {
	Fuddle   *fuddle.Fuddle
	RPCProxy *Proxy
}

func (n *Node) Shutdown() {
	n.RPCProxy.Close()
	n.Fuddle.Shutdown()
}

type Cluster struct {
	nodes map[*Node]interface{}
	seeds []string

	options options
}

func NewCluster(opts ...Option) (*Cluster, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(&options)
	}

	c := &Cluster{
		nodes:   make(map[*Node]interface{}),
		options: options,
	}

	for i := 0; i != options.nodes; i++ {
		_, err := c.AddNode()
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Cluster) Nodes() []*Node {
	var nodes []*Node
	for n := range c.nodes {
		nodes = append(nodes, n)
	}
	return nodes
}

func (c *Cluster) RPCAddrs() []string {
	var addrs []string
	for n := range c.nodes {
		addrs = append(addrs, fmt.Sprintf("127.0.0.1:%d", n.Fuddle.Config.RPC.AdvPort))
	}
	return addrs
}

func (c *Cluster) AddNode() (*Node, error) {
	gossipTCPLn, err := tcpListen(0)
	if err != nil {
		return nil, fmt.Errorf("cluster: add node: %w", err)
	}

	gossipPort, err := parseAddrPort(gossipTCPLn.Addr().String())
	if err != nil {
		return nil, fmt.Errorf("cluster: add node: %w", err)
	}

	gossipUDPLn, err := udpListen(gossipPort)
	if err != nil {
		return nil, fmt.Errorf("cluster: add node: %w", err)
	}

	rpcLn, err := tcpListen(0)
	if err != nil {
		return nil, fmt.Errorf("cluster: add node: %w", err)
	}

	rpcPort, err := parseAddrPort(rpcLn.Addr().String())
	if err != nil {
		return nil, fmt.Errorf("cluster: add node: %w", err)
	}

	rpcProxy, err := NewProxy(rpcLn.Addr().String())
	if err != nil {
		return nil, fmt.Errorf("cluster: add node: %w", err)
	}

	rpcProxyPort, err := parseAddrPort(rpcProxy.Addr())
	if err != nil {
		return nil, fmt.Errorf("cluster: add node: %w", err)
	}

	adminLn, err := tcpListen(0)
	if err != nil {
		return nil, fmt.Errorf("cluster: add node: %w", err)
	}

	adminPort, err := parseAddrPort(adminLn.Addr().String())
	if err != nil {
		return nil, fmt.Errorf("cluster: add node: %w", err)
	}

	conf := config.DefaultConfig()

	conf.RPC.BindAddr = "127.0.0.1"
	conf.RPC.AdvAddr = "127.0.0.1"
	conf.RPC.BindPort = rpcPort
	conf.RPC.AdvPort = rpcProxyPort

	conf.Admin.BindAddr = "127.0.0.1"
	conf.Admin.BindPort = adminPort
	conf.Admin.AdvAddr = "127.0.0.1"
	conf.Admin.AdvPort = adminPort

	conf.Gossip.BindAddr = "127.0.0.1"
	conf.Gossip.AdvAddr = "127.0.0.1"
	conf.Gossip.BindPort = gossipPort
	conf.Gossip.AdvPort = gossipPort
	conf.Gossip.Seeds = c.seeds

	if c.options.registryConfig != nil {
		conf.Registry = c.options.registryConfig
	}

	f, err := fuddle.NewFuddle(
		conf,
		fuddle.WithRPCListener(rpcLn),
		fuddle.WithAdminListener(adminLn),
		fuddle.WithGossipTCPListener(gossipTCPLn),
		fuddle.WithGossipUDPListener(gossipUDPLn),
		fuddle.WithLogLevel(testutils.LogLevel()),
	)
	if err != nil {
		return nil, fmt.Errorf("cluster: %w", err)
	}

	node := &Node{
		Fuddle:   f,
		RPCProxy: rpcProxy,
	}

	c.nodes[node] = struct{}{}
	c.seeds = append(c.seeds, fmt.Sprintf("127.0.0.1:%d", gossipPort))

	return node, nil
}

func (c *Cluster) DropActiveConns() {
	for n := range c.nodes {
		n.RPCProxy.Drop()
	}
}

func (c *Cluster) BlockActiveConns() {
	for n := range c.nodes {
		n.RPCProxy.BlockActiveConns()
	}
}

func (c *Cluster) RemoveNode(node *Node) {
	node.Shutdown()
	delete(c.nodes, node)
}

func (c *Cluster) WaitForHealthy(ctx context.Context) error {
	for n := range c.nodes {
		for {
			serviceDiscoveryHealthy := len(n.Fuddle.Nodes()) == len(c.nodes)
			registryHealthy := len(n.Fuddle.Registry().UpMembers()) == len(c.nodes)
			if serviceDiscoveryHealthy && registryHealthy {
				break
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Millisecond * 10):
			}
		}
	}
	return nil
}

func (c *Cluster) Shutdown() {
	for n := range c.nodes {
		n.Shutdown()
	}
}

func tcpListen(port int) (*net.TCPListener, error) {
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("tcp listen: %w", err)
	}
	return ln, nil
}

func udpListen(port int) (*net.UDPConn, error) {
	ln, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("udp listen: %w", err)
	}
	return ln, nil
}

func parseAddrPort(addr string) (int, error) {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, fmt.Errorf("parse addr: %w", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("parse addr: %w", err)
	}
	return port, nil
}
