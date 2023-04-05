package cluster

import (
	"fmt"
	"io"
	"net"
	"sync"

	"go.uber.org/atomic"
)

type Proxy struct {
	conns map[*proxyConn]interface{}

	// mu is a mutex protecting the fields above.
	mu sync.Mutex

	target string
	ln     net.Listener
}

func NewProxy(target string) (*Proxy, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("proxy: listen: %w", err)
	}

	proxy := &Proxy{
		conns:  make(map[*proxyConn]interface{}),
		target: target,
		ln:     ln,
	}
	go proxy.acceptLoop()
	return proxy, nil
}

// Drop drops all existing connections.
func (p *Proxy) Drop() {
	p.mu.Lock()
	for c := range p.conns {
		c.Close()
	}
	p.conns = make(map[*proxyConn]interface{})
	p.mu.Unlock()
}

func (p *Proxy) Addr() string {
	return p.ln.Addr().String()
}

func (p *Proxy) Close() {
	for c := range p.conns {
		c.Close()
	}
	p.ln.Close()
}

func (p *Proxy) acceptLoop() {
	for {
		downstream, err := p.ln.Accept()
		if err != nil {
			return
		}

		upstream, err := net.Dial("tcp", p.target)
		if err != nil {
			return
		}

		conn := newProxyConn(downstream, upstream)
		p.mu.Lock()
		p.conns[conn] = struct{}{}
		p.mu.Unlock()
	}
}

type proxyConn struct {
	upstream   net.Conn
	downstream net.Conn
	blocked    *atomic.Bool
}

func newProxyConn(upstream net.Conn, downstream net.Conn) *proxyConn {
	conn := &proxyConn{
		upstream:   upstream,
		downstream: downstream,
		blocked:    atomic.NewBool(false),
	}
	go conn.forward(downstream, upstream)
	go conn.forward(upstream, downstream)
	return conn
}

func (c *proxyConn) SetBlock(blocked bool) {
	c.blocked.Store(blocked)
}

func (c *proxyConn) Close() {
	c.upstream.Close()
	c.downstream.Close()
}

func (c *proxyConn) forward(dst io.Writer, src io.Reader) {
	buf := make([]byte, 1024)
	for {
		n, err := src.Read(buf)
		if err != nil {
			return
		}

		if c.blocked.Load() {
			continue
		}

		_, err = dst.Write(buf[0:n])
		if err != nil {
			return
		}
	}
}
