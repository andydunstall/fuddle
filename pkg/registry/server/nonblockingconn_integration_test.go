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

//go:build integration

package server

import (
	"testing"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/stretchr/testify/assert"
)

func TestPendingMessage_PushMessages(t *testing.T) {
	pending := newPendingMessages()
	pending.Push(&rpc.Message{
		MessageType: rpc.MessageType_HEARTBEAT,
	})
	pending.Push(&rpc.Message{
		MessageType: rpc.MessageType_NODE_UPDATE,
	})
	msgs, ok := pending.Wait()
	assert.True(t, ok)
	assert.Equal(t, []*rpc.Message{
		{
			MessageType: rpc.MessageType_HEARTBEAT,
		},
		{
			MessageType: rpc.MessageType_NODE_UPDATE,
		},
	}, msgs)
}

func TestPendingMessage_WaitWhenClosed(t *testing.T) {
	pending := newPendingMessages()
	pending.Close()
	_, ok := pending.Wait()
	assert.False(t, ok)
}

func TestPendingMessage_WaitThenPush(t *testing.T) {
	pending := newPendingMessages()
	go func() {
		<-time.After(time.Millisecond * 10)
		pending.Push(&rpc.Message{
			MessageType: rpc.MessageType_HEARTBEAT,
		})
	}()

	msgs, ok := pending.Wait()
	assert.True(t, ok)
	assert.Equal(t, []*rpc.Message{
		{
			MessageType: rpc.MessageType_HEARTBEAT,
		},
	}, msgs)
}

func TestPendingMessage_WaitThenClose(t *testing.T) {
	pending := newPendingMessages()
	go func() {
		<-time.After(time.Millisecond * 10)
		pending.Close()
	}()
	_, ok := pending.Wait()
	assert.False(t, ok)
}
