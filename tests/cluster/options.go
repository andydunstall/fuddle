package cluster

import (
	"github.com/fuddle-io/fuddle/pkg/config"
)

type options struct {
	nodes          int
	registryConfig *config.Registry
}

func defaultOptions() options {
	return options{
		nodes:          3,
		registryConfig: config.DefaultRegistryConfig(),
	}
}

type Option interface {
	apply(*options)
}

type nodesOption struct {
	nodes int
}

func (o nodesOption) apply(opts *options) {
	opts.nodes = o.nodes
}

func WithNodes(nodes int) Option {
	return nodesOption{nodes: nodes}
}

type registryConfigOption struct {
	config *config.Registry
}

func (o registryConfigOption) apply(opts *options) {
	opts.registryConfig = o.config
}

func WithRegistryConfig(config *config.Registry) Option {
	return registryConfigOption{config: config}
}
