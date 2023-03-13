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

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/fuddle-io/fuddle/pkg/registry/cluster"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Server exposes endpoints for nodes to register themselves and inspect the
// registry.
type Server struct {
	cluster *cluster.Cluster

	listener   net.Listener
	httpServer *http.Server

	logger *zap.Logger
}

func NewServer(addr string, cluster *cluster.Cluster, opts ...Option) *Server {
	options := options{
		logger:   zap.NewNop(),
		listener: nil,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	server := &Server{
		cluster:  cluster,
		listener: options.listener,
		logger:   options.logger,
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/register", server.registerRoute)
	r.HandleFunc("/api/v1/cluster", server.clusterRoute)
	r.HandleFunc("/api/v1/node/{id}", server.nodeRoute)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	server.httpServer = httpServer

	return server
}

// Start starts the server on the given listener.
func (s *Server) Start() error {
	ln := s.listener
	if ln == nil {
		// Setup the listener before starting to the goroutine to return any errors
		// binding or listening to the configured address.
		var err error
		ln, err = net.Listen("tcp", s.httpServer.Addr)
		if err != nil {
			return fmt.Errorf("registry server: %w", err)
		}
	}

	s.logger.Info(
		"starting registry server",
		zap.String("addr", ln.Addr().String()),
	)

	go func() {
		if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			s.logger.Error("registry serve", zap.Error(err))
		}
	}()

	return nil
}

// GracefulStop closes the server and gracefully sheds connections.
// TODO(AD) Doesn't yet handle gracefully closing websocket connections.
func (s *Server) GracefulStop() {
	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		s.logger.Error("registry server stop", zap.Error(err))
	}
}

func (s *Server) registerRoute(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) clusterRoute(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(s.cluster.Nodes()); err != nil {
		s.logger.Error("failed to encode cluster response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) nodeRoute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	node, ok := s.cluster.Node(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(node); err != nil {
		s.logger.Error("failed to encode node response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
