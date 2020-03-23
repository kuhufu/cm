package cm

import "sync"

type GroupId = string
type ConnId = string
type UserId = string

type DeviceGroup struct {
	UserId UserId
	inner  map[ConnId]*Conn
	mu     *sync.RWMutex
}

type Group struct {
	inner map[UserId]*DeviceGroup
	mu    *sync.RWMutex
}

type ConnManager struct {
	IdMap map[ConnId]*Conn

	userDevice map[UserId]*DeviceGroup

	groups map[GroupId]*Group

	mu *sync.RWMutex
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		IdMap: map[ConnId]*Conn{},

		userDevice: map[UserId]*DeviceGroup{},
		groups:     map[GroupId]*Group{},
		mu:         &sync.RWMutex{},
	}
}

func (m *ConnManager) GetDeviceGroup(userId UserId) (*DeviceGroup, bool) {
	group, ok := m.userDevice[userId]
	return group, ok
}

func (m *ConnManager) getOrCreateDeviceGroup(userId UserId) *DeviceGroup {
	var group *DeviceGroup
	var ok bool
	if group, ok = m.GetDeviceGroup(userId); !ok {
		group = NewDeviceGroup(userId)
		m.userDevice[userId] = group
	}

	return group
}

//添加到设备组
func (m *ConnManager) AddToDeviceGroup(conn *Conn) {
	deviceGroup := m.getOrCreateDeviceGroup(conn.UserId)
	deviceGroup.Add(conn.Id, conn)
}

//从设备组移除
func (m *ConnManager) RemoveFromDeviceGroup(conn *Conn) {
	if deviceGroup, ok := m.GetDeviceGroup(conn.UserId); ok {
		if c, ok := deviceGroup.Get(conn.Id); ok && c == conn {
			deviceGroup.Del(conn.Id)
		}

		//如果设备组为空则删除设备组
		if deviceGroup.Size() == 0 {
			delete(m.userDevice, conn.UserId)
		}
	}
}

//添加到群组
func (m *ConnManager) AddToGroup(userId UserId, groupIds []string) {
	for _, groupId := range groupIds {
		m.getOrCreateGroup(groupId).Add(userId, m.userDevice[userId])
	}
}

//从群组中移除
func (m *ConnManager) RemoveFromGroup(userId UserId, groupIds []string) {
	for _, groupId := range groupIds {
		if group, ok := m.GetGroup(groupId); ok {
			group.Del(userId)
		}
	}
}

func (m *ConnManager) GetGroup(id GroupId) (*Group, bool) {
	group, ok := m.groups[id]

	return group, ok
}

func (m *ConnManager) getOrCreateGroup(id GroupId) *Group {
	var group *Group
	var ok bool
	if group, ok = m.GetGroup(id); !ok {
		group = NewGroup()
		m.groups[id] = group
	}

	return group
}

func (m *ConnManager) AddOrReplace(connId ConnId, conn *Conn) *Conn {
	var oldConn *Conn
	var ok bool
	if oldConn, ok = m.IdMap[connId]; ok {
		m.Remove(oldConn) //删除旧连接
	}

	m.IdMap[connId] = conn

	m.AddToDeviceGroup(conn)

	return oldConn
}

func (m *ConnManager) Remove(conn *Conn) {
	if c, ok := m.IdMap[conn.Id]; ok && c == conn {
		delete(m.IdMap, conn.Id)
		m.RemoveFromDeviceGroup(conn)
	}
}

func (m *ConnManager) Get(connId ConnId) (*Conn, bool) {
	conn, ok := m.IdMap[connId]

	return conn, ok
}

func (m *ConnManager) AddOrReplaceSync(connId ConnId, conn *Conn) *Conn {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.AddOrReplace(connId, conn)
}

func (m *ConnManager) RemoveSync(conn *Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Remove(conn)
}

func (m *ConnManager) GetSync(connId ConnId) (*Conn, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.Get(connId)
}

func (m *ConnManager) WithSync(f func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	f()
}
