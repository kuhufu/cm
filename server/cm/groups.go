package cm

import "sync"

type Groups struct {
	inner map[GroupId]*Group
	mu    *sync.RWMutex
}

func NewGroups() *Groups {
	return &Groups{
		inner: map[GroupId]*Group{},
		mu:    &sync.RWMutex{},
	}
}

func (gs *Groups) Get(id GroupId) (*Group, bool) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	group, ok := gs.inner[id]
	return group, ok
}

func (gs *Groups) GetOrCreate(id GroupId) *Group {
	if group, ok := gs.Get(id); ok {
		return group
	}

	gs.mu.Lock()
	defer gs.mu.Unlock()
	if group, ok := gs.inner[id]; ok { //双重检查锁
		return group
	}

	group := NewGroup(id)
	gs.inner[id] = group

	return group
}

func (gs *Groups) Del(id GroupId) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	delete(gs.inner, id)
}