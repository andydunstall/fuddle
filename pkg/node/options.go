package node

import (
	"net"

	"go.uber.org/zap/zapcore"
)

type options struct {
	gossipTCPListener *net.TCPListener
	gossipUDPListener *net.UDPConn
	rpcListener       *net.TCPListener
	adminListener     *net.TCPListener
	logLevel          zapcore.Level
	logPath           string
}

func defaultOptions() options {
	return options{
		logLevel: zapcore.InfoLevel,
		logPath:  "",
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

type adminListenerOption struct {
	ln *net.TCPListener
}

func (o adminListenerOption) apply(opts *options) {
	opts.adminListener = o.ln
}

func WithAdminListener(ln *net.TCPListener) Option {
	return &adminListenerOption{
		ln: ln,
	}
}

type logLevelOption struct {
	level zapcore.Level
}

func (o logLevelOption) apply(opts *options) {
	opts.logLevel = o.level
}

func WithLogLevel(level zapcore.Level) Option {
	return logLevelOption{level: level}
}

type logPathOption struct {
	path string
}

func (o logPathOption) apply(opts *options) {
	opts.logPath = o.path
}

// WithLogPath sets the logger output path. If unset defaults to stdout.
func WithLogPath(path string) Option {
	return logPathOption{path: path}
}
