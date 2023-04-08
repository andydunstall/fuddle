package fuddle

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
)

type Member struct {
	ID       string
	Service  string
	Locality string
	Created  int64
	Revision string
	Metadata map[string]string
}

func fromRPC(m *rpc.Member) Member {
	return Member{
		ID:       m.Id,
		Service:  m.Service,
		Locality: m.Locality,
		Created:  m.Created,
		Revision: m.Revision,
		Metadata: m.Metadata,
	}
}
