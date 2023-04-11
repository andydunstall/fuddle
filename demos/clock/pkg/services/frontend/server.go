package frontend

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/fuddle-io/fuddle/demos/clock/pkg/services/clock"
	"go.uber.org/zap"
)

type server struct {
	httpServer  *http.Server
	clockClient *clock.Client
	logger      *zap.Logger
}

func newServer(ln *net.TCPListener, clockClient *clock.Client, logger *zap.Logger) *server {
	s := &server{
		clockClient: clockClient,
		logger:      logger,
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

	ts, err := s.clockClient.Time(ctx)
	if err != nil {
		s.logger.Error("error resolving time", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if _, err = w.Write([]byte(strconv.FormatInt(ts, 10))); err != nil {
		s.logger.Warn("error sending response", zap.Error(err))
	}

	s.logger.Debug("time request", zap.Int64("timestamp", ts))
}

func (s *server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	s.httpServer.Shutdown(ctx)
}
