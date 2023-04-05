package registry

import (
	"context"
	"fmt"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	addr string

	conn   *grpc.ClientConn
	client rpc.RegistryClient

	cancelCtx context.Context
	cancel    func()

	onConnectionStateChange func(state ConnState)

	logger *zap.Logger
}

// Connect will setup a connection to the given address.
//
// If the client cannot connect, or the connection drops, the client will keep
// trying to reconnect in the background until it is closed.
func Connect(addr string, opts ...ClientOption) (*Client, error) {
	options := defaultClientOptions()
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
	c := &Client{
		addr:                    addr,
		conn:                    conn,
		cancelCtx:               ctx,
		cancel:                  cancel,
		client:                  rpc.NewRegistryClient(conn),
		onConnectionStateChange: options.onConnectionStateChange,
		logger:                  options.logger,
	}

	go c.monitorConnection()

	return c, nil
}

func (c *Client) Close() {
	c.logger.Info("closing")
	c.cancel()
	c.conn.Close()
}

// monitorConnection handles connection disconnects and reconnects.
func (c *Client) monitorConnection() {
	for {
		s := c.conn.GetState()
		if s == connectivity.Ready {
			c.onConnected()
		} else {
			c.conn.Connect()
		}

		if !c.conn.WaitForStateChange(c.cancelCtx, s) {
			// Only returns if the client is closed.
			return
		}

		// If we were ready but now the state has changed we must have
		// droped the connection.
		if s == connectivity.Ready {
			c.onDisconnected()
		}
	}
}

func (c *Client) onConnected() {
	c.logger.Info("connected")

	if c.onConnectionStateChange != nil {
		c.onConnectionStateChange(StateConnected)
	}

	stream, err := c.client.Subscribe(
		context.Background(),
		&rpc.SubscribeRequest{
			OwnerOnly: true,
		},
	)
	if err != nil {
		// If subscribe fails, the connection is likely already closed, so
		// it will be retried once connected.
		c.logger.Warn("subscribe failed", zap.Error(err))
		return
	}

	// streamUpdates will exit when the connection is closed.
	go c.streamUpdates(stream)
}

func (c *Client) onDisconnected() {
	c.logger.Info("disconnected")

	if c.onConnectionStateChange != nil {
		c.onConnectionStateChange(StateDisconnected)
	}
}

func (c *Client) streamUpdates(stream rpc.Registry_SubscribeClient) {
	for {
		update, err := stream.Recv()
		if err != nil {
			c.logger.Warn("stream error", zap.Error(err))
			return
		}

		c.logger.Debug(
			"stream update",
			zap.String("id", update.Id),
			zap.String("type", update.UpdateType.String()),
		)
	}
}
