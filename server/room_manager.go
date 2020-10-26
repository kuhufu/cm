package server

import (
	"sync"
)

type Manager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewManager() *Manager {
	return &Manager{
		rooms: map[string]*Room{},
	}
}

func (m *Manager) Add(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rooms[id] = NewRoom(id)
}

func (m *Manager) Get(id string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.rooms[id]

	return val, ok
}

func (m *Manager) Del(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, id)
}

func (m *Manager) Exist(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.rooms[id]

	return ok
}

func (m *Manager) GetOrCreate(id string) *Room {
	m.mu.RLock()
	if val, ok := m.rooms[id]; ok {
		m.mu.RUnlock()
		return val
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()
	if val, ok := m.rooms[id]; ok {
		return val
	}

	c := NewRoom(id)
	m.rooms[id] = c
	return c
}

func (m *Manager) Range(f func(key string, val *Room) bool) {
	m.mu.RLock()
	size := len(m.rooms)
	if size == 0 {
		return
	}

	keys := make([]string, 0, size)
	vals := make([]*Room, 0, size)

	for key, val := range m.rooms {
		keys = append(keys, key)
		vals = append(vals, val)
	}
	m.mu.RUnlock()

	for i := 0; i < len(keys); i++ {
		if !f(keys[i], vals[i]) {
			return
		}
	}
}
