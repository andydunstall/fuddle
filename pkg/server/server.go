package server

import (
	"fmt"
	"net"
	"time"

	"github.com/fuddle-io/fuddle/pkg/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type Server struct {
	conf       *config.Config
	ln         net.Listener
	grpcServer *grpc.Server
	logger     *zap.Logger
}

func NewServer(conf *config.Config, opts ...Option) *Server {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(&options)
	}

	logger := options.logger

	enforcementPolicy := keepalive.EnforcementPolicy{
		MinTime:             10 * time.Second,
		PermitWithoutStream: true,
	}
	grpcServer := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(enforcementPolicy),
	)

	return &Server{
		conf:       conf,
		ln:         options.listener,
		grpcServer: grpcServer,
		logger:     logger,
	}
}

func (s *Server) GRPCServer() *grpc.Server {
	return s.grpcServer
}

func (s *Server) Serve() error {
	ln := s.ln
	if ln == nil {
		ip := net.ParseIP(s.conf.RPC.BindAddr)
		tcpAddr := &net.TCPAddr{IP: ip, Port: s.conf.RPC.BindPort}

		var err error
		ln, err = net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			s.logger.Info(
				"failed to start listener",
				zap.String("addr", ln.Addr().String()),
				zap.Error(err),
			)
			return fmt.Errorf("server: start listener: %w", err)
		}
	}

	s.logger.Info("starting grpc server", zap.String("addr", ln.Addr().String()))

	go func() {
		if err := s.grpcServer.Serve(ln); err != nil {
			s.logger.Error("grpc serve", zap.Error(err))
		}
	}()
	return nil
}

func (s *Server) Shutdown() {
	s.logger.Info("stopping grpc server")
	s.grpcServer.Stop()
}
