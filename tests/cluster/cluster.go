package cluster

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/fuddle"
)

type Cluster struct {
	nodes map[*fuddle.Fuddle]interface{}
	seeds []string
}

func NewCluster(opts ...Option) (*Cluster, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(&options)
	}

	c := &Cluster{
		nodes: make(map[*fuddle.Fuddle]interface{}),
	}

	for i := 0; i != options.nodes; i++ {
		_, err := c.AddNode()
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Cluster) AddNode() (*fuddle.Fuddle, error) {
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

	registryLn, err := tcpListen(0)
	if err != nil {
		return nil, fmt.Errorf("cluster: add node: %w", err)
	}

	registryPort, err := parseAddrPort(registryLn.Addr().String())
	if err != nil {
		return nil, fmt.Errorf("cluster: add node: %w", err)
	}

	conf := config.DefaultConfig()

	conf.Registry.BindAddr = "127.0.0.1"
	conf.Registry.AdvAddr = "127.0.0.1"
	conf.Registry.BindPort = registryPort
	conf.Registry.AdvPort = registryPort

	conf.Gossip.BindAddr = "127.0.0.1"
	conf.Gossip.AdvAddr = "127.0.0.1"
	conf.Gossip.BindPort = gossipPort
	conf.Gossip.AdvPort = gossipPort
	conf.Gossip.Seeds = c.seeds

	node, err := fuddle.NewFuddle(
		conf,
		fuddle.WithRegistryListener(registryLn),
		fuddle.WithGossipTCPListener(gossipTCPLn),
		fuddle.WithGossipUDPListener(gossipUDPLn),
		fuddle.WithLogger(Logger()),
	)
	if err != nil {
		return nil, fmt.Errorf("cluster: %w", err)
	}

	c.nodes[node] = struct{}{}
	c.seeds = append(c.seeds, fmt.Sprintf("127.0.0.1:%d", gossipPort))

	return node, nil
}

func (c *Cluster) WaitForHealthy(ctx context.Context) error {
	for n := range c.nodes {
		for {
			if len(n.Nodes()) == len(c.nodes) {
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