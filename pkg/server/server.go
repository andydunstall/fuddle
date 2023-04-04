package server

import (
	"fmt"
	"net"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	logger     *zap.Logger
}

func NewServer(conf *config.Config, opts ...Option) (*Server, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(&options)
	}

	logger := options.logger

	grpcServer := grpc.NewServer()
	ln := options.listener
	if ln == nil {
		ip := net.ParseIP(conf.RPC.BindAddr)
		tcpAddr := &net.TCPAddr{IP: ip, Port: conf.RPC.BindPort}

		var err error
		ln, err = net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			logger.Info(
				"failed to start listener",
				zap.String("addr", ln.Addr().String()),
				zap.Error(err),
			)
			return nil, fmt.Errorf("server: start listener: %w", err)
		}
	}

	registryServer := newRegistryServer()
	rpc.RegisterRegistryServer(grpcServer, registryServer)

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
