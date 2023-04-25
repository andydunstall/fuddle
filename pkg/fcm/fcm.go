package fcm

import (
	"fmt"

	"github.com/fuddle-io/fuddle/pkg/fcm/cluster"
	"github.com/fuddle-io/fuddle/pkg/fcm/server"
	"go.uber.org/zap"
)

type FCM struct {
	server *server.Server
}

func NewFCM(addr string, port int, opts ...Option) (*FCM, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	logger := options.logger
	logger.Info("starting fcm")

	clusters := cluster.NewManager()
	if options.defaultCluster {
		c, err := cluster.NewCluster(
			cluster.WithFuddleNodes(3),
			cluster.WithMemberNodes(20),
			cluster.WithDefaultCluster(),
			cluster.WithLogDir(options.clusterLogDir),
		)
		if err != nil {
			return nil, fmt.Errorf("fcm: %w", err)
		}
		logger.Info("created default cluster", zap.String("id", c.ID()))
		clusters.Add(c)
	}

	server, err := server.NewServer(
		addr,
		port,
		clusters,
		server.WithLogger(options.logger),
		server.WithListener(options.listener),
	)
	if err != nil {
		return nil, fmt.Errorf("fcm: %w", err)
	}

	return &FCM{
		server: server,
	}, nil
}

func (fcm *FCM) Shutdown() {
	fcm.server.Shutdown()
}
