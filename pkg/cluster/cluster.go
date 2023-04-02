package cluster

import (
	"go.uber.org/zap"
)

type Cluster struct {
	logger *zap.Logger
}

func NewCluster(logger *zap.Logger) *Cluster {
	return &Cluster{
		logger: logger,
	}
}

func (c *Cluster) OnJoin(id string, addr string) {
	c.logger.Info(
		"cluster on join",
		zap.String("id", id),
		zap.String("addr", addr),
	)
}

func (c *Cluster) OnLeave(id string) {
	c.logger.Info("cluster on leave", zap.String("id", id))
}
