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

package random

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type server struct {
	httpServer *http.Server

	logger *zap.Logger
}

func newServer(addr string, logger *zap.Logger) *server {
	server := &server{
		logger: logger,
	}

	r := mux.NewRouter()
	r.HandleFunc("/", server.randomRoute)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	server.httpServer = httpServer

	return server
}

func (s *server) Start() error {
	// Setup the listener before starting to the goroutine to return any errors
	// binding or listening to the configured address.
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("random server: %w", err)
	}

	s.logger.Info(
		"starting random server",
		zap.String("addr", s.httpServer.Addr),
	)

	go func() {
		if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			s.logger.Error("random serve error", zap.Error(err))
		}
	}()

	return nil
}

func (s *server) GracefulStop() {
	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		s.logger.Error("failed to shut down random server", zap.Error(err))
	}
}

func (s *server) randomRoute(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte(strconv.Itoa(rand.Int()))); err != nil {
		s.logger.Debug("failed to write response", zap.Error(err))
	}
}
