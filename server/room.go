package server

import (
	"sync"
)

type RoomId = string

type Room struct {
	Id      RoomId
	clients map[ChannelId]*Channel
	mu      sync.RWMutex
}

func NewRoom(id RoomId) *Room {
	return &Room{
		Id:      id,
		clients: map[ChannelId]*Channel{},
	}
}

func (c *Room) Add(id ChannelId, channel *Channel) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.clients[id] = channel
}

func (c *Room) AddOrReplace(id ChannelId, channel *Channel) *Channel {
	c.mu.Lock()
	defer c.mu.Unlock()

	old, _ := c.clients[id]
	c.clients[id] = channel
	return old
}

func (c *Room) Del(id ChannelId) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.clients[id]
	if ok {
		delete(c.clients, id)
	}

	return ok
}

func (c *Room) DelIfEqual(id ChannelId, channel *Channel) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	val, ok := c.clients[id]
	if ok && val == channel {
		delete(c.clients, id)
	}

	return ok
}

func (c *Room) Exist(id ChannelId) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, ok := c.clients[id]
	return ok
}

func (c *Room) Get(id ChannelId) (*Channel, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.clients[id]
	return val, ok
}

func (c *Room) Range(f func(id ChannelId, channel *Channel)) {
	c.mu.RLock()
	size := len(c.clients)

	keys := make([]ChannelId, 0, size)
	vals := make([]*Channel, 0, size)

	for key, val := range c.clients {
		keys = append(keys, key)
		vals = append(vals, val)
	}
	c.mu.RUnlock()

	for i := 0; i < len(keys); i++ {
		f(keys[i], vals[i])
	}
}
