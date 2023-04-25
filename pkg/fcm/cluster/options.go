package cluster

type options struct {
	fuddleNodes    int
	memberNodes    int
	defaultCluster bool
}

func defaultOptions() options {
	return options{
		fuddleNodes:    3,
		memberNodes:    0,
		defaultCluster: false,
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

type memberNodesOption struct {
	nodes int
}

func (o memberNodesOption) apply(opts *options) {
	opts.memberNodes = o.nodes
}

func WithMemberNodes(nodes int) Option {
	return memberNodesOption{nodes: nodes}
}

type defaultClusterOption bool

func (o defaultClusterOption) apply(opts *options) {
	opts.defaultCluster = bool(o)
}

func WithDefaultCluster() Option {
	return defaultClusterOption(true)
}
