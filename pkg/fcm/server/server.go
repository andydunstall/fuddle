package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/fuddle-io/fuddle/pkg/fcm/cluster"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type clusterRequest struct {
	Nodes   int `json:"nodes,omitempty"`
	Members int `json:"members,omitempty"`
}

type nodesRequest struct {
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

type clusterHealth struct {
	Healthy bool `json:"healthy,omitempty"`
}

type nodesResponse struct {
	Nodes   []nodeResponse   `json:"nodes,omitempty"`
	Members []memberResponse `json:"members,omitempty"`
}

type promTarget struct {
	Targets []string          `json:"targets,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
}

type Server struct {
	clusters *cluster.Manager

	httpServer *http.Server

	logger *zap.Logger
}

func NewServer(addr string, port int, clusters *cluster.Manager, opts ...Option) (*Server, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	s := &Server{
		clusters: clusters,
		logger:   options.logger,
	}

	r := mux.NewRouter()
	r.HandleFunc("/cluster", s.createCluster).Methods("POST")
	r.HandleFunc("/cluster/{id}", s.describeCluster).Methods("GET")
	r.HandleFunc("/cluster/{id}/health", s.clusterHealth).Methods("GET")
	r.HandleFunc("/cluster/{id}", s.deleteCluster).Methods("DELETE")

	r.HandleFunc("/cluster/{id}/nodes/add", s.addNodes).Methods("POST")
	r.HandleFunc("/cluster/{id}/nodes/remove", s.removeNodes).Methods("POST")

	r.HandleFunc("/cluster/{id}/prometheus/targets", s.clusterPromTargets).Methods("GET")

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

	s.clusters.Shutdown()
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

	s.clusters.Add(c)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to encode cluster response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) describeCluster(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	c, ok := s.clusters.Get(id)
	if !ok {
		s.logger.Warn("describe cluster; not found", zap.String("id", id))
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

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

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to encode cluster response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) clusterHealth(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	c, ok := s.clusters.Get(id)
	if !ok {
		s.logger.Warn("cluster health; not found", zap.String("id", id))
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	resp := clusterHealth{
		Healthy: c.Healthy(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to encode cluster response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) deleteCluster(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if !s.clusters.Delete(id) {
		s.logger.Warn("delete cluster; not found", zap.String("id", id))
		return
	}

	s.logger.Info("delete cluster; ok", zap.String("id", id))
}

func (s *Server) addNodes(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	c, ok := s.clusters.Get(id)
	if !ok {
		s.logger.Warn("add nodes; cluster not found", zap.String("id", id))
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var req nodesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Warn("failed to decode nodes request", zap.Error(err))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	resp := nodesResponse{}
	for i := 0; i != req.Nodes; i++ {
		n, err := c.AddFuddleNode()
		if err != nil {
			s.logger.Error("failed to add fuddle node", zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		resp.Nodes = append(resp.Nodes, nodeResponse{
			ID:        n.Fuddle.Config.NodeID,
			RPCAddr:   n.Fuddle.Config.RPC.JoinAdvAddr(),
			AdminAddr: n.Fuddle.Config.Admin.JoinAdvAddr(),
			LogPath:   c.LogPath(n.Fuddle.Config.NodeID),
		})
	}
	for i := 0; i != req.Members; i++ {
		n, err := c.AddMemberNode()
		if err != nil {
			s.logger.Error("failed to add client node", zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		resp.Members = append(resp.Members, memberResponse{
			ID:      n.ID,
			LogPath: c.LogPath(n.ID),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to encode nodes response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) removeNodes(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	c, ok := s.clusters.Get(id)
	if !ok {
		s.logger.Warn("remove nodes; cluster not found", zap.String("id", id))
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var req nodesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Warn("failed to decode nodes request", zap.Error(err))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	resp := nodesResponse{}
	for i := 0; i != req.Nodes; i++ {
		id := c.RemoveFuddleNode()
		if id != "" {
			resp.Nodes = append(resp.Nodes, nodeResponse{
				ID: id,
			})
		}
	}
	for i := 0; i != req.Members; i++ {
		id := c.RemoveMemberNode()
		if id != "" {
			resp.Members = append(resp.Members, memberResponse{
				ID: id,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to encode members response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) clusterPromTargets(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	targets := make(map[string]string)
	c, ok := s.clusters.Get(id)
	if ok {
		for _, node := range c.FuddleNodes() {
			targets[node.Fuddle.Config.NodeID] = node.Fuddle.Config.Admin.JoinAdvAddr()
		}
	}

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
