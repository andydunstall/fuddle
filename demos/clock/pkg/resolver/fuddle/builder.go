package fuddle

import (
	fuddle "github.com/fuddle-io/fuddle-go"
	"google.golang.org/grpc/resolver"
)

type Builder struct {
	registry *fuddle.Fuddle
	service  string
}

func NewBuilder(registry *fuddle.Fuddle, service string) *Builder {
	return &Builder{
		registry: registry,
		service:  service,
	}
}

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &Resolver{
		registry: b.registry,
		service:  b.service,
		cc:       cc,
	}
	r.start()
	return r, nil
}

func (b *Builder) Scheme() string {
	return "fuddle"
}
