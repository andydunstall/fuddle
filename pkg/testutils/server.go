// Copyright (C) 2023 Andrew Dunstall
//
// Fuddle is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Fuddle is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package testutils

import (
	"net"

	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/server"
	"github.com/google/uuid"
)

type Server struct {
	id      string
	rpcAddr string

	server *server.Server
}

func StartServer() (*Server, error) {
	conf := testConfig()
	s := &Server{
		id:      conf.ID,
		rpcAddr: conf.AdvAddr,
		server:  server.NewServer(conf),
	}
	if err := s.server.Start(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Server) ID() string {
	return s.id
}

func (s *Server) RPCAddr() string {
	return s.rpcAddr
}

func (s *Server) GracefulStop() {
	s.server.GracefulStop()
}

func testConfig() *config.Config {
	conf := &config.Config{}

	conf.ID = "fuddle-" + uuid.New().String()[:8]

	conf.BindAddr = GetSystemAddress()
	conf.AdvAddr = conf.BindAddr

	conf.BindAdminAddr = GetSystemAddress()
	conf.AdvAdminAddr = conf.BindAdminAddr

	return conf
}

func GetSystemAddress() string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	return ln.Addr().String()
}
