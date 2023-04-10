package frontend

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/fuddle-io/fuddle/demos/clock/pkg/services/clock"
)

type server struct {
	httpServer  *http.Server
	clockClient *clock.Client
}

func newServer(ln *net.TCPListener, clockClient *clock.Client) *server {
	s := &server{
		clockClient: clockClient,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/time", s.timeRoute)
	s.httpServer = &http.Server{
		Handler:           mux,
		Addr:              ln.Addr().String(),
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	go func() {
		// nolint
		s.httpServer.Serve(ln)
	}()

	return s
}

func (s *server) timeRoute(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	t, err := s.clockClient.Time(ctx)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if _, err = w.Write([]byte(strconv.FormatInt(t, 10))); err != nil {
		// TODO log
		fmt.Println(err)
	}
}

func (s *server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	s.httpServer.Shutdown(ctx)
}
