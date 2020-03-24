package cm

import "sync"

type DeviceGroups struct {
	inner map[UserId]*DeviceGroup
	mu    *sync.RWMutex
}

func NewDeviceGroups() *DeviceGroups {
	return &DeviceGroups{
		inner: map[UserId]*DeviceGroup{},
		mu:    &sync.RWMutex{},
	}
}

func (d *DeviceGroups) Get(id UserId) (*DeviceGroup, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	g, ok := d.inner[id]

	return g, ok
}

func (d *DeviceGroups) GetOrCreate(id UserId) *DeviceGroup {
	if group, ok := d.Get(id); ok {
		return group
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if group, ok := d.inner[id]; ok { //双重检查锁
		return group
	}

	group := NewDeviceGroup(id)
	d.inner[id] = group

	return group
}

func (d *DeviceGroups) Del(id UserId) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.inner, id)
}
