// Copyright (C) 2023 Andrew Dunstall
//
// Fuddle is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Fuddle is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package counter

import (
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// grpcServer runs the RPC server over gRPC.
type grpcServer struct {
	addr   string
	server *grpc.Server

	logger *zap.Logger
}

// newGRPCServer returns a new GRPC server.
func newGRPCServer(addr string, logger *zap.Logger) *grpcServer {
	return &grpcServer{
		addr:   addr,
		server: grpc.NewServer(),
		logger: logger,
	}
}

// Start listens for RPC requests in a background goroutine.
func (s *grpcServer) Start(ln net.Listener) error {
	if ln == nil {
		// Setup the listener before starting to the goroutine to return any errors
		// binding or listening to the configured address.
		var err error
		ln, err = net.Listen("tcp", s.addr)
		if err != nil {
			return fmt.Errorf("grpc server: %w", err)
		}
	}

	s.logger.Info(
		"starting grpc server",
		zap.String("addr", s.addr),
	)

	go func() {
		if err := s.server.Serve(ln); err != nil {
			s.logger.Error("grpc serve error", zap.Error(err))
		}
	}()

	return nil
}

func (s *grpcServer) GracefulStop() {
	s.logger.Info("starting grpc server graceful shutdown")
	s.server.GracefulStop()
}

// grpcServer returns the underlying GRPC server. Used to register handlers.
func (s *grpcServer) GRPCServer() *grpc.Server {
	return s.server
}
