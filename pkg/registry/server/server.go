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

package server

import (
	"context"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registry/cluster"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Server exposes endpoints for nodes to register themselves and inspect the
// registry.
type Server struct {
	cluster *cluster.Cluster

	promRegistry          *prometheus.Registry
	nodeCountMetric       prometheus.Gauge
	updateCountMetric     *prometheus.CounterVec
	connectionCountMetric prometheus.Gauge

	logger *zap.Logger

	rpc.UnimplementedRegistryServer
}

func NewServer(addr string, cluster *cluster.Cluster, opts ...Option) *Server {
	options := options{
		logger:       zap.NewNop(),
		listener:     nil,
		promRegistry: nil,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	server := &Server{
		cluster:      cluster,
		promRegistry: options.promRegistry,
		logger:       options.logger,
	}

	if options.promRegistry != nil {
		nodeCountMetric := prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "fuddle_registry_node_count",
				Help: "Number of nodes in the registry.",
			},
		)
		options.promRegistry.MustRegister(nodeCountMetric)
		nodeCountMetric.Set(float64(len(cluster.Nodes())))
		server.nodeCountMetric = nodeCountMetric

		updateCountMetric := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fuddle_registry_update_count",
				Help: "Number of updates to the registry.",
			},
			[]string{"type"},
		)
		options.promRegistry.MustRegister(updateCountMetric)
		server.updateCountMetric = updateCountMetric

		connectionCountMetric := prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "fuddle_registry_connection_count",
				Help: "Number of connections to the registry.",
			},
		)
		options.promRegistry.MustRegister(connectionCountMetric)
		server.connectionCountMetric = connectionCountMetric
	}

	return server
}

func (s *Server) Register(stream rpc.Registry_RegisterServer) error {
	conn := newNonBlockingConn(stream)

	if s.connectionCountMetric != nil {
		s.connectionCountMetric.Inc()
		defer s.connectionCountMetric.Dec()
	}

	// Wait for the connected node to register.
	m, err := conn.Recv()
	if err != nil {
		s.logger.Debug("register: not received register update")
		return nil
	}
	// If the first update is not the node joining this is a protocol error.
	if m.MessageType != rpc.MessageType_NODE_UPDATE || m.NodeUpdate == nil {
		s.logger.Warn("register: protocol error: message not a node update")
		return nil
	}
	registerUpdate := m.NodeUpdate
	if registerUpdate.UpdateType != rpc.NodeUpdateType_REGISTER {
		s.logger.Warn("register: protocol error: update not a register")
		return nil
	}

	if s.updateCountMetric != nil {
		s.updateCountMetric.With(prometheus.Labels{
			"type": string(registerUpdate.UpdateType),
		}).Inc()
	}

	metadata := cluster.Metadata{}
	for k, v := range registerUpdate.Metadata {
		metadata[k] = v.Value
	}
	registerUpdateRPC := &cluster.NodeUpdate{
		ID:         registerUpdate.NodeId,
		UpdateType: cluster.UpdateTypeRegister,
		Attributes: &cluster.NodeAttributes{
			Service:  registerUpdate.Attributes.Service,
			Locality: registerUpdate.Attributes.Locality,
			Created:  registerUpdate.Attributes.Created,
			Revision: registerUpdate.Attributes.Revision,
		},
		Metadata: metadata,
	}

	s.logger.Debug(
		"register: received update",
		zap.Object("update", registerUpdateRPC),
	)

	if err := s.cluster.ApplyUpdate(registerUpdateRPC); err != nil {
		s.logger.Warn("register: failed to apply register update", zap.Error(err))
		return nil
	}

	if s.nodeCountMetric != nil {
		s.nodeCountMetric.Set(float64(len(s.cluster.Nodes())))
	}

	nodeID := registerUpdateRPC.ID

	logger := s.logger.With(zap.String("client-node", nodeID))

	// Subscribe to the node map and send updates to the client. This will
	// replay all existing nodes as JOIN updates to ensure the subscriber
	// doesn't miss any updates.
	unsubscribe := s.cluster.Subscribe(true, func(update *cluster.NodeUpdate) {
		// Avoid echoing back updates from the connected nodes.
		if update.ID == nodeID {
			return
		}

		logger.Debug(
			"register: send update",
			zap.Object("update", update),
		)

		var updateType rpc.NodeUpdateType
		switch update.UpdateType {
		case cluster.UpdateTypeRegister:
			updateType = rpc.NodeUpdateType_REGISTER
		case cluster.UpdateTypeUnregister:
			updateType = rpc.NodeUpdateType_UNREGISTER
		case cluster.UpdateTypeMetadata:
			updateType = rpc.NodeUpdateType_METADATA
		}

		metadata := make(map[string]*rpc.VersionedValue)
		for k, v := range update.Metadata {
			metadata[k] = &rpc.VersionedValue{Value: v}
		}
		m := &rpc.Message{
			MessageType: rpc.MessageType_NODE_UPDATE,
			NodeUpdate: &rpc.NodeUpdate{
				NodeId:     update.ID,
				UpdateType: updateType,
				Attributes: &rpc.NodeAttributes{
					Service:  update.Attributes.Service,
					Locality: update.Attributes.Locality,
					Created:  update.Attributes.Created,
					Revision: update.Attributes.Revision,
				},
				Metadata: metadata,
			},
		}
		if err := conn.Send(m); err != nil {
			return
		}
	})
	defer unsubscribe()

	for {
		m, err := conn.Recv()
		if err != nil {
			return nil
		}

		switch m.MessageType {
		case rpc.MessageType_NODE_UPDATE:
			update := m.NodeUpdate
			if update == nil {
				logger.Warn("register: missing node update")
				return nil
			}

			metadataRPC := cluster.Metadata{}
			for k, v := range update.Metadata {
				metadataRPC[k] = v.Value
			}
			var updateType cluster.UpdateType
			switch update.UpdateType {
			case rpc.NodeUpdateType_REGISTER:
				updateType = cluster.UpdateTypeRegister
			case rpc.NodeUpdateType_UNREGISTER:
				updateType = cluster.UpdateTypeUnregister
			case rpc.NodeUpdateType_METADATA:
				updateType = cluster.UpdateTypeMetadata
			}

			updateRPC := &cluster.NodeUpdate{
				ID:         registerUpdate.NodeId,
				UpdateType: updateType,
				Attributes: &cluster.NodeAttributes{
					Service:  registerUpdate.Attributes.Service,
					Locality: registerUpdate.Attributes.Locality,
					Created:  registerUpdate.Attributes.Created,
					Revision: registerUpdate.Attributes.Revision,
				},
				Metadata: metadataRPC,
			}

			if s.updateCountMetric != nil {
				s.updateCountMetric.With(prometheus.Labels{
					"type": string(updateRPC.UpdateType),
				}).Inc()
			}

			if err := s.cluster.ApplyUpdate(updateRPC); err != nil {
				logger.Warn("apply update", zap.Error(err))
			}

			if s.nodeCountMetric != nil {
				s.nodeCountMetric.Set(float64(len(s.cluster.Nodes())))
			}
		default:
			logger.Warn(
				"register: unknown message type",
				zap.String("message-type", string(m.MessageType)),
			)
			return nil
		}
	}
}

func (s *Server) Nodes(ctx context.Context, req *rpc.NodesRequest) (*rpc.NodesResponse, error) {
	nodes := s.cluster.Nodes()
	var rpcNodes []*rpc.Node

	for _, node := range nodes {
		metadata := make(map[string]*rpc.VersionedValue)
		if req.IncludeMetadata {
			for k, v := range node.Metadata {
				metadata[k] = &rpc.VersionedValue{Value: v}
			}
		}
		rpcNode := &rpc.Node{
			Id: node.ID,
			Attributes: &rpc.NodeAttributes{
				Service:  node.Service,
				Locality: node.Locality,
				Created:  node.Created,
				Revision: node.Revision,
			},
			Metadata: metadata,
		}
		rpcNodes = append(rpcNodes, rpcNode)
	}
	return &rpc.NodesResponse{
		Nodes: rpcNodes,
	}, nil
}

func (s *Server) Node(ctx context.Context, req *rpc.NodeRequest) (*rpc.NodeResponse, error) {
	node, ok := s.cluster.Node(req.NodeId)
	if !ok {
		return &rpc.NodeResponse{
			Error: &rpc.Error{
				Status:      rpc.ErrorStatus_NOT_FOUND,
				Description: "node not found",
			},
		}, nil
	}

	metadata := make(map[string]*rpc.VersionedValue)
	for k, v := range node.Metadata {
		metadata[k] = &rpc.VersionedValue{Value: v}
	}
	rpcNode := &rpc.Node{
		Id: node.ID,
		Attributes: &rpc.NodeAttributes{
			Service:  node.Service,
			Locality: node.Locality,
			Created:  node.Created,
			Revision: node.Revision,
		},
		Metadata: metadata,
	}
	return &rpc.NodeResponse{
		Node: rpcNode,
	}, nil
}
