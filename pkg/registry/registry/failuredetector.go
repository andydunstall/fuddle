package registry

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"go.uber.org/zap"
)

// UpdateLiveness updates the liveness of registered members, where:
//   - Up owned members who have not received a heartbeat for the heartbeat
//     timeout are marked as down
//   - Down owned members that have expired are marked as left
//   - Left owned members that have expired are removed
//   - Takes ownership of members whose owner is down for at least the heartbeat
//     their liveness is updated using the owners last contact as the members
//     last contact
func (r *Registry) UpdateLiveness(timestamp int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id := range r.members {
		r.updateMemberLivenessLocked(id, timestamp)
	}
}

func (r *Registry) updateMemberLivenessLocked(id string, timestamp int64) {
	// The local members liveness it never changed, it is always up.
	if id == r.localID {
		return
	}

	m := r.members[id]
	if m.Version.OwnerId == r.localID {
		r.updateOwnedMemberLivenessLocked(id, timestamp)
	} else {
		r.updateRemoteMemberLivenessLocked(id, timestamp)
	}
}

// updateOwnedMemberLivenessLocked updates the liveness of a member owned by
// this node.
func (r *Registry) updateOwnedMemberLivenessLocked(id string, timestamp int64) {
	member := r.members[id]
	lastSeen := r.lastSeen[id]

	switch member.Liveness {
	case rpc.Liveness_UP:
		// If we have not heard from the node in the heartbeat timeout, mark
		// it as down.
		if timestamp-lastSeen > r.heartbeatTimeout {
			r.logger.Info(
				"member missed heartbeat timeout; marking down",
				zap.String("member-id", id),
				zap.Int64("last-seen", lastSeen),
				zap.Int64("expiry", timestamp+r.reconnectTimeout),
			)

			r.updateMemberLocked(
				member.State,
				rpc.Liveness_DOWN,
				timestamp+r.reconnectTimeout,
			)
		}
	case rpc.Liveness_DOWN:
		// If the member has been down for the reconnect timeout, mark it as
		// left.
		if timestamp > member.Expiry {
			r.logger.Info(
				"down member expired; marking left",
				zap.String("member-id", id),
				zap.Int64("old-expiry", member.Expiry),
				zap.Int64("new-expiry", timestamp+r.tombstoneTimeout),
			)

			r.updateMemberLocked(
				member.State,
				rpc.Liveness_LEFT,
				timestamp+r.tombstoneTimeout,
			)
		}
	case rpc.Liveness_LEFT:
		// If the member has left and expired, it can be removed.
		if timestamp > member.Expiry {
			r.logger.Info(
				"left member expired; removing",
				zap.String("member-id", id),
				zap.Int64("old-expiry", member.Expiry),
			)

			r.deleteMemberLocked(id)
		}
	}
}

func (r *Registry) updateRemoteMemberLivenessLocked(id string, timestamp int64) {
	member := r.members[id]

	// If the member has left and expired it can be removed.
	if member.Liveness == rpc.Liveness_LEFT && timestamp > member.Expiry {
		r.logger.Info(
			"left member expired; removing",
			zap.String("member-id", id),
			zap.Int64("old-expiry", member.Expiry),
		)

		r.deleteMemberLocked(id)
		return
	}

	// If the owner of the node is still in the cluster, do nothing.
	ownerLastContact, ok := r.leftNodes[member.Version.OwnerId]
	if !ok {
		return
	}

	// If the owner has left the cluster for more than the heartbeat timeout,
	// try to take ownership of the member. Multiple nodes may end up completing
	// for ownership which is ok as one node will quickly win.
	if timestamp-ownerLastContact > r.heartbeatTimeout {
		// If the member is up, mark it as down as it hasn't reconnected to
		// another node in the heartbeat timeout. Otherwise leave its status
		// unchanged as its already down or left with an expiry.
		if member.Liveness == rpc.Liveness_UP {
			r.logger.Info(
				"members owner node down; taking ownership and marking down",
				zap.String("member-id", id),
				zap.String("owner-id", member.Version.OwnerId),
				zap.Int64("expiry", timestamp+r.reconnectTimeout),
			)

			r.updateMemberLocked(
				member.State,
				rpc.Liveness_DOWN,
				timestamp+r.reconnectTimeout,
			)
		} else {
			r.logger.Info(
				"members owner node down; taking ownership",
				zap.String("member-id", id),
				zap.String("owner-id", member.Version.OwnerId),
			)

			r.updateMemberLocked(
				member.State,
				member.Liveness,
				member.Expiry,
			)
		}
	}
}
