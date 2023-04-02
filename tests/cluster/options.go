package cluster

type options struct {
	nodes int
}

func defaultOptions() options {
	return options{
		nodes: 3,
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
