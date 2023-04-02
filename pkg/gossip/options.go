package gossip

import (
	"net"

	"go.uber.org/zap"
)

type options struct {
	onJoin      func(node Node)
	onLeave     func(node Node)
	tcpListener *net.TCPListener
	udpListener *net.UDPConn
	logger      *zap.Logger
}

func defaultOptions() options {
	return options{
		logger: zap.NewNop(),
	}
}

type Option interface {
	apply(*options)
}

type onJoinOption struct {
	cb func(node Node)
}

func (o onJoinOption) apply(opts *options) {
	opts.onJoin = o.cb
}

func WithOnJoin(cb func(node Node)) Option {
	return onJoinOption{cb: cb}
}

type onLeaveOption struct {
	cb func(node Node)
}

func (o onLeaveOption) apply(opts *options) {
	opts.onLeave = o.cb
}

func WithOnLeave(cb func(node Node)) Option {
	return onLeaveOption{cb: cb}
}

type tcpListenerOption struct {
	ln *net.TCPListener
}

func (o tcpListenerOption) apply(opts *options) {
	opts.tcpListener = o.ln
}

func WithTCPListener(ln *net.TCPListener) Option {
	return tcpListenerOption{ln: ln}
}

type udpListenerOption struct {
	ln *net.UDPConn
}

func (o udpListenerOption) apply(opts *options) {
	opts.udpListener = o.ln
}

func WithUDPListener(ln *net.UDPConn) Option {
	return udpListenerOption{ln: ln}
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
