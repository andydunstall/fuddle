package server

import (
	"fmt"
	"net"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/registry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	logger     *zap.Logger
}

func NewServer(conf *config.Config, registry *registry.Registry, opts ...Option) (*Server, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(&options)
	}

	logger := options.logger

	registryServer := newRegistryServer(registry)

	grpcServer := grpc.NewServer()
	rpc.RegisterRegistryServer(grpcServer, registryServer)

	ln := options.listener
	if ln == nil {
		ip := net.ParseIP(conf.Registry.BindAddr)
		tcpAddr := &net.TCPAddr{IP: ip, Port: conf.Registry.BindPort}

		var err error
		ln, err = net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			return nil, fmt.Errorf("server: start listener: %w", err)
		}
	}

	logger.Info("starting grpc server", zap.String("addr", ln.Addr().String()))

	go func() {
		if err := grpcServer.Serve(ln); err != nil {
			logger.Error("grpc serve", zap.Error(err))
		}
	}()

	return &Server{
		grpcServer: grpcServer,
		logger:     logger,
	}, nil
}

func (s *Server) Shutdown() {
	s.logger.Info("stopping grpc server")
	s.grpcServer.Stop()
}
