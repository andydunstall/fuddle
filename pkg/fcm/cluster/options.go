package cluster

type options struct {
	fuddleNodes int
}

func defaultOptions() options {
	return options{
		fuddleNodes: 3,
	}
}

type Option interface {
	apply(*options)
}

type fuddleNodesOption struct {
	nodes int
}

func (o fuddleNodesOption) apply(opts *options) {
	opts.fuddleNodes = o.nodes
}

func WithFuddleNodes(nodes int) Option {
	return fuddleNodesOption{nodes: nodes}
}
