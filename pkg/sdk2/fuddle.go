package fuddle

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/sdk2/resolvers"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/keepalive"
)

type Fuddle struct {
	connectAttemptTimeout time.Duration
	keepAlivePingInterval time.Duration
	keepAlivePingTimeout  time.Duration

	onConnectionStateChange func(state ConnState)

	registry *registry

	conn   *grpc.ClientConn
	client rpc.RegistryClient

	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup
	closed *atomic.Bool

	logger              *zap.Logger
	grpcLoggerVerbosity int
}

func Connect(ctx context.Context, addrs []string, opts ...Option) (*Fuddle, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	cancelCtx, cancel := context.WithCancel(context.Background())
	f := &Fuddle{
		connectAttemptTimeout: options.connectAttemptTimeout,
		keepAlivePingInterval: options.keepAlivePingInterval,
		keepAlivePingTimeout:  options.keepAlivePingTimeout,

		onConnectionStateChange: options.onConnectionStateChange,

		registry: newRegistry(options.logger),

		ctx:    cancelCtx,
		cancel: cancel,
		closed: atomic.NewBool(false),

		logger:              options.logger,
		grpcLoggerVerbosity: options.grpcLoggerVerbosity,
	}
	if err := f.connect(ctx, addrs); err != nil {
		return nil, fmt.Errorf("fuddle: %w", err)
	}

	return f, nil
}

func (f *Fuddle) Members() []Member {
	return f.registry.Members()
}

func (f *Fuddle) Subscribe(cb func()) func() {
	return f.registry.Subscribe(cb)
}

func (f *Fuddle) Close() {
	f.closed.Store(true)
	f.cancel()
	f.conn.Close()
}

func (f *Fuddle) connect(ctx context.Context, addrs []string) error {
	if f.grpcLoggerVerbosity > 0 {
		grpclog.SetLoggerV2(grpclog.NewLoggerV2WithVerbosity(
			os.Stderr, os.Stderr, os.Stderr, f.grpcLoggerVerbosity,
		))
	}

	if len(addrs) == 0 {
		f.logger.Error("failed to connect: no seed addresses")
		return fmt.Errorf("connect: no seeds addresses")
	}

	// Since we use a 'first pick' load balancer, shuffle the addrs so multiple
	// clients with the same addrs don't all try the same node.
	shuffleStrings(addrs)

	f.logger.Info("connecting", zap.Strings("addrs", addrs))

	// Send keep alive pings to detect unresponsive connections and trigger
	// a reconnect.
	keepAliveParams := keepalive.ClientParameters{
		Time:                f.keepAlivePingInterval,
		Timeout:             f.keepAlivePingTimeout,
		PermitWithoutStream: true,
	}
	conn, err := grpc.DialContext(
		ctx,
		// Use the static resolver which uses the configured seed addresses.
		"static:///fuddle",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithResolvers(resolvers.NewStaticResolverBuilder(addrs)),
		// Add a custom dialer so we can set a per connection attempt timeout.
		grpc.WithContextDialer(f.dialerWithTimeout),
		// Block until the connection succeeds so we can fail the initial
		// connection.
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepAliveParams),
	)
	if err != nil {
		f.logger.Error(
			"failed to connect",
			zap.Strings("seeds", addrs),
			zap.Error(err),
		)
		return fmt.Errorf("connect: %w", err)
	}

	f.conn = conn
	f.client = rpc.NewRegistryClient(conn)

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		f.monitorConnection()
	}()

	return nil
}

// monitorConnection detects disconnects and reconnects.
func (f *Fuddle) monitorConnection() {
	for {
		s := f.conn.GetState()
		if s == connectivity.Ready {
			f.onConnected()
		} else {
			f.conn.Connect()
		}

		if !f.conn.WaitForStateChange(f.ctx, s) {
			// Only returns if the client is closed.
			return
		}

		// If we were ready but now the state has changed we must have
		// droped the connection.
		if s == connectivity.Ready {
			f.onDisconnect()
		}
	}
}

func (f *Fuddle) onConnected() {
	f.logger.Info("connected")

	if f.onConnectionStateChange != nil {
		f.onConnectionStateChange(StateConnected)
	}

	subscription, err := f.client.Subscribe(
		context.Background(),
		&rpc.SubscribeRequest{
			KnownMembers: f.registry.KnownVersions(),
			// Receive all updates from the connected node..
			OwnerOnly: false,
		},
	)
	if err != nil {
		// If we can't subscribe, this will typically mean we've disconnected
		// so will retry once reconnected.
		f.logger.Warn("failed to subscribe", zap.Error(err))
		return
	}

	f.wg.Add(1)
	go func() {
		go f.streamUpdates(subscription)
	}()
}

func (f *Fuddle) onDisconnect() {
	f.logger.Info("disconnected")

	if f.onConnectionStateChange != nil {
		f.onConnectionStateChange(StateDisconnected)
	}
}

func (f *Fuddle) streamUpdates(stream rpc.Registry_SubscribeClient) {
	for {
		update, err := stream.Recv()
		if err != nil {
			// Avoid redundent logs if we've closed.
			if f.closed.Load() {
				return
			}
			f.logger.Warn("subscribe error", zap.Error(err))
			return
		}

		f.registry.RemoteUpdate(update)
	}
}

func (f *Fuddle) dialerWithTimeout(ctx context.Context, addr string) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: f.connectAttemptTimeout,
	}
	return dialer.DialContext(ctx, "tcp", addr)
}

func shuffleStrings(s []string) {
	for i := range s {
		j := rand.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
}
