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

package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/fuddle-io/fuddle/pkg/registry/cluster"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type server struct {
	cluster    *cluster.Cluster
	httpServer *http.Server
	listener   net.Listener

	logger *zap.Logger
}

func newServer(addr string, cluster *cluster.Cluster, options options) *server {
	server := &server{
		cluster:  cluster,
		listener: options.listener,
		logger:   options.logger,
	}

	r := mux.NewRouter()
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
	}
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./console/ui/build/static"))))
	r.Handle("/", http.FileServer(http.Dir("./console/ui/build")))

	httpServer := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	server.httpServer = httpServer

	return server
}

func (s *server) Start() error {
	ln := s.listener
	if ln == nil {
		// Setup the listener before starting to the goroutine to return any errors
		// binding or listening to the configured address.
		var err error
		ln, err = net.Listen("tcp", s.httpServer.Addr)
		if err != nil {
			return fmt.Errorf("admin server: %w", err)
		}
	}

	s.logger.Info(
		"starting admin server",
		zap.String("addr", s.httpServer.Addr),
	)

	go func() {
		if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			s.logger.Error("admin serve error", zap.Error(err))
		}
	}()

	return nil
}

func (s *server) GracefulStop() {
	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		s.logger.Error("failed to shut down admin server", zap.Error(err))
	}
}

func (s *server) clusterRoute(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(s.cluster.Nodes()); err != nil {
		s.logger.Error("failed to encode cluster response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *server) nodeRoute(w http.ResponseWriter, r *http.Request) {
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
