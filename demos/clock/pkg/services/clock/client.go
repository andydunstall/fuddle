package clock

import (
	"context"
	"fmt"

	fuddle "github.com/fuddle-io/fuddle-go"
	fuddleResolver "github.com/fuddle-io/fuddle/demos/clock/pkg/resolver/fuddle"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	registry *fuddle.Fuddle
	conn     *grpc.ClientConn
	client   ClockClient
}

func NewClient(registry *fuddle.Fuddle) (*Client, error) {
	var retryPolicy = `{
		"methodConfig": [{
			"name": [{"service": "clock.Clock"}],
			"waitForReady": true,

			"retryPolicy": {
				"MaxAttempts": 4,
				"InitialBackoff": ".1s",
				"MaxBackoff": "5s",
				"BackoffMultiplier": 2.0,
				"RetryableStatusCodes": [ "UNAVAILABLE" ]
			}
		}]
	}`
	conn, err := grpc.DialContext(
		context.Background(),
		"fuddle:///clock",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithResolvers(fuddleResolver.NewBuilder(registry, "clock")),
		grpc.WithDefaultServiceConfig(retryPolicy),
	)
	if err != nil {
		return nil, fmt.Errorf("clock client: %w", err)
	}

	return &Client{
		registry: registry,
		conn:     conn,
		client:   NewClockClient(conn),
	}, nil
}

func (c *Client) Time(ctx context.Context) (int64, error) {
	resp, err := c.client.Time(ctx, &TimeRequest{})
	if err != nil {
		return 0, fmt.Errorf("clock client: time: %w", err)
	}
	return resp.Time, nil
}

func (c *Client) Close() {
	c.conn.Close()
}
