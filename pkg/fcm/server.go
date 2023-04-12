package fcm

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server

	logger *zap.Logger
}

func NewServer(addr string, port int, opts ...Option) (*Server, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	s := &Server{
		logger: options.logger,
	}

	r := mux.NewRouter()
	r.HandleFunc("/cluster", s.createCluster).Methods("GET")
	r.HandleFunc("/cluster/{id}", s.deleteCluster).Methods("DELETE")

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
}

func (s *Server) createCluster(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("create cluster")
}

func (s *Server) deleteCluster(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	s.logger.Info("delete cluster", zap.String("id", id))
}
