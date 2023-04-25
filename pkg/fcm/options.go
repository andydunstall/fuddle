package fcm

import (
	"net"

	"go.uber.org/zap"
)

type options struct {
	defaultCluster bool
	listener       net.Listener
	logger         *zap.Logger
	clusterLogDir  string
}

func defaultOptions() *options {
	return &options{
		defaultCluster: false,
		listener:       nil,
		logger:         zap.NewNop(),
		clusterLogDir:  "",
	}
}

type Option interface {
	apply(*options)
}

type defaultClusterOption bool

func (o defaultClusterOption) apply(opts *options) {
	opts.defaultCluster = bool(o)
}

func WithDefaultCluster(cluster bool) Option {
	return defaultClusterOption(cluster)
}

type listenerOption struct {
	listener net.Listener
}

func (o listenerOption) apply(opts *options) {
	opts.listener = o.listener
}

func WithListener(ln net.Listener) Option {
	return listenerOption{listener: ln}
}

type loggerOption struct {
	Log *zap.Logger
}

func (o loggerOption) apply(opts *options) {
	opts.logger = o.Log
}

func WithLogger(log *zap.Logger) Option {
	return loggerOption{Log: log}
}

type clusterLogDirOption struct {
	dir string
}

func (o clusterLogDirOption) apply(opts *options) {
	opts.clusterLogDir = o.dir
}

func WithClusterLogDir(dir string) Option {
	return clusterLogDirOption{dir: dir}
}
