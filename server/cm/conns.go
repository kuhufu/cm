package cm

import "sync"

type Conns struct {
	inner map[ConnId]*Conn
	mu    *sync.RWMutex
}

func NewConns() *Conns {
	return &Conns{
		inner: map[ConnId]*Conn{},
		mu:    &sync.RWMutex{},
	}
}

func (cs *Conns) Get(id ConnId) (*Conn, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	c, ok := cs.inner[id]
	return c, ok
}

func (cs *Conns) Add(id ConnId, conn *Conn)  {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.inner[id] = conn
}

func (cs *Conns) Del(conn *Conn) bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()


	if c, ok := cs.inner[conn.Id]; ok && conn == c {
		delete(cs.inner, conn.Id)
		return true
	}

	return false
}
