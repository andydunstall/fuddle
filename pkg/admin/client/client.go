package admin

import (
	"context"
	"fmt"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	addr   string
	conn   *grpc.ClientConn
	client rpc.ClientReadRegistryClient
}

func Connect(addr string, opts ...Option) (*Client, error) {
	options := options{
		connectTimeout: time.Second,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	conn, err := connect(addr, options.connectTimeout)
	if err != nil {
		return nil, fmt.Errorf("admin client: connect: %w", err)
	}
	return &Client{
		addr:   addr,
		conn:   conn,
		client: rpc.NewClientReadRegistryClient(conn),
	}, nil
}

func (c *Client) Members(ctx context.Context) ([]*rpc.Member2, error) {
	resp, err := c.client.Members(ctx, &rpc.MembersRequest{})
	if err != nil {
		return nil, fmt.Errorf("admin client: members: %w", err)
	}
	return resp.Members, nil
}

func (c *Client) Member(ctx context.Context, id string) (*rpc.Member2, error) {
	resp, err := c.client.Member(ctx, &rpc.MemberRequest{
		Id: id,
	})
	if err != nil {
		return nil, fmt.Errorf("admin client: member: %w", err)
	}
	if resp.Member == nil {
		return nil, fmt.Errorf("admin client: member: not found")
	}

	return resp.Member, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func connect(addr string, timeout time.Duration) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	return conn, nil
}
