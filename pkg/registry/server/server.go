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
	"fmt"
	"net"
	"net/http"

	"github.com/fuddle-io/fuddle/pkg/registry/cluster"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Server exposes endpoints for nodes to register themselves and inspect the
// registry.
type Server struct {
	cluster *cluster.Cluster

	listener   net.Listener
	httpServer *http.Server
	upgrader   websocket.Upgrader

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

	httpServer := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	server.httpServer = httpServer
	server.upgrader = websocket.Upgrader{}

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

func (s *Server) Stop() {
	s.httpServer.Close()
}

func (s *Server) registerRoute(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	conn := newConn(ws, s.logger)
	defer conn.Close()

	// Wait for the connected node to register.
	registerUpdate, err := conn.RecvUpdate()
	if err != nil {
		return
	}
	// If the first update is not the node joining this is a protocol error.
	if registerUpdate.UpdateType != cluster.UpdateTypeRegister {
		s.logger.Warn("protocol error: node not registered")
		return
	}
	if err := s.cluster.ApplyUpdate(registerUpdate); err != nil {
		s.logger.Warn("apply update", zap.Error(err))
		return
	}

	nodeID := registerUpdate.ID

	// Subscribe to the node map and send updates to the client. This will
	// replay all existing nodes as JOIN updates to ensure the subscriber
	// doesn't miss any updates.
	unsubscribe := s.cluster.Subscribe(true, func(update *cluster.NodeUpdate) {
		// Avoid echoing back updates from the connected nodes.
		if update.ID == nodeID {
			return
		}

		conn.AddUpdate(update)
	})
	defer unsubscribe()

	for {
		update, err := conn.RecvUpdate()
		if err != nil {
			return
		}

		if err := s.cluster.ApplyUpdate(update); err != nil {
			s.logger.Warn("apply update", zap.Error(err))
		}
	}
}
