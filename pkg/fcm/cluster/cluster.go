package cluster

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/fuddle"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type FuddleNode struct {
	Fuddle   *fuddle.Fuddle
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
	seeds       []string

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
	conf.Gossip.Seeds = c.seeds

	f, err := fuddle.NewFuddle(
		conf,
		fuddle.WithRPCListener(rpcLn),
		fuddle.WithAdminListener(adminLn),
		fuddle.WithGossipTCPListener(gossipTCPLn),
		fuddle.WithGossipUDPListener(gossipUDPLn),
		fuddle.WithLogger(c.logger(conf.NodeID)),
	)
	if err != nil {
		return nil, fmt.Errorf("cluster: %w", err)
	}

	node := &FuddleNode{
		Fuddle:   f,
		RPCProxy: rpcProxy,
	}

	c.fuddleNodes[node] = struct{}{}
	c.seeds = append(c.seeds, fmt.Sprintf("127.0.0.1:%d", gossipPort))

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

func (c *Cluster) LogPath(id string) string {
	return c.logDir + "/" + id + ".log"
}

func (c *Cluster) Shutdown() {
	for n := range c.fuddleNodes {
		n.Shutdown()
	}
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
