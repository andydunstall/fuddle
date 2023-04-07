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
	"github.com/fuddle-io/fuddle/pkg/sdk/resolvers"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
)

type Fuddle struct {
	connectAttemptTimeout time.Duration
	keepAlivePingInterval time.Duration
	keepAlivePingTimeout  time.Duration

	onConnectionStateChange func(state ConnState)

	registry *registry

	conn   *grpc.ClientConn
	client rpc.RegistryClient

	// cancel is a function called when the client is shutdown to stop any
	// waiting contexts.
	cancelCtx context.Context
	cancel    func()
	wg        sync.WaitGroup
	closed    *atomic.Bool

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

		registry: newRegistry(),

		cancelCtx: cancelCtx,
		cancel:    cancel,
		closed:    atomic.NewBool(false),

		logger:              options.logger,
		grpcLoggerVerbosity: options.grpcLoggerVerbosity,
	}

	if err := f.connect(ctx, addrs); err != nil {
		return nil, fmt.Errorf("fuddle: %w", err)
	}

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		f.monitorConnection()
	}()

	return f, nil
}

func (f *Fuddle) Members() []Member {
	return f.registry.Members()
}

func (f *Fuddle) Subscribe(cb func()) func() {
	return f.registry.Subscribe(cb)
}

func (f *Fuddle) Register(ctx context.Context, member Member) error {
	if member.Metadata == nil {
		member.Metadata = make(map[string]string)
	}

	stream, err := f.client.Register(context.Background())
	if err != nil {
		return err
	}

	rpcMember := member.toRPC()
	if err = stream.Send(&rpc.ClientUpdate{
		UpdateType: rpc.ClientUpdateType_CLIENT_REGISTER,
		Member:     rpcMember,
	}); err != nil {
		return err
	}

	f.registry.RegisterLocal(rpcMember)

	f.logger.Debug("member registered", zap.String("id", member.ID))

	// TODO start goroutine to send heartbeats

	return nil
}

func (f *Fuddle) connect(ctx context.Context, seeds []string) error {
	if f.grpcLoggerVerbosity > 0 {
		grpclog.SetLoggerV2(grpclog.NewLoggerV2WithVerbosity(
			os.Stderr, os.Stderr, os.Stderr, f.grpcLoggerVerbosity,
		))
	}

	if len(seeds) == 0 {
		f.logger.Error("failed to connect: no seed addresses")
		return fmt.Errorf("connect: no seeds addresses")
	}

	// Since we use a 'first pick' load balancer, shuffle the seeds so multiple
	// clients with the same seeds don't all try the same node.
	for i := range seeds {
		j := rand.Intn(i + 1)
		seeds[i], seeds[j] = seeds[j], seeds[i]
	}

	f.logger.Info("connecting", zap.Strings("seeds", seeds))

	conn, err := grpc.DialContext(
		ctx,
		// Use the status resolver which uses the configured seed addresses.
		"static:///fuddle",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithResolvers(resolvers.NewStaticResolverBuilder(seeds)),
		// Add a custom dialer so we can set a per connection attempt timeout.
		grpc.WithContextDialer(f.dialerWithTimeout),
		// Block until the connection succeeds.
		grpc.WithBlock(),
	)
	if err != nil {
		f.logger.Error(
			"failed to connect",
			zap.Strings("seeds", seeds),
			zap.Error(err),
		)
		return fmt.Errorf("connect: %w", err)
	}

	f.conn = conn
	f.client = rpc.NewRegistryClient(conn)

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

		if !f.conn.WaitForStateChange(f.cancelCtx, s) {
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

	f.reenterLocalMembers(context.Background())

	subscribeStream, err := f.client.Subscribe(
		context.Background(), &rpc.SubscribeRequest{
			KnownMembers: f.registry.KnownVersions(),
			OwnerOnly:    false,
		},
	)
	if err != nil {
		f.logger.Warn("create stream subscribe error", zap.Error(err))
	} else {
		// Start streaming updates. If the connection closes streamUpdates will
		// exit.
		go f.streamUpdates(subscribeStream)
	}
}

func (f *Fuddle) onDisconnect() {
	f.logger.Info("disconnected")

	if f.onConnectionStateChange != nil {
		f.onConnectionStateChange(StateDisconnected)
	}
}

func (f *Fuddle) dialerWithTimeout(ctx context.Context, addr string) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: f.connectAttemptTimeout,
	}
	return dialer.DialContext(ctx, "tcp", addr)
}

func (f *Fuddle) reenterLocalMembers(ctx context.Context) {
	f.logger.Debug(
		"reregistering members",
		zap.Strings("members", f.registry.LocalMemberIDs()),
	)

	for _, member := range f.registry.LocalMembers() {
		rpcMember := member.toRPC()

		stream, err := f.client.Register(context.Background())
		if err != nil {
			f.logger.Error(
				"failed to reregister member",
				zap.String("id", member.ID),
				zap.Error(err),
			)
			continue
		}

		if err = stream.Send(&rpc.ClientUpdate{
			UpdateType: rpc.ClientUpdateType_CLIENT_REGISTER,
			Member:     rpcMember,
		}); err != nil {
			f.logger.Error(
				"failed to reregister member",
				zap.String("id", member.ID),
				zap.Error(err),
			)
			continue
		}

		// TODO heartbeats

		f.logger.Debug("member re-registered", zap.String("id", member.ID))
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

		f.logger.Debug(
			"received update",
			zap.String("id", update.Member.Id),
			zap.String("update-type", update.UpdateType.String()),
		)

		f.registry.ApplyRemoteUpdate(update)
	}
}
