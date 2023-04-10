package frontend

import (
	"context"
	"net"
	"net/http"
	"time"
)

type server struct {
	httpServer *http.Server
}

func newServer(ln *net.TCPListener) *server {
	s := &server{}

	mux := http.NewServeMux()
	mux.HandleFunc("/time", s.timeRoute)
	httpServer := &http.Server{
		Handler:           mux,
		Addr:              ln.Addr().String(),
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	go func() {
		// nolint
		httpServer.Serve(ln)
	}()

	return &server{
		httpServer: httpServer,
	}
}

func (s *server) timeRoute(w http.ResponseWriter, r *http.Request) {
}

func (s *server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	s.httpServer.Shutdown(ctx)
}
