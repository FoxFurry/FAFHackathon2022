package connections

import (
	"sync"

	"github.com/gorilla/websocket"
)

var (
	currentConnections = make(map[string]*websocket.Conn)
	connectionsMutex   = sync.RWMutex{}
)

func GetConnection(name string) (*websocket.Conn, bool) {
	connectionsMutex.RLock()
	defer connectionsMutex.RUnlock()
	res, ok := currentConnections[name]
	return res, ok
}

func SetConnection(name string, conn *websocket.Conn) {
	connectionsMutex.Lock()
	currentConnections[name] = conn
	connectionsMutex.Unlock()
}

func DeleteConnection(name string) {
	connectionsMutex.Lock()
	delete(currentConnections, name)
	connectionsMutex.Unlock()
}
