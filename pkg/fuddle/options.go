package fuddle

import (
	"net"

	"go.uber.org/zap"
)

type options struct {
	gossipTCPListener *net.TCPListener
	gossipUDPListener *net.UDPConn
	rpcListener       *net.TCPListener
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

type rpcListenerOption struct {
	ln *net.TCPListener
}

func (o rpcListenerOption) apply(opts *options) {
	opts.rpcListener = o.ln
}

func WithRPCListener(ln *net.TCPListener) Option {
	return &rpcListenerOption{
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
