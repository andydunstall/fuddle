package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/metrics"
	"github.com/fuddle-io/fuddle/pkg/registry/registry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ReplicaClientMetrics struct {
	ReplicaUpdatesOutbound *metrics.Counter
}

func NewReplicaClientMetrics() *ReplicaClientMetrics {
	return &ReplicaClientMetrics{
		ReplicaUpdatesOutbound: metrics.NewCounter(
			"registry",
			"replica.updates.outbound",
			[]string{"target", "status"},
			"Number of outbound updates send to a replica",
		),
	}
}

func (m *ReplicaClientMetrics) Register(collector metrics.Collector) {
	collector.AddCounter(m.ReplicaUpdatesOutbound)
}

// ReplicaClient is used to make RPCs to other Fuddle nodes in the cluster.
//
// If the connection to the replica drops, the client will keep trying to
// reconnect until it is closed.
type ReplicaClient struct {
	targetID string
	registry *registry.Registry

	digestLimit int

	pending *pendingUpdates

	updateTimeout time.Duration

	conn   *grpc.ClientConn
	client rpc.ReplicaRegistry2Client

	ctx    context.Context
	cancel func()

	wg sync.WaitGroup

	metrics *ReplicaClientMetrics
	logger  *zap.Logger
}

func ReplicaConnect(addr string, targetID string, registry *registry.Registry, metrics *ReplicaClientMetrics, opts ...Option) (*ReplicaClient, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	var retryPolicy = `{
		"methodConfig": [{
			"name": [{"service": "registry.ReplicaRegistry2", "method": "Update"}],
			"waitForReady": true,

			"retryPolicy": {
				"MaxAttempts": 5,
				"InitialBackoff": ".2s",
				"MaxBackoff": "10s",
				"BackoffMultiplier": 2.0,
				"RetryableStatusCodes": [ "UNAVAILABLE" ]
			}
		}]
	}`
	// Dial won't connect yet so should never fail.
	conn, err := grpc.Dial(
		addr,
		grpc.WithDefaultServiceConfig(retryPolicy),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("replica client: connect: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &ReplicaClient{
		targetID:      targetID,
		registry:      registry,
		digestLimit:   options.digestLimit,
		pending:       newPendingUpdates(options.pendingUpdatesLimit),
		updateTimeout: options.updateTimeout,
		conn:          conn,
		client:        rpc.NewReplicaRegistry2Client(conn),
		ctx:           ctx,
		cancel:        cancel,
		metrics:       metrics,
		logger:        options.logger,
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.sendLoop()
	}()

	return c, nil
}

// Update forwards the given member update to the connected replica.
//
// This sends the update in the background to avoid blocking. If the number of
// pending updates exceeds pendingUpdatesLimit, the older updates are dropped.
// Therefore if the client cannot connector for a long time, updates may be
// dropped and have to be repaired by replica repair.
func (c *ReplicaClient) Update(u *rpc.Member2) {
	c.pending.Push(u)
}

func (c *ReplicaClient) Sync(ctx context.Context) error {
	resp, err := c.client.Sync(ctx, &rpc.ReplicaSyncRequest{
		Digest: c.registry.Digest(c.digestLimit),
	})
	if err != nil {
		return fmt.Errorf("replica client: client: sync: %w", err)
	}

	c.logger.Info(
		"replica sync",
		zap.String("target", c.targetID),
		zap.Int("digest-len", len(resp.Members)),
	)

	for _, m := range resp.Members {
		c.registry.RemoteUpdate(m)
	}

	return nil
}

func (c *ReplicaClient) Close() {
	c.cancel()
	c.pending.Close()
	c.wg.Wait()
}

func (c *ReplicaClient) sendLoop() {
	for {
		m, ok := c.pending.Take()
		if !ok {
			// Client closed.
			return
		}

		// Update will keep retrying for until cancelled. If it still does not
		// succeed, the update will be dropped and the replica will get the
		// update via read repair when it comes back.
		ctx, cancel := context.WithTimeout(c.ctx, c.updateTimeout)
		defer cancel()
		if _, err := c.client.Update(ctx, &rpc.UpdateRequest{
			Member:       m,
			SourceNodeId: c.registry.LocalID(),
		}); err != nil {
			c.logger.Warn(
				"failed to forward update",
				zap.String("member-id", m.State.Id),
				zap.String("target", c.targetID),
				zap.Error(err),
			)

			c.metrics.ReplicaUpdatesOutbound.Inc(map[string]string{
				"target": c.targetID,
				"status": "fail",
			})
		} else {
			c.metrics.ReplicaUpdatesOutbound.Inc(map[string]string{
				"target": c.targetID,
				"status": "ok",
			})
		}
	}
}

type pendingUpdates struct {
	limit int

	pending []*rpc.Member2
	closed  bool

	cv *sync.Cond

	// mu protects the fields above.
	mu *sync.Mutex
}

func newPendingUpdates(limit int) *pendingUpdates {
	mu := &sync.Mutex{}
	return &pendingUpdates{
		limit:  limit,
		cv:     sync.NewCond(mu),
		closed: false,
		mu:     mu,
	}
}

func (p *pendingUpdates) Push(u *rpc.Member2) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If the number of pending items exceeds the limit, drop the older updates.
	// These updates will dropped updates be repaired by replica repair.
	if len(p.pending) > p.limit {
		p.pending = p.pending[1:]
	}

	if p.closed {
		return
	}

	p.pending = append(p.pending, u)
	p.cv.Signal()
}

// Take returns the next pending update and removes it, or false if the client
// is closed.
func (p *pendingUpdates) Take() (*rpc.Member2, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Since we can miss signals, only block if empty.
	for len(p.pending) == 0 && !p.closed {
		p.cv.Wait()
	}

	if p.closed {
		return nil, false
	}

	u := p.pending[0]
	p.pending = p.pending[1:]
	return u, true
}

func (p *pendingUpdates) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true
	// Signal the write loop so it closes.
	p.cv.Signal()
}
