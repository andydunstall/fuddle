package resolvers

import (
	"google.golang.org/grpc/resolver"
)

type StaticResolverBuilder struct {
	addrs []string
}

func NewStaticResolverBuilder(addrs []string) *StaticResolverBuilder {
	return &StaticResolverBuilder{
		addrs: addrs,
	}
}

func (s *StaticResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	var addrs []resolver.Address
	for _, addr := range s.addrs {
		addrs = append(addrs, resolver.Address{Addr: addr})
	}

	r := &StaticResolver{
		target: target,
		cc:     cc,
		addrs:  addrs,
	}
	r.start()
	return r, nil
}

func (s *StaticResolverBuilder) Scheme() string {
	return "static"
}

type StaticResolver struct {
	target resolver.Target
	cc     resolver.ClientConn
	addrs  []resolver.Address
}

func (s *StaticResolver) start() {
	s.updateAddresses(s.addrs)
}

func (s *StaticResolver) ResolveNow(resolver.ResolveNowOptions) {
	s.updateAddresses(s.addrs)
}

func (s *StaticResolver) Close() {
}

func (s *StaticResolver) updateAddresses(addrs []resolver.Address) {
	//nolint
	s.cc.UpdateState(resolver.State{Addresses: addrs})
}

var _ resolver.Builder = &StaticResolverBuilder{}
