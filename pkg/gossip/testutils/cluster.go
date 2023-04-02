package testutils

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/gossip"
)

type node struct {
	addr   string
	gossip *gossip.Gossip
}

func (n *node) WaitForNodes(ctx context.Context, count int) ([]gossip.Node, error) {
	for {
		if len(n.gossip.Nodes()) == count {
			return n.gossip.Nodes(), nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Millisecond * 10):
		}
	}
}

type Cluster struct {
	nodes []*node
	seeds []string
}

func NewCluster(nodes int) (*Cluster, error) {
	c := &Cluster{}
	for i := 0; i != nodes; i++ {
		if err := c.AddNode(nil, nil); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *Cluster) AddNode(onJoin func(node gossip.Node), onLeave func(node gossip.Node)) error {
	tcpLn, err := tcpListen()
	if err != nil {
		return fmt.Errorf("cluster: add node: %w", err)
	}
	addr := tcpLn.Addr().String()
	udpLn, err := udpListen(addr)
	if err != nil {
		return fmt.Errorf("cluster: add node: %w", err)
	}

	addr, port, err := parseAddr(addr)
	if err != nil {
		return fmt.Errorf("cluster: add node: %w", err)
	}

	num := len(c.nodes)
	conf := &config.Config{
		NodeID: fmt.Sprintf("node-%d", num),
		Registry: config.Registry{
			BindAddr: "127.0.0.1",
			BindPort: 1000 + num,
			AdvAddr:  "127.0.0.1",
			AdvPort:  1000 + num,
		},
		Gossip: config.Gossip{
			BindAddr: addr,
			BindPort: port,
			AdvAddr:  addr,
			AdvPort:  port,
			Seeds:    c.seeds,
		},
	}
	g, err := gossip.NewGossip(
		conf,
		gossip.WithOnJoin(onJoin),
		gossip.WithOnLeave(onLeave),
		gossip.WithTCPListener(tcpLn),
		gossip.WithUDPListener(udpLn),
		gossip.WithLogger(Logger()),
	)
	if err != nil {
		return fmt.Errorf("cluster: %w", err)
	}

	c.seeds = append(c.seeds, tcpLn.Addr().String())

	n := &node{
		addr:   fmt.Sprintf("%s: %d", conf.Gossip.AdvAddr, conf.Gossip.AdvPort),
		gossip: g,
	}
	c.nodes = append(c.nodes, n)
	return nil
}

func (c *Cluster) RemoveNode(n int) {
	node := c.nodes[n]
	node.gossip.Shutdown()
	c.nodes = append(c.nodes[:n], c.nodes[n+1:]...)
}

func (c *Cluster) WaitForHealthy(ctx context.Context) error {
	var discovered []gossip.Node
	for _, node := range c.nodes {
		nodes, err := node.WaitForNodes(ctx, len(c.nodes))
		if err != nil {
			return err
		}
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].ID < nodes[j].ID
		})

		if discovered != nil && !reflect.DeepEqual(nodes, discovered) {
			return fmt.Errorf("nodes not equal: %v != %v", nodes, discovered)
		}
		discovered = nodes
	}
	return nil
}

func (c *Cluster) Shutdown() {
	for _, node := range c.nodes {
		node.gossip.Shutdown()
	}
}

func tcpListen() (*net.TCPListener, error) {
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("tcp listen: %w", err)
	}
	return ln, nil
}

func udpListen(addr string) (*net.UDPConn, error) {
	_, port, err := parseAddr(addr)
	if err != nil {
		return nil, fmt.Errorf("udp listen: %w", err)
	}

	ln, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("udp listen: %w", err)
	}
	return ln, nil
}

func parseAddr(addr string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, fmt.Errorf("parse addr: %w", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("parse addr: %w", err)
	}
	return host, port, nil
}
