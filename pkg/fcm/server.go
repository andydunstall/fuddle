package fcm

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/fuddle-io/fuddle/pkg/fcm/cluster"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type clusterRequest struct {
	Nodes   int `json:"nodes,omitempty"`
	Members int `json:"members,omitempty"`
}

type nodeResponse struct {
	ID        string `json:"id,omitempty"`
	RPCAddr   string `json:"rpc_addr,omitempty"`
	AdminAddr string `json:"admin_addr,omitempty"`
	LogPath   string `json:"log_path,omitempty"`
}

type memberResponse struct {
	ID      string `json:"id,omitempty"`
	LogPath string `json:"log_path,omitempty"`
}

type clusterResponse struct {
	ID      string           `json:"id,omitempty"`
	Nodes   []nodeResponse   `json:"nodes,omitempty"`
	Members []memberResponse `json:"members,omitempty"`
}

type promTarget struct {
	Targets []string          `json:"targets,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
}

type Server struct {
	clusters map[string]*cluster.Cluster

	// mu is a mutex protecting the fields above.
	mu sync.Mutex

	httpServer *http.Server

	logger *zap.Logger
}

func NewServer(addr string, port int, opts ...Option) (*Server, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	s := &Server{
		clusters: make(map[string]*cluster.Cluster),
		logger:   options.logger,
	}

	r := mux.NewRouter()
	r.HandleFunc("/cluster", s.createCluster).Methods("POST")
	r.HandleFunc("/cluster/{id}", s.deleteCluster).Methods("DELETE")
	r.HandleFunc("/cluster/{id}/prometheus", s.clusterPromTargets).Methods("GET")

	ln := options.listener
	if ln == nil {
		ip := net.ParseIP(addr)
		tcpAddr := &net.TCPAddr{IP: ip, Port: port}

		var err error
		ln, err = net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			s.logger.Info(
				"failed to start listener",
				zap.String("addr", ln.Addr().String()),
				zap.Error(err),
			)
			return nil, fmt.Errorf("admin server: start listener: %w", err)
		}
	}

	s.httpServer = &http.Server{
		Handler:           r,
		Addr:              ln.Addr().String(),
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}
	go func() {
		s.logger.Info("starting server", zap.String("addr", ln.Addr().String()))

		if err := s.httpServer.Serve(ln); err != nil {
			s.logger.Error("http serve error", zap.Error(err))
		}
	}()

	return s, nil
}

func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	s.httpServer.Shutdown(ctx)

	for _, c := range s.clusters {
		c.Shutdown()
	}
}

func (s *Server) createCluster(w http.ResponseWriter, r *http.Request) {
	var req clusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Warn("failed to decode cluster request", zap.Error(err))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	c, err := cluster.NewCluster(
		cluster.WithFuddleNodes(req.Nodes),
		cluster.WithMemberNodes(req.Members),
	)
	if err != nil {
		s.logger.Error("failed to create cluster", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	s.logger.Info("created cluster", zap.String("id", c.ID()))

	resp := clusterResponse{
		ID: c.ID(),
	}
	for _, node := range c.FuddleNodes() {
		resp.Nodes = append(resp.Nodes, nodeResponse{
			ID:        node.Fuddle.Config.NodeID,
			RPCAddr:   node.Fuddle.Config.RPC.JoinAdvAddr(),
			AdminAddr: node.Fuddle.Config.Admin.JoinAdvAddr(),
			LogPath:   c.LogPath(node.Fuddle.Config.NodeID),
		})
	}
	for _, node := range c.MemberNodes() {
		resp.Members = append(resp.Members, memberResponse{
			ID:      node.ID,
			LogPath: c.LogPath(node.ID),
		})
	}

	s.mu.Lock()
	s.clusters[c.ID()] = c
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to encode cluster response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) deleteCluster(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	s.mu.Lock()
	c, ok := s.clusters[id]
	delete(s.clusters, id)
	s.mu.Unlock()

	if !ok {
		s.logger.Warn("delete cluster; not found", zap.String("id", id))
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	c.Shutdown()

	s.logger.Info("delete cluster; ok", zap.String("id", id))
}

func (s *Server) clusterPromTargets(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	s.mu.Lock()
	targets := make(map[string]string)
	c, ok := s.clusters[id]
	if ok {
		for _, node := range c.FuddleNodes() {
			targets[node.Fuddle.Config.NodeID] = node.Fuddle.Config.Admin.JoinAdvAddr()
		}
	}

	s.mu.Unlock()

	if !ok {
		s.logger.Warn("cluster prom targets; not found", zap.String("id", id))
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var resp []promTarget
	for nodeID, addr := range targets {
		resp = append(resp, promTarget{
			Targets: []string{addr},
			Labels: map[string]string{
				"instance": nodeID,
			},
		})
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to encode prom targets response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
