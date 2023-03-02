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

package frontend

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"

	"github.com/andydunstall/fuddle/pkg/client"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type server struct {
	httpServer *http.Server

	registry *client.Registry

	logger *zap.Logger
}

func newServer(addr string, logger *zap.Logger) *server {
	server := &server{
		logger: logger,
	}

	r := mux.NewRouter()
	r.HandleFunc("/random", server.randomRoute)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	server.httpServer = httpServer

	return server
}

func (s *server) Start(registry *client.Registry) error {
	s.registry = registry

	// Setup the listener before starting to the goroutine to return any errors
	// binding or listening to the configured address.
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("frontend server: %w", err)
	}

	s.logger.Info(
		"starting frontend server",
		zap.String("addr", s.httpServer.Addr),
	)

	go func() {
		if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			s.logger.Error("frontend serve error", zap.Error(err))
		}
	}()

	return nil
}

func (s *server) GracefulStop() {
	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		s.logger.Error("failed to shut down frontend server", zap.Error(err))
	}
}

func (s *server) randomRoute(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.registry.Nodes(context.Background())
	if err != nil {
		s.logger.Error("failed to query registry", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}

	var randomServiceAddrs []string
	for _, node := range nodes {
		if node.Service == "random" {
			randomServiceAddrs = append(randomServiceAddrs, node.State["addr"])
		}
	}
	if len(randomServiceAddrs) == 0 {
		s.logger.Error("could not find any random service nodes")
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}

	addr := randomServiceAddrs[rand.Int()%len(randomServiceAddrs)]
	resp, err := http.Get(fmt.Sprintf("http://%s", addr))
	if err != nil {
		s.logger.Error("could not query random service nodes")
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	if _, err = io.Copy(w, resp.Body); err != nil {
		s.logger.Debug("failed to write response", zap.Error(err))
	}
}
