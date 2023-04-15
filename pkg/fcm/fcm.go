package fcm

import (
	"fmt"

	"github.com/fuddle-io/fuddle/pkg/fcm/server"
)

type FCM struct {
	server *server.Server
}

func NewFCM(addr string, port int, opts ...Option) (*FCM, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	logger := options.logger
	logger.Info("starting fcm")

	server, err := server.NewServer(
		addr,
		port,
		server.WithLogger(options.logger),
		server.WithListener(options.listener),
	)
	if err != nil {
		return nil, fmt.Errorf("fcm: %w", err)
	}
	return &FCM{
		server: server,
	}, nil
}

func (fcm *FCM) Shutdown() {
	fcm.server.Shutdown()
}
