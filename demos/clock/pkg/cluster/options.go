package cluster

type options struct {
	fuddleNodes   int
	clockNodes    int
	frontendNodes int
}

func defaultOptions() options {
	return options{
		fuddleNodes:   3,
		clockNodes:    3,
		frontendNodes: 3,
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

type clockNodesOption struct {
	nodes int
}

func (o clockNodesOption) apply(opts *options) {
	opts.clockNodes = o.nodes
}

func WithClockNodes(nodes int) Option {
	return clockNodesOption{nodes: nodes}
}

type frontendNodesOption struct {
	nodes int
}

func (o frontendNodesOption) apply(opts *options) {
	opts.frontendNodes = o.nodes
}

func WithFrontendNodes(nodes int) Option {
	return frontendNodesOption{nodes: nodes}
}
