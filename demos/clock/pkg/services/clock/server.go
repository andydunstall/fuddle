package clock

import (
	"context"
	"time"
)

type server struct {
	UnimplementedClockServer
}

func (s *server) Time(context.Context, *TimeRequest) (*TimeResponse, error) {
	return &TimeResponse{
		Time: time.Now().UnixMilli(),
	}, nil
}
