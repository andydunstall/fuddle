package gossip

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/armon/go-metrics"
	"github.com/fuddle-io/fuddle/pkg/config"
	sockaddr "github.com/hashicorp/go-sockaddr"
	"github.com/hashicorp/memberlist"
)

const (
	// udpPacketBufSize is used to buffer incoming packets during read
	// operations.
	udpPacketBufSize = 65536

	// udpRecvBufSize is a large buffer size that we attempt to set UDP
	// sockets to in order to handle a large volume of messages.
	udpRecvBufSize = 2 * 1024 * 1024
)

// transport is a memberlist.Transport implementation that uses connectionless
// UDP for packet operations, and ad-hoc TCP connections for stream operations.
//
// transport is the same as memberlist.NetTransport except it supports optional
// listeners. See memberlist/net_transport.go.
type transport struct {
	packetCh    chan *memberlist.Packet
	streamCh    chan net.Conn
	logger      *log.Logger
	wg          sync.WaitGroup
	tcpListener *net.TCPListener
	udpListener *net.UDPConn
	shutdown    int32

	metricLabels []metrics.Label
}

var _ memberlist.NodeAwareTransport = (*transport)(nil)

// Newtransport returns a net transport with the given configuration. On
// success all the network listeners will be created and listening.
func newTransport(conf *config.Gossip, options options) (*transport, error) {
	// Build out the new transport.
	var ok bool
	t := transport{
		packetCh: make(chan *memberlist.Packet),
		streamCh: make(chan net.Conn),
		logger:   log.Default(),
	}

	// Clean up listeners if there's an error.
	defer func() {
		if !ok {
			t.Shutdown()
		}
	}()

	ip := net.ParseIP(conf.BindAddr)

	tcpLn := options.tcpListener
	if tcpLn == nil {

		tcpAddr := &net.TCPAddr{IP: ip, Port: conf.BindPort}

		var err error
		tcpLn, err = net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			return nil, fmt.Errorf("Failed to start TCP listener on %q port %d: %v", conf.BindAddr, conf.BindPort, err)
		}
	}
	t.tcpListener = tcpLn

	udpLn := options.udpListener
	if udpLn == nil {
		udpAddr := &net.UDPAddr{IP: ip, Port: conf.BindPort}

		var err error
		udpLn, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			return nil, fmt.Errorf("Failed to start UDP listener on %q port %d: %v", conf.BindAddr, conf.BindPort, err)
		}
		if err := setUDPRecvBuf(udpLn); err != nil {
			return nil, fmt.Errorf("Failed to resize UDP buffer: %v", err)
		}
	}
	t.udpListener = udpLn

	// Fire them up now that we've been able to create them all.
	t.wg.Add(2)
	go t.tcpListen(t.tcpListener)
	go t.udpListen(t.udpListener)

	ok = true
	return &t, nil
}

// GetAutoBindPort returns the bind port that was automatically given by the
// kernel, if a bind port of 0 was given.
func (t *transport) GetAutoBindPort() int {
	return t.tcpListener.Addr().(*net.TCPAddr).Port
}

// See Transport.
func (t *transport) FinalAdvertiseAddr(ip string, port int) (net.IP, int, error) {
	var advertiseAddr net.IP
	var advertisePort int
	if ip != "" {
		// If they've supplied an address, use that.
		advertiseAddr = net.ParseIP(ip)
		if advertiseAddr == nil {
			return nil, 0, fmt.Errorf("Failed to parse advertise address %q", ip)
		}

		// Ensure IPv4 conversion if necessary.
		if ip4 := advertiseAddr.To4(); ip4 != nil {
			advertiseAddr = ip4
		}
		advertisePort = port
	} else {
		// Otherwise, if we're not bound to a specific IP, let's
		// use a suitable private IP address.
		var err error
		ip, err = sockaddr.GetPrivateIP()
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to get interface addresses: %v", err)
		}
		if ip == "" {
			return nil, 0, fmt.Errorf("No private IP address found, and explicit IP not provided")
		}

		advertiseAddr = net.ParseIP(ip)
		if advertiseAddr == nil {
			return nil, 0, fmt.Errorf("Failed to parse advertise address: %q", ip)
		}

		// Use the port we are bound to.
		advertisePort = t.GetAutoBindPort()

	}

	return advertiseAddr, advertisePort, nil
}

// See Transport.
func (t *transport) WriteTo(b []byte, addr string) (time.Time, error) {
	a := memberlist.Address{Addr: addr, Name: ""}
	return t.WriteToAddress(b, a)
}

// See NodeAwareTransport.
func (t *transport) WriteToAddress(b []byte, a memberlist.Address) (time.Time, error) {
	addr := a.Addr

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return time.Time{}, err
	}

	// We made sure there's at least one UDP listener, so just use the
	// packet sending interface on the first one. Take the time after the
	// write call comes back, which will underestimate the time a little,
	// but help account for any delays before the write occurs.
	_, err = t.udpListener.WriteTo(b, udpAddr)
	return time.Now(), err
}

// See Transport.
func (t *transport) PacketCh() <-chan *memberlist.Packet {
	return t.packetCh
}

// See IngestionAwareTransport.
func (t *transport) IngestPacket(conn net.Conn, addr net.Addr, now time.Time, shouldClose bool) error {
	if shouldClose {
		defer conn.Close()
	}

	// Copy everything from the stream into packet buffer.
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, conn); err != nil {
		return fmt.Errorf("failed to read packet: %v", err)
	}

	// Check the length - it needs to have at least one byte to be a proper
	// message. This is checked elsewhere for writes coming in directly from
	// the UDP socket.
	if n := buf.Len(); n < 1 {
		return fmt.Errorf("packet too short (%d bytes) %s", n, memberlist.LogAddress(addr))
	}

	// Inject the packet.
	t.packetCh <- &memberlist.Packet{
		Buf:       buf.Bytes(),
		From:      addr,
		Timestamp: now,
	}
	return nil
}

// See Transport.
func (t *transport) DialTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	a := memberlist.Address{Addr: addr, Name: ""}
	return t.DialAddressTimeout(a, timeout)
}

// See NodeAwareTransport.
func (t *transport) DialAddressTimeout(a memberlist.Address, timeout time.Duration) (net.Conn, error) {
	addr := a.Addr

	dialer := net.Dialer{Timeout: timeout}
	return dialer.Dial("tcp", addr)
}

// See Transport.
func (t *transport) StreamCh() <-chan net.Conn {
	return t.streamCh
}

// See IngestionAwareTransport.
func (t *transport) IngestStream(conn net.Conn) error {
	t.streamCh <- conn
	return nil
}

// See Transport.
func (t *transport) Shutdown() error {
	// This will avoid log spam about errors when we shut down.
	atomic.StoreInt32(&t.shutdown, 1)

	// Rip through all the connections and shut them down.
	t.tcpListener.Close()
	t.udpListener.Close()

	// Block until all the listener threads have died.
	t.wg.Wait()
	return nil
}

// tcpListen is a long running goroutine that accepts incoming TCP connections
// and hands them off to the stream channel.
func (t *transport) tcpListen(tcpLn *net.TCPListener) {
	defer t.wg.Done()

	// baseDelay is the initial delay after an AcceptTCP() error before attempting again
	const baseDelay = 5 * time.Millisecond

	// maxDelay is the maximum delay after an AcceptTCP() error before attempting again.
	// In the case that tcpListen() is error-looping, it will delay the shutdown check.
	// Therefore, changes to maxDelay may have an effect on the latency of shutdown.
	const maxDelay = 1 * time.Second

	var loopDelay time.Duration
	for {
		conn, err := tcpLn.AcceptTCP()
		if err != nil {
			if s := atomic.LoadInt32(&t.shutdown); s == 1 {
				break
			}

			if loopDelay == 0 {
				loopDelay = baseDelay
			} else {
				loopDelay *= 2
			}

			if loopDelay > maxDelay {
				loopDelay = maxDelay
			}

			t.logger.Printf("[ERR] memberlist: Error accepting TCP connection: %v", err)
			time.Sleep(loopDelay)
			continue
		}
		// No error, reset loop delay
		loopDelay = 0

		t.streamCh <- conn
	}
}

// udpListen is a long running goroutine that accepts incoming UDP packets and
// hands them off to the packet channel.
func (t *transport) udpListen(udpLn *net.UDPConn) {
	defer t.wg.Done()
	for {
		// Do a blocking read into a fresh buffer. Grab a time stamp as
		// close as possible to the I/O.
		buf := make([]byte, udpPacketBufSize)
		n, addr, err := udpLn.ReadFrom(buf)
		ts := time.Now()
		if err != nil {
			if s := atomic.LoadInt32(&t.shutdown); s == 1 {
				break
			}

			t.logger.Printf("[ERR] memberlist: Error reading UDP packet: %v", err)
			continue
		}

		// Check the length - it needs to have at least one byte to be a
		// proper message.
		if n < 1 {
			t.logger.Printf("[ERR] memberlist: UDP packet too short (%d bytes) %s",
				len(buf), memberlist.LogAddress(addr))
			continue
		}

		// Ingest the packet.
		metrics.IncrCounterWithLabels([]string{"memberlist", "udp", "received"}, float32(n), t.metricLabels)
		t.packetCh <- &memberlist.Packet{
			Buf:       buf[:n],
			From:      addr,
			Timestamp: ts,
		}
	}
}

// setUDPRecvBuf is used to resize the UDP receive window. The function
// attempts to set the read buffer to `udpRecvBuf` but backs off until
// the read buffer can be set.
func setUDPRecvBuf(c *net.UDPConn) error {
	size := udpRecvBufSize
	var err error
	for size > 0 {
		if err = c.SetReadBuffer(size); err == nil {
			return nil
		}
		size = size / 2
	}
	return err
}
