package cluster

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/fuddle-io/fuddle/demos/clock/pkg/services/clock"
	"github.com/fuddle-io/fuddle/demos/clock/pkg/services/frontend"
	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/fuddle"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Cluster struct {
	fuddleNodes   map[*fuddle.Fuddle]interface{}
	clockNodes    map[*clock.Service]interface{}
	frontendNodes map[*frontend.Service]interface{}
	seeds         []string
	logDir        string
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
		fuddleNodes:   make(map[*fuddle.Fuddle]interface{}),
		clockNodes:    make(map[*clock.Service]interface{}),
		frontendNodes: make(map[*frontend.Service]interface{}),
		logDir:        logDir,
	}
	for i := 0; i != options.fuddleNodes; i++ {
		if err := c.addFuddleNode(); err != nil {
			return nil, fmt.Errorf("cluster: %w", err)
		}
	}
	for i := 0; i != options.clockNodes; i++ {
		if err := c.addClockNode(); err != nil {
			return nil, fmt.Errorf("cluster: %w", err)
		}
	}
	for i := 0; i != options.frontendNodes; i++ {
		if err := c.addFrontendNode(); err != nil {
			return nil, fmt.Errorf("cluster: %w", err)
		}
	}

	return c, nil
}

func (c *Cluster) FuddleNodes() []*fuddle.Fuddle {
	var nodes []*fuddle.Fuddle
	for n := range c.fuddleNodes {
		nodes = append(nodes, n)
	}
	return nodes
}

func (c *Cluster) ClockNodes() []*clock.Service {
	var nodes []*clock.Service
	for n := range c.clockNodes {
		nodes = append(nodes, n)
	}
	return nodes
}

func (c *Cluster) FrontendNodes() []*frontend.Service {
	var nodes []*frontend.Service
	for n := range c.frontendNodes {
		nodes = append(nodes, n)
	}
	return nodes
}

func (c *Cluster) LogPath(id string) string {
	return c.logDir + "/" + id + ".log"
}

func (c *Cluster) Shutdown() {
	for n := range c.fuddleNodes {
		n.Shutdown()
	}
	for n := range c.clockNodes {
		n.Shutdown()
	}
	for n := range c.frontendNodes {
		n.Shutdown()
	}
}

func (c *Cluster) addFuddleNode() error {
	gossipTCPLn, err := tcpListen(0)
	if err != nil {
		return fmt.Errorf("add fuddle node: %w", err)
	}

	gossipPort, err := parseAddrPort(gossipTCPLn.Addr().String())
	if err != nil {
		return fmt.Errorf("add fuddle node: %w", err)
	}

	gossipUDPLn, err := udpListen(gossipPort)
	if err != nil {
		return fmt.Errorf("add fuddle node: %w", err)
	}

	rpcLn, err := tcpListen(0)
	if err != nil {
		return fmt.Errorf("add fuddle node: %w", err)
	}

	rpcPort, err := parseAddrPort(rpcLn.Addr().String())
	if err != nil {
		return fmt.Errorf("add fuddle node: %w", err)
	}

	conf := config.DefaultConfig()

	conf.RPC.BindAddr = "127.0.0.1"
	conf.RPC.AdvAddr = "127.0.0.1"
	conf.RPC.BindPort = rpcPort
	conf.RPC.AdvPort = rpcPort

	conf.Gossip.BindAddr = "127.0.0.1"
	conf.Gossip.AdvAddr = "127.0.0.1"
	conf.Gossip.BindPort = gossipPort
	conf.Gossip.AdvPort = gossipPort
	conf.Gossip.Seeds = c.seeds

	node, err := fuddle.NewFuddle(
		conf,
		fuddle.WithRPCListener(rpcLn),
		fuddle.WithGossipTCPListener(gossipTCPLn),
		fuddle.WithGossipUDPListener(gossipUDPLn),
		fuddle.WithLogger(c.logger(conf.NodeID)),
	)
	if err != nil {
		return fmt.Errorf("add fuddle node: %w", err)
	}

	c.fuddleNodes[node] = struct{}{}
	c.seeds = append(c.seeds, fmt.Sprintf("127.0.0.1:%d", gossipPort))

	return nil
}

func (c *Cluster) addClockNode() error {
	var fuddleAddrs []string
	for n := range c.fuddleNodes {
		fuddleAddrs = append(fuddleAddrs, n.Config.RPC.JoinAdvAddr())
	}

	ln, err := tcpListen(0)
	if err != nil {
		return fmt.Errorf("add clock node: %w", err)
	}

	node, err := clock.NewService(ln, fuddleAddrs)
	if err != nil {
		return fmt.Errorf("add clock node: %w", err)
	}
	c.clockNodes[node] = struct{}{}
	return nil
}

func (c *Cluster) addFrontendNode() error {
	var fuddleAddrs []string
	for n := range c.fuddleNodes {
		fuddleAddrs = append(fuddleAddrs, n.Config.RPC.JoinAdvAddr())
	}

	ln, err := tcpListen(0)
	if err != nil {
		return fmt.Errorf("add frontend node: %w", err)
	}

	node, err := frontend.NewService(ln, fuddleAddrs)
	if err != nil {
		return fmt.Errorf("add frontend node: %w", err)
	}
	c.frontendNodes[node] = struct{}{}
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
