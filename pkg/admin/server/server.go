package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server

	logger *zap.Logger
}

func NewServer(conf *config.Config, opts ...Option) (*Server, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	mux := http.NewServeMux()
	if options.collector != nil {
		http.Handle(
			"/metrics",
			promhttp.HandlerFor(
				options.collector.Registry(),
				promhttp.HandlerOpts{
					Registry: options.collector.Registry(),
				},
			),
		)
	}

	s := &Server{
		logger: options.logger,
	}

	ln := options.listener
	if ln == nil {
		ip := net.ParseIP(conf.Admin.BindAddr)
		tcpAddr := &net.TCPAddr{IP: ip, Port: conf.Admin.BindPort}

		var err error
		ln, err = net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			s.logger.Info(
				"failed to start listener",
				zap.String("addr", ln.Addr().String()),
				zap.Error(err),
			)
			return nil, fmt.Errorf("admin server: start listener: %w", err)
		}
	}

	s.httpServer = &http.Server{
		Handler:           mux,
		Addr:              ln.Addr().String(),
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}
	go func() {
		if err := s.httpServer.Serve(ln); err != nil {
			s.logger.Error("http serve error", zap.Error(err))
		}
	}()

	return s, nil
}

func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	s.httpServer.Shutdown(ctx)
}
