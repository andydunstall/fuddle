package cluster

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/node"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type FuddleNode struct {
	Fuddle   *node.Node
	RPCProxy *Proxy
}

func (n *FuddleNode) Shutdown() {
	n.RPCProxy.Close()
	n.Fuddle.Shutdown()
}

type Cluster struct {
	id          string
	fuddleNodes map[*FuddleNode]interface{}
	memberNodes map[*MemberNode]interface{}

	logDir string
}

func NewCluster(opts ...Option) (*Cluster, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(&options)
	}

	logDir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, fmt.Errorf("cluster: create log dir: %w", err)
	}

	c := &Cluster{
		id:          uuid.New().String()[:8],
		fuddleNodes: make(map[*FuddleNode]interface{}),
		memberNodes: make(map[*MemberNode]interface{}),
		logDir:      logDir,
	}
	for i := 0; i != options.fuddleNodes; i++ {
		_, err := c.AddFuddleNode()
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i != options.memberNodes; i++ {
		_, err := c.AddMemberNode()
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Cluster) ID() string {
	return c.id
}

func (c *Cluster) FuddleNodes() []*FuddleNode {
	var nodes []*FuddleNode
	for n := range c.fuddleNodes {
		nodes = append(nodes, n)
	}
	return nodes
}

func (c *Cluster) MemberNodes() []*MemberNode {
	var nodes []*MemberNode
	for n := range c.memberNodes {
		nodes = append(nodes, n)
	}
	return nodes
}

func (c *Cluster) RPCAddrs() []string {
	var addrs []string
	for n := range c.fuddleNodes {
		addrs = append(addrs, fmt.Sprintf("127.0.0.1:%d", n.Fuddle.Config.RPC.AdvPort))
	}
	return addrs
}

func (c *Cluster) GossipAddrs() []string {
	var addrs []string
	for n := range c.fuddleNodes {
		addrs = append(addrs, fmt.Sprintf("127.0.0.1:%d", n.Fuddle.Config.Gossip.AdvPort))
	}
	return addrs
}

func (c *Cluster) AddFuddleNode() (*FuddleNode, error) {
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
	conf.Gossip.Seeds = c.GossipAddrs()

	f, err := node.NewNode(
		conf,
		node.WithRPCListener(rpcLn),
		node.WithAdminListener(adminLn),
		node.WithGossipTCPListener(gossipTCPLn),
		node.WithGossipUDPListener(gossipUDPLn),
		node.WithLogPath(c.LogPath(conf.NodeID)),
	)
	if err != nil {
		return nil, fmt.Errorf("cluster: %w", err)
	}

	node := &FuddleNode{
		Fuddle:   f,
		RPCProxy: rpcProxy,
	}

	c.fuddleNodes[node] = struct{}{}

	return node, nil
}

func (c *Cluster) AddMemberNode() (*MemberNode, error) {
	id := "member-" + uuid.New().String()[:8]
	node, err := NewMemberNode(id, c.RPCAddrs(), c.logger(id))
	if err != nil {
		return nil, fmt.Errorf("add member node: %w", err)
	}
	c.memberNodes[node] = struct{}{}
	return node, nil
}

func (c *Cluster) RemoveFuddleNode() string {
	node := c.randomFuddleNode()
	if node == nil {
		return ""
	}

	node.Shutdown()
	delete(c.fuddleNodes, node)

	return node.Fuddle.Config.NodeID
}

func (c *Cluster) RemoveMemberNode() string {
	node := c.randomMemberNode()
	if node == nil {
		return ""
	}

	node.Shutdown()
	delete(c.memberNodes, node)

	return node.ID
}

func (c *Cluster) WaitForHealthy(ctx context.Context) error {
	for n := range c.fuddleNodes {
		for {
			serviceDiscoveryHealthy := len(n.Fuddle.Nodes()) == len(c.fuddleNodes)
			registryHealthy := len(n.Fuddle.Registry().UpMembers()) == len(c.fuddleNodes)
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

func (c *Cluster) DropActiveConns() {
	for n := range c.fuddleNodes {
		n.RPCProxy.Drop()
	}
}

func (c *Cluster) BlockActiveConns() {
	for n := range c.fuddleNodes {
		n.RPCProxy.BlockActiveConns()
	}
}

func (c *Cluster) LogPath(id string) string {
	return c.logDir + "/" + id + ".log"
}

func (c *Cluster) Shutdown() {
	for n := range c.memberNodes {
		n.Shutdown()
	}
	for n := range c.fuddleNodes {
		n.Shutdown()
	}
}

func (c *Cluster) randomFuddleNode() *FuddleNode {
	for k := range c.fuddleNodes {
		return k
	}
	return nil
}

func (c *Cluster) randomMemberNode() *MemberNode {
	for k := range c.memberNodes {
		return k
	}
	return nil
}

func (c *Cluster) logger(id string) *zap.Logger {
	path := c.LogPath(id)
	loggerConf := zap.NewProductionConfig()
	loggerConf.Level.SetLevel(zapcore.DebugLevel)
	loggerConf.OutputPaths = []string{path}
	return zap.Must(loggerConf.Build())
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
