package fuddle

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/fuddle-io/fuddle/pkg/sdk2/resolvers"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
)

type Fuddle struct {
	connectAttemptTimeout time.Duration

	conn *grpc.ClientConn

	logger              *zap.Logger
	grpcLoggerVerbosity int
}

func Connect(ctx context.Context, addrs []string, opts ...Option) (*Fuddle, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	f := &Fuddle{
		connectAttemptTimeout: options.connectAttemptTimeout,
		logger:                options.logger,
		grpcLoggerVerbosity:   options.grpcLoggerVerbosity,
	}
	if err := f.connect(ctx, addrs); err != nil {
		return nil, fmt.Errorf("fuddle: %w", err)
	}

	return f, nil
}

func (f *Fuddle) Close() {
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

	return nil
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
