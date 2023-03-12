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

package counter

import (
	"sync"
)

type counterSubHandler struct {
	Callback func(id string, count uint64)
}

// counter aggregates the registered counts from each contributor.
type counter struct {
	// contributions contains the set of counts from each contributor.
	contributions map[interface{}]map[string]uint64

	// aggregates contains the aggregate counts of each contributor.
	aggregates map[string]uint64

	subscribers map[*counterSubHandler]interface{}

	// mu is a mutex protecting the fields above.
	mu sync.Mutex
}

func newCounter() *counter {
	return &counter{
		contributions: make(map[interface{}]map[string]uint64),
		aggregates:    make(map[string]uint64),
		subscribers:   make(map[*counterSubHandler]interface{}),
	}
}

func (c *counter) Aggregates() map[string]uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	aggregates := make(map[string]uint64)
	for k, v := range c.aggregates {
		aggregates[k] = v
	}
	return aggregates
}

// Subscribe subscribers to updates when any contributiosn are changed.
//
// Returns a function to unsubsribe.
func (c *counter) Subscribe(cb func(id string, count uint64)) func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	handle := &counterSubHandler{
		Callback: cb,
	}
	c.subscribers[handle] = struct{}{}

	return func() {
		c.unsubscribe(handle)
	}
}

// Register updates the counts for the contributor.
func (c *counter) Register(contributor interface{}, counts map[string]uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.contributions[contributor] = counts
	c.aggregateLocked()
}

// Unregister removes the counts for the contributor.
func (c *counter) Unregister(contributor interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.contributions, contributor)
	c.aggregateLocked()
}

func (c *counter) unsubscribe(handle *counterSubHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.subscribers, handle)
}

func (c *counter) aggregateLocked() {
	updatedAggregates := make(map[string]uint64)
	for _, counts := range c.contributions {
		for id, count := range counts {
			updatedAggregates[id] += count
		}
	}

	updates := make(map[string]uint64)
	// Compare the new aggregate to the existing (which default to zero if not
	// found).
	for id, count := range updatedAggregates {
		if count != c.aggregates[id] {
			updates[id] = count
		}
	}
	// Check for any IDs in the old aggregates but not in the updated.
	for id := range c.aggregates {
		if _, ok := updatedAggregates[id]; !ok {
			updates[id] = 0
		}
	}

	c.aggregates = updatedAggregates

	// Update the subscribers.
	for id, count := range updates {
		for sub := range c.subscribers {
			sub.Callback(id, count)
		}
	}
}
