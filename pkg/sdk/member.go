package fuddle

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
)

type Member struct {
	ID       string
	Metadata map[string]string
}

func fromRPC(m *rpc.Member) Member {
	return Member{
		ID:       m.Id,
		Metadata: m.Metadata,
	}
}

func (m *Member) toRPC() *rpc.Member {
	return &rpc.Member{
		Id:       m.ID,
		Metadata: m.Metadata,
	}
}
