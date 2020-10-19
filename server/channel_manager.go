package server

import (
	"sync"
)

type Manager struct {
	mu    sync.RWMutex
	rooms map[RoomId]*Room
}

func NewManager() *Manager {
	return &Manager{
		rooms: map[RoomId]*Room{},
	}
}

func (m *Manager) Add(id RoomId, channel *Channel) {
	m.GetOrCreate(id).Add(channel.Id, channel)
	m.mu.Lock()
}

func (m *Manager) Get(id RoomId) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.rooms[id]

	return val, ok
}

func (m *Manager) Del(id RoomId) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, id)
}

func (m *Manager) Exist(id RoomId) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.rooms[id]

	return ok
}

func (m *Manager) GetOrCreate(id RoomId) *Room {
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

func (m *Manager) Range(f func(key RoomId, val *Room)) {
	m.mu.RLock()
	size := len(m.rooms)

	keys := make([]RoomId, 0, size)
	vals := make([]*Room, 0, size)

	for key, val := range m.rooms {
		keys = append(keys, key)
		vals = append(vals, val)
	}
	m.mu.RUnlock()

	for i := 0; i < len(keys); i++ {
		f(keys[i], vals[i])
	}
}
