package cm

import "sync"

type Group struct {
	Id    GroupId
	inner map[UserId]*DeviceGroup
	mu    *sync.RWMutex
}

func NewGroup(id GroupId) *Group {
	return &Group{
		Id:    id,
		inner: map[UserId]*DeviceGroup{},
		mu:    &sync.RWMutex{},
	}
}

func (g *Group) Get(id UserId) *DeviceGroup {
	g.mu.RLock()
	g.mu.RUnlock()

	return g.inner[id]
}

func (g *Group) Add(id UserId, group *DeviceGroup) {
	g.mu.Lock()
	g.mu.Unlock()

	g.inner[id] = group
}

func (g *Group) Del(id UserId) {
	g.mu.Lock()
	g.mu.Unlock()

	delete(g.inner, id)
}

func (g *Group) Size() int {
	g.mu.RLock()
	g.mu.RUnlock()

	return len(g.inner)
}

func (g *Group) Items() []*DeviceGroup {
	groups := make([]*DeviceGroup, 0, len(g.inner))
	g.mu.RLock()
	for _, g := range g.inner {
		groups = append(groups, g)
	}
	g.mu.RUnlock()
	return groups
}

func (g *Group) ForEach(f func(id UserId, g *DeviceGroup)) {
	groups := g.Items() //通常foreach中会做一些耗时的操作，这样做可以避免锁住太久，空间换时间

	for _, g := range groups {
		f(g.UserId, g)
	}
}
