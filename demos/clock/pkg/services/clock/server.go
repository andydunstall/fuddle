package clock

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type server struct {
	logger *zap.Logger

	UnimplementedClockServer
}

func newServer(logger *zap.Logger) *server {
	return &server{
		logger: logger,
	}
}

func (s *server) Time(context.Context, *TimeRequest) (*TimeResponse, error) {
	ts := time.Now().UnixMilli()

	s.logger.Debug("time request", zap.Int64("timestamp", ts))

	return &TimeResponse{
		Time: ts,
	}, nil
}
