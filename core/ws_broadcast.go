// core/ws_broadcast.go

package core

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type BlockInfo struct {
	BlockNumber uint64 `json:"blockNumber"`
	ToAddress   string `json:"toAddress"`
}

type WsHub struct {
	clients   map[*websocket.Conn]bool
	mutex     sync.Mutex
	broadcast chan BlockInfo
}

func NewWsHub() *WsHub {
	return &WsHub{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan BlockInfo),
	}
}

func (h *WsHub) Run() {
	for {
		select {
		case blockInfo := <-h.broadcast:
			h.mutex.Lock()
			message, _ := json.Marshal(blockInfo)
			for client := range h.clients {
				client.WriteMessage(websocket.TextMessage, message)
			}
			h.mutex.Unlock()
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *WsHub) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.mutex.Lock()
	h.clients[conn] = true
	h.mutex.Unlock()

	conn.SetCloseHandler(func(code int, text string) error {
		h.mutex.Lock()
		delete(h.clients, conn)
		h.mutex.Unlock()
		return nil
	})
}

func (h *WsHub) BroadcastBlockInfo(blockInfo BlockInfo) {
	h.broadcast <- blockInfo
}
