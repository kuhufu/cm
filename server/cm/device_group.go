package cm

import "sync"

type DeviceGroup struct {
	UserId UserId
	inner  map[ConnId]*Conn
	mu     sync.RWMutex
}

func NewDeviceGroup(userId UserId) *DeviceGroup {
	return &DeviceGroup{
		UserId: userId,
		inner:  map[ConnId]*Conn{},
	}
}

func (g *DeviceGroup) Add(id ConnId, conn *Conn) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.inner[id] = conn
}

func (g *DeviceGroup) Get(id ConnId) (*Conn, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	conn, ok := g.inner[id]

	return conn, ok
}

func (g *DeviceGroup) Del(id ConnId) {
	g.mu.Lock()
	defer g.mu.Unlock()

	delete(g.inner, id)
}

func (g *DeviceGroup) Size() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.inner)
}

func (g *DeviceGroup) Items() []*Conn {
	g.mu.RLock()
	defer g.mu.RUnlock()

	conns := make([]*Conn, 0, len(g.inner))
	for _, g := range g.inner {
		conns = append(conns, g)
	}
	return conns
}

func (g *DeviceGroup) ForEach(f func(conn *Conn)) {
	conns := g.Items() //通常foreach中会做一些耗时的操作，这样做可以避免锁住太久，空间换时间

	for _, conn := range conns {
		f(conn)
	}
}
