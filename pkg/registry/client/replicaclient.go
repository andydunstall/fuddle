package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registry/registry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ReplicaClient struct {
	addr string

	registry *registry.Registry

	conn   *grpc.ClientConn
	client rpc.ReplicaReadRegistryClient

	cancelCtx context.Context
	cancel    func()

	pending   []*rpc.Member2
	pendingMu sync.Mutex

	logger *zap.Logger
}

func ReplicaConnect(addr string, registry *registry.Registry, opts ...Option) (*ReplicaClient, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	// Note this won't actually connect to the server so should not fail.
	conn, err := grpc.DialContext(
		// Use background context as this isn't actually connecting but just
		// doing setup.
		context.Background(),
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("registry client: dial: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &ReplicaClient{
		addr:      addr,
		registry:  registry,
		conn:      conn,
		cancelCtx: ctx,
		cancel:    cancel,
		client:    rpc.NewReplicaReadRegistryClient(conn),
		logger:    options.logger,
	}
	go c.sendLoop()
	return c, nil
}

func (c *ReplicaClient) Update(member *rpc.Member2) {
	// TODO(AD) this must not block, instead queue up, and if queue gets full
	// will have to drop messages (which will be fixed by read repair)
	// TODO(AD) add retries with backoff (order doesn't matter)
	c.pendingMu.Lock()
	defer c.pendingMu.Unlock()

	c.pending = append(c.pending, member)
}

func (c *ReplicaClient) Close() {
	c.logger.Info("closing")
	c.cancel()
	c.conn.Close()
}

func (c *ReplicaClient) sendLoop() {
	// TODO(AD) for now just poll
	for {
		select {
		case <-c.cancelCtx.Done():
			return
		case <-time.After(time.Millisecond * 100):
		}

		c.pendingMu.Lock()
		pending := c.pending
		c.pending = nil
		c.pendingMu.Unlock()

		for _, m := range pending {
			if _, err := c.client.Update(context.Background(), &rpc.UpdateRequest{
				Member: m,
			}); err != nil {
				c.logger.Warn("failed to send update", zap.Error(err))
			}
		}
	}
}
