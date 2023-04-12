package fcm

import (
	"fmt"
)

type FCM struct {
	server *Server
}

func NewFCM(addr string, port int, opts ...Option) (*FCM, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	logger := options.logger
	logger.Info("starting fcm")

	server, err := NewServer(
		addr,
		port,
		WithLogger(options.logger),
		WithListener(options.listener),
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
