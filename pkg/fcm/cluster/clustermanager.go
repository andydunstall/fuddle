package cluster

import (
	"sync"
)

type Manager struct {
	clusters map[string]*Cluster

	// mu is a mutex protecting the fields above.
	mu sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		clusters: make(map[string]*Cluster),
	}
}

func (m *Manager) Get(id string) (*Cluster, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.clusters[id]
	return c, ok
}

func (m *Manager) Add(c *Cluster) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clusters[c.ID()] = c
}

func (m *Manager) Delete(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.clusters[id]
	if !ok {
		return false
	}

	delete(m.clusters, id)
	c.Shutdown()
	return true
}

func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, c := range m.clusters {
		c.Shutdown()
	}
}
