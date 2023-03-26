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

package registry

import (
	"context"
	"testing"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func TestServer_RegisterThenLookupMember(t *testing.T) {
	s := NewServer(NewRegistry(testutils.RandomMember()))

	registeredMember := testutils.RandomMember()
	registerResp, err := s.RegisterMember(context.Background(), &rpc.RegisterMemberRequest{
		Member: registeredMember,
	})
	assert.NoError(t, err)
	assert.Nil(t, registerResp.Error)

	expectedMember := registeredMember
	expectedMember.Version = 1

	memberResp, err := s.Member(context.Background(), &rpc.MemberRequest{
		Id: registeredMember.Id,
	})
	assert.NoError(t, err)
	assert.Nil(t, memberResp.Error)
	assert.Equal(t, expectedMember, memberResp.Member)
}

func TestServer_RegisterAlreadyRegistered(t *testing.T) {
	s := NewServer(NewRegistry(testutils.RandomMember()))

	registeredMember := testutils.RandomMember()

	resp, err := s.RegisterMember(context.Background(), &rpc.RegisterMemberRequest{
		Member: registeredMember,
	})
	assert.NoError(t, err)
	assert.Nil(t, resp.Error)

	resp, err = s.RegisterMember(context.Background(), &rpc.RegisterMemberRequest{
		Member: registeredMember,
	})
	assert.NoError(t, err)
	assert.Equal(t, rpc.ErrorStatusV2_ALREADY_REGISTERED, resp.Error.Status)
}

func TestServer_RegisterInvalidMember(t *testing.T) {
	s := NewServer(NewRegistry(testutils.RandomMember()))

	registeredMember := testutils.RandomMember()
	registeredMember.Id = ""

	resp, err := s.RegisterMember(context.Background(), &rpc.RegisterMemberRequest{
		Member: registeredMember,
	})
	assert.NoError(t, err)
	assert.Equal(t, rpc.ErrorStatusV2_INVALID_MEMBER, resp.Error.Status)
}

func TestServer_Unregister(t *testing.T) {
	s := NewServer(NewRegistry(testutils.RandomMember()))

	registeredMember := testutils.RandomMember()
	registerResp, err := s.RegisterMember(context.Background(), &rpc.RegisterMemberRequest{
		Member: registeredMember,
	})
	assert.NoError(t, err)
	assert.Nil(t, registerResp.Error)

	_, err = s.UnregisterMember(context.Background(), &rpc.UnregisterMemberRequest{
		Id: registeredMember.Id,
	})
	assert.NoError(t, err)

	resp, err := s.Member(context.Background(), &rpc.MemberRequest{
		Id: registeredMember.Id,
	})
	assert.NoError(t, err)
	assert.Equal(t, rpc.ErrorStatusV2_NOT_FOUND, resp.Error.Status)
}

func TestServer_UpdateMemberMetadataThenLookup(t *testing.T) {
	s := NewServer(NewRegistry(testutils.RandomMember()))

	registeredMember := testutils.RandomMember()
	registerResp, err := s.RegisterMember(context.Background(), &rpc.RegisterMemberRequest{
		Member: registeredMember,
	})
	assert.NoError(t, err)
	assert.Nil(t, registerResp.Error)

	update := testutils.RandomMetadata()

	updateResp, err := s.UpdateMemberMetadata(context.Background(), &rpc.UpdateMemberMetadataRequest{
		Id:       registeredMember.Id,
		Metadata: update,
	})
	assert.NoError(t, err)
	assert.Nil(t, updateResp.Error)

	expectedMember := registeredMember
	for k, v := range update {
		expectedMember.Metadata[k] = v
	}
	expectedMember.Version = 2

	memberResp, err := s.Member(context.Background(), &rpc.MemberRequest{
		Id: registeredMember.Id,
	})
	assert.NoError(t, err)
	assert.Nil(t, memberResp.Error)
	assert.Equal(t, expectedMember, memberResp.Member)
}

func TestServer_UpdateMemberMetadataNotRegistered(t *testing.T) {
	s := NewServer(NewRegistry(testutils.RandomMember()))

	registeredMember := testutils.RandomMember()
	registerResp, err := s.RegisterMember(context.Background(), &rpc.RegisterMemberRequest{
		Member: registeredMember,
	})
	assert.NoError(t, err)
	assert.Nil(t, registerResp.Error)

	updateResp, err := s.UpdateMemberMetadata(context.Background(), &rpc.UpdateMemberMetadataRequest{
		Id:       registeredMember.Id,
		Metadata: nil,
	})
	assert.NoError(t, err)
	assert.Equal(t, rpc.ErrorStatusV2_INVALID_MEMBER, updateResp.Error.Status)
}

func TestServer_UpdateMemberMetadataInvalidMember(t *testing.T) {
	s := NewServer(NewRegistry(testutils.RandomMember()))

	resp, err := s.UpdateMemberMetadata(context.Background(), &rpc.UpdateMemberMetadataRequest{
		Id:       "not-found",
		Metadata: testutils.RandomMetadata(),
	})
	assert.NoError(t, err)
	assert.Equal(t, rpc.ErrorStatusV2_NOT_REGISTERED, resp.Error.Status)
}

func TestServer_MemberNotFound(t *testing.T) {
	s := NewServer(NewRegistry(testutils.RandomMember()))

	resp, err := s.Member(context.Background(), &rpc.MemberRequest{
		Id: "not-found",
	})
	assert.NoError(t, err)
	assert.Equal(t, rpc.ErrorStatusV2_NOT_FOUND, resp.Error.Status)
}
