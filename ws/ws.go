package ws

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ConnectionPool struct {
	connections map[*websocket.Conn]bool
	mu          sync.Mutex
}

func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[*websocket.Conn]bool),
	}
}

func (pool *ConnectionPool) AddConnection(conn *websocket.Conn) {
	pool.mu.Lock()
	pool.connections[conn] = true
	pool.mu.Unlock()
}

func (pool *ConnectionPool) RemoveConnection(conn *websocket.Conn) {
	pool.mu.Lock()
	delete(pool.connections, conn)
	pool.mu.Unlock()
}

func (pool *ConnectionPool) Broadcast(message []byte) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	for conn := range pool.connections {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			conn.Close()
			delete(pool.connections, conn)
		}
	}
}

var connectionPool = NewConnectionPool()

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to upgrade connection:", err)
		return
	}
	defer func() {
		connectionPool.RemoveConnection(conn)
		conn.Close()
	}()

	connectionPool.AddConnection(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func StartWebSocketServer(port string) {
	http.HandleFunc("/ws", wsHandler)
	http.ListenAndServe(":"+port, nil)
}

func BroadcastNewBlock(blockNumber string) {
	message := fmt.Sprintf("New block: %s", blockNumber)
	connectionPool.Broadcast([]byte(message))
}
