package cm

import "sync"

type GroupId = string
type ConnId = string
type UserId = string

type ConnManager struct {
	connMap      *Conns
	deviceGroups *DeviceGroups
	groups       *Groups
	mu           sync.RWMutex
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		connMap:      NewConns(),
		deviceGroups: NewDeviceGroups(),
		groups:       NewGroups(),
	}
}

func (m *ConnManager) GetDeviceGroup(userId UserId) (*DeviceGroup, bool) {
	return m.deviceGroups.Get(userId)
}

//添加到群组
func (m *ConnManager) AddToGroupNoSync(userId UserId, groupIds []string) {
	for _, groupId := range groupIds {
		m.groups.GetOrCreate(groupId).Add(userId, m.deviceGroups.GetOrCreate(userId))
	}
}

func (m *ConnManager) AddToGroup(userId UserId, groupIds []string) {
	m.mu.Lock()
	m.mu.Unlock()
	for _, groupId := range groupIds {
		m.groups.GetOrCreate(groupId).Add(userId, m.deviceGroups.GetOrCreate(userId))
	}
}

//从群组中移除
func (m *ConnManager) RemoveFromGroup(userId UserId, groupIds ...string) {
	for _, groupId := range groupIds {
		if group, ok := m.GetGroup(groupId); ok {
			group.Del(userId)
		}
	}
}

func (m *ConnManager) GetGroup(id GroupId) (*Group, bool) {
	return m.groups.Get(id)
}

func (m *ConnManager) AddOrReplaceNoSync(connId ConnId, conn *Conn) *Conn {
	var oldConn *Conn
	var ok bool
	if oldConn, ok = m.connMap.Get(connId); ok {
		m.remove(oldConn) //删除旧连接
	}

	m.connMap.Add(connId, conn)

	m.addToDeviceGroup(conn)

	return oldConn
}

func (m *ConnManager) AddOrReplace(connId ConnId, conn *Conn) *Conn {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.AddOrReplaceNoSync(connId, conn)
}

func (m *ConnManager) RemoveConn(conn *Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.remove(conn)
}

func (m *ConnManager) GetConn(connId ConnId) (*Conn, bool) {
	return m.connMap.Get(connId)
}

func (m *ConnManager) With(f func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	f()
}

//添加到设备组
func (m *ConnManager) addToDeviceGroup(conn *Conn) {
	deviceGroup := m.deviceGroups.GetOrCreate(conn.UserId)
	deviceGroup.Add(conn.Id, conn)
}

//从设备组移除
func (m *ConnManager) removeFromDeviceGroup(conn *Conn) {
	if deviceGroup, ok := m.deviceGroups.Get(conn.UserId); ok {
		if c, ok := deviceGroup.Get(conn.Id); ok && c == conn {
			deviceGroup.Del(conn.Id)
		}

		//如果设备组为空则删除设备组
		if deviceGroup.Size() == 0 {
			m.deviceGroups.Del(conn.UserId)
		}
	}
}

func (m *ConnManager) remove(conn *Conn) {
	if m.connMap.Del(conn) {
		m.removeFromDeviceGroup(conn)
	}
}

func (m *ConnManager) AllConn() []*Conn {
	return m.connMap.All()
}
