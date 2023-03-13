// Copyright (C) 2023 Andrew Dunstall
//
// Fuddle is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Fuddle is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package registry

import (
	"fmt"
	"io"

	"github.com/fuddle-io/fuddle/pkg/registry/cluster"
	"github.com/fuddle-io/fuddle/pkg/rpc"
	"go.uber.org/zap"
)

// Server exposes gRPC endpoints for the registry.
type Server struct {
	cluster *cluster.Cluster

	logger *zap.Logger

	rpc.UnimplementedRegistryServer
}

func NewServer(cluster *cluster.Cluster, logger *zap.Logger) *Server {
	return &Server{
		cluster: cluster,
		logger:  logger,
	}
}

func (s *Server) Register(stream rpc.Registry_RegisterServer) error {
	conn := newConnection(stream)
	defer conn.Close()

	// Wait for the connected node to join.
	joinUpdate, err := conn.RecvUpdate()
	if err != nil {
		return err
	}
	// If the first update is not the node joining this is a protocol error.
	if joinUpdate.UpdateType != rpc.NodeUpdateType_JOIN {
		return fmt.Errorf("protocol error: node must register")
	}
	if err := s.cluster.ApplyUpdate(joinUpdate); err != nil {
		return err
	}

	nodeID := joinUpdate.NodeId

	// Subscribe to the node map and send updates to the client. This will
	// replay all existing nodes as JOIN updates to ensure the subscriber
	// doesn't miss any updates.
	unsubscribe := s.cluster.Subscribe(true, func(update *rpc.NodeUpdate) {
		// Avoid echoing back updates from the connected nodes.
		if update.NodeId == nodeID {
			return
		}

		conn.AddUpdate(update)
	})
	defer unsubscribe()

	for {
		update, err := conn.RecvUpdate()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if err := s.cluster.ApplyUpdate(update); err != nil {
			s.logger.Error("update error", zap.Error(err))
		}
	}
}
