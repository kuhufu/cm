package server

import (
	"sync"
)

type Room struct {
	Id      string
	members map[string]*Channel
	mu      sync.RWMutex
}

func NewRoom(id string) *Room {
	return &Room{
		Id:      id,
		members: map[string]*Channel{},
	}
}

func (c *Room) Add(id string, channel *Channel) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.members[id] = channel
}

func (c *Room) AddOrReplace(id string, channel *Channel) *Channel {
	c.mu.Lock()
	defer c.mu.Unlock()

	old, _ := c.members[id]
	c.members[id] = channel
	return old
}

func (c *Room) Del(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.members[id]
	if ok {
		delete(c.members, id)
	}

	return ok
}

func (c *Room) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.members)
}

func (c *Room) DelIfEqual(id string, channel *Channel) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	val, ok := c.members[id]
	if ok && val == channel {
		delete(c.members, id)
	}

	return ok
}

func (c *Room) Exist(id string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, ok := c.members[id]
	return ok
}

func (c *Room) Get(id string) (*Channel, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.members[id]
	return val, ok
}

func (c *Room) Range(f func(id string, channel *Channel) bool) {
	c.mu.RLock()
	size := len(c.members)
	if size == 0 {
		return
	}

	keys := make([]string, 0, size)
	vals := make([]*Channel, 0, size)

	for key, val := range c.members {
		keys = append(keys, key)
		vals = append(vals, val)
	}
	c.mu.RUnlock()

	for i := 0; i < len(keys); i++ {
		if !f(keys[i], vals[i]) {
			return
		}
	}
}
