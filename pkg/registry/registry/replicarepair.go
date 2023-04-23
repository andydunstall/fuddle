package registry

import (
	"math/rand"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
)

// Digest returns the timestamps of the known nodes in the registry upto the
// limit.
//
// If the limit is exceeded, members are added in random order until the limit
// is hit.
func (r *Registry) Digest(limit int) map[string]*rpc.MonotonicTimestamp {
	r.mu.Lock()
	defer r.mu.Unlock()

	digest := make(map[string]*rpc.MonotonicTimestamp)

	for id := range r.priorityMembers {
		// The member may not exist, so if not use a timestamp of 0.
		if m, ok := r.members[id]; ok {
			digest[id] = &rpc.MonotonicTimestamp{
				Timestamp: m.Version.Timestamp.Timestamp,
				Counter:   m.Version.Timestamp.Counter,
			}
		} else {
			digest[id] = &rpc.MonotonicTimestamp{
				Timestamp: 0,
			}
		}
		delete(r.priorityMembers, id)

		if len(digest) >= limit {
			return digest
		}
	}

	var memberIDs []string
	for id := range r.members {
		memberIDs = append(memberIDs, id)
	}
	memberIDs = shuffleStrings(memberIDs)

	for _, id := range memberIDs {
		digest[id] = &rpc.MonotonicTimestamp{
			Timestamp: r.members[id].Version.Timestamp.Timestamp,
			Counter:   r.members[id].Version.Timestamp.Counter,
		}
		if len(digest) >= limit {
			return digest
		}
	}

	return digest
}

// Delta returns any members that the registry knows that are more up to date
// than those in the digest.
//
// It also tracks which known members are out of date, or unknown, to request
// in the next digest.
func (r *Registry) Delta(digest map[string]*rpc.MonotonicTimestamp) []*rpc.Member2 {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Compare our local state with the requested version in the digest.
	// If our version is greater, then we include the full member in the
	// response. Otherwise if either we don't know about the member, or the
	// sender is more up to date, then we save the member ID as a 'priority
	// member' to be included in the next digest as a priority.
	//
	// Note only compare timestamps, not owner IDs. If there is a conflict where
	// multiple nodes took ownership of a member in the same millisecond, it
	// will be resolved later so it doesn't matter which we pick.

	var members []*rpc.Member2
	for id, timestamp := range digest {
		if m, ok := r.members[id]; ok {
			if compareTimestamps(timestamp, m.Version.Timestamp) > 0 {
				members = append(members, copyMember(m))
			} else {
				r.priorityMembers[id] = struct{}{}
			}
		} else {
			r.priorityMembers[id] = struct{}{}
		}
	}

	return members
}

func shuffleStrings(src []string) []string {
	res := make([]string, len(src))
	perm := rand.Perm(len(src))
	for i, v := range perm {
		res[v] = src[i]
	}
	return res
}

func compareTimestamps(lhs *rpc.MonotonicTimestamp, rhs *rpc.MonotonicTimestamp) int {
	if lhs.Timestamp < rhs.Timestamp {
		return 1
	}
	if lhs.Timestamp > rhs.Timestamp {
		return -1
	}
	if lhs.Counter < rhs.Counter {
		return 1
	}
	if lhs.Counter > rhs.Counter {
		return -1
	}
	return 0
}
