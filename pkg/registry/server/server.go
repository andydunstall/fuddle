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
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Server exposes endpoints for nodes to register themselves and inspect the
// registry.
type Server struct {
	cluster *cluster.Cluster

	listener   net.Listener
	httpServer *http.Server
	upgrader   websocket.Upgrader

	promRegistry          *prometheus.Registry
	nodeCountMetric       prometheus.Gauge
	updateCountMetric     *prometheus.CounterVec
	connectionCountMetric prometheus.Gauge

	logger *zap.Logger
}

func NewServer(addr string, cluster *cluster.Cluster, opts ...Option) *Server {
	options := options{
		logger:       zap.NewNop(),
		listener:     nil,
		promRegistry: nil,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	server := &Server{
		cluster:      cluster,
		listener:     options.listener,
		promRegistry: options.promRegistry,
		logger:       options.logger,
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/register", server.registerRoute)
	r.HandleFunc("/api/v1/cluster", server.clusterRoute)
	r.HandleFunc("/api/v1/node/{id}", server.nodeRoute)
	if options.promRegistry != nil {
		r.Handle(
			"/metrics",
			promhttp.HandlerFor(
				options.promRegistry,
				promhttp.HandlerOpts{Registry: options.promRegistry},
			),
		)

		nodeCountMetric := prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "fuddle_registry_node_count",
				Help: "Number of nodes in the registry.",
			},
		)
		options.promRegistry.MustRegister(nodeCountMetric)
		nodeCountMetric.Set(float64(len(cluster.Nodes())))
		server.nodeCountMetric = nodeCountMetric

		updateCountMetric := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fuddle_registry_update_count",
				Help: "Number of updates to the registry.",
			},
			[]string{"type"},
		)
		options.promRegistry.MustRegister(updateCountMetric)
		server.updateCountMetric = updateCountMetric

		connectionCountMetric := prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "fuddle_registry_connection_count",
				Help: "Number of connections to the registry.",
			},
		)
		options.promRegistry.MustRegister(connectionCountMetric)
		server.connectionCountMetric = connectionCountMetric
	}

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
	logger := s.logger.With(zap.String("path", r.URL.Path))

	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Debug("register: upgrade error", zap.Error(err))
		return
	}
	conn := newConn(ws, logger)
	defer conn.Close()

	if s.connectionCountMetric != nil {
		s.connectionCountMetric.Inc()
		defer s.connectionCountMetric.Dec()
	}

	// Wait for the connected node to register.
	registerUpdate, err := conn.RecvUpdate()
	if err != nil {
		logger.Debug("register: not received register update")
		return
	}
	// If the first update is not the node joining this is a protocol error.
	if registerUpdate.UpdateType != cluster.UpdateTypeRegister {
		logger.Warn("register: protocol error: update not a register")
		return
	}

	if s.updateCountMetric != nil {
		s.updateCountMetric.With(prometheus.Labels{
			"type": string(registerUpdate.UpdateType),
		}).Inc()
	}

	logger.Debug(
		"register: received update",
		zap.Object("update", registerUpdate),
	)

	if err := s.cluster.ApplyUpdate(registerUpdate); err != nil {
		logger.Warn("register: failed to apply register update", zap.Error(err))
		return
	}

	if s.nodeCountMetric != nil {
		s.nodeCountMetric.Set(float64(len(s.cluster.Nodes())))
	}

	nodeID := registerUpdate.ID

	logger = logger.With(zap.String("client-node", nodeID))

	// Subscribe to the node map and send updates to the client. This will
	// replay all existing nodes as JOIN updates to ensure the subscriber
	// doesn't miss any updates.
	unsubscribe := s.cluster.Subscribe(true, func(update *cluster.NodeUpdate) {
		// Avoid echoing back updates from the connected nodes.
		if update.ID == nodeID {
			return
		}

		logger.Debug(
			"register: send update",
			zap.Object("update", update),
		)

		conn.AddUpdate(update)
	})
	defer unsubscribe()

	for {
		update, err := conn.RecvUpdate()
		if err != nil {
			return
		}

		logger.Debug(
			"register: received update",
			zap.Object("update", update),
		)

		if s.updateCountMetric != nil {
			s.updateCountMetric.With(prometheus.Labels{
				"type": string(registerUpdate.UpdateType),
			}).Inc()
		}

		if err := s.cluster.ApplyUpdate(update); err != nil {
			logger.Warn("apply update", zap.Error(err))
		}

		if s.nodeCountMetric != nil {
			s.nodeCountMetric.Set(float64(len(s.cluster.Nodes())))
		}
	}
}

func (s *Server) clusterRoute(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.With(zap.String("path", r.URL.Path))

	if err := json.NewEncoder(w).Encode(s.cluster.Nodes()); err != nil {
		logger.Error(
			"cluster request: failed to encode response",
			zap.Error(err),
			zap.Int("status", http.StatusInternalServerError),
		)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	logger.Debug("cluster request: ok", zap.Int("status", http.StatusOK))
}

func (s *Server) nodeRoute(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.With(zap.String("path", r.URL.Path))

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		logger.Debug(
			"node request: missing ID",
			zap.Int("status", http.StatusBadRequest),
		)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	node, ok := s.cluster.Node(id)
	if !ok {
		logger.Debug(
			"node request: node not found",
			zap.Int("status", http.StatusNotFound),
		)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(node); err != nil {
		logger.Error(
			"node request: failed to encode response",
			zap.Error(err),
			zap.Int("status", http.StatusInternalServerError),
		)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	logger.Debug("node request: ok", zap.Int("status", http.StatusOK))
}
