package fuddle

import (
	"fmt"

	fuddle "github.com/fuddle-io/fuddle-go"
	"google.golang.org/grpc/resolver"
)

type Resolver struct {
	registry *fuddle.Fuddle
	service  string

	cc resolver.ClientConn

	unsubscribe func()
}

func (r *Resolver) ResolveNow(resolver.ResolveNowOptions) {
	r.resolve()
}

func (r *Resolver) Close() {
	if r.unsubscribe != nil {
		r.unsubscribe()
	}
}

func (r *Resolver) start() {
	r.unsubscribe = r.registry.Subscribe(func() {
		r.resolve()
	})
}

func (r *Resolver) resolve() {
	var addrs []resolver.Address
	for _, m := range r.registry.Members() {
		if m.Service == r.service && m.Metadata != nil {
			addr, ok := m.Metadata["rpc-addr"]
			if ok {
				addrs = append(addrs, resolver.Address{
					Addr: addr,
				})
			}
		}
	}
	if err := r.cc.UpdateState(resolver.State{Addresses: addrs}); err != nil {
		fmt.Println(err)
	}
}
