package fuddle

import (
	"net"

	"go.uber.org/zap"
)

type options struct {
	gossipTCPListener *net.TCPListener
	gossipUDPListener *net.UDPConn
	registryListener  *net.TCPListener
	logger            *zap.Logger
}

func defaultOptions() options {
	return options{
		logger: zap.NewNop(),
	}
}

type Option interface {
	apply(*options)
}

type gossipTCPListenerOption struct {
	ln *net.TCPListener
}

func (o gossipTCPListenerOption) apply(opts *options) {
	opts.gossipTCPListener = o.ln
}

func WithGossipTCPListener(ln *net.TCPListener) Option {
	return &gossipTCPListenerOption{
		ln: ln,
	}
}

type gossipUDPListenerOption struct {
	ln *net.UDPConn
}

func (o gossipUDPListenerOption) apply(opts *options) {
	opts.gossipUDPListener = o.ln
}

func WithGossipUDPListener(ln *net.UDPConn) Option {
	return &gossipUDPListenerOption{
		ln: ln,
	}
}

type registryListenerOption struct {
	ln *net.TCPListener
}

func (o registryListenerOption) apply(opts *options) {
	opts.registryListener = o.ln
}

func WithRegistryListener(ln *net.TCPListener) Option {
	return &registryListenerOption{
		ln: ln,
	}
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
