package server

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
)

func clientUpdateTypeToString(updateType rpc.ClientUpdateType) string {
	switch updateType {
	case rpc.ClientUpdateType_CLIENT_REGISTER:
		return "register"
	case rpc.ClientUpdateType_CLIENT_UNREGISTER:
		return "unregister"
	case rpc.ClientUpdateType_CLIENT_HEARTBEAT:
		return "heartbeat"
	default:
		return "unknown"
	}
}
