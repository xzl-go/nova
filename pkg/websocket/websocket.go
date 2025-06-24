package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocketServer 封装 WebSocket 服务端
// 支持多客户端连接、广播、单发、连接管理

type WebSocketServer struct {
	upgrader  websocket.Upgrader
	clients   map[*websocket.Conn]bool
	lock      sync.RWMutex
	broadcast chan []byte
}

// NewWebSocketServer 创建 WebSocket 服务端
func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
	}
}

// Handle 处理 WebSocket 连接升级和消息
func (ws *WebSocketServer) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	ws.lock.Lock()
	ws.clients[conn] = true
	ws.lock.Unlock()
	go ws.readPump(conn)
}

// readPump 读取客户端消息并广播
func (ws *WebSocketServer) readPump(conn *websocket.Conn) {
	defer func() {
		ws.lock.Lock()
		delete(ws.clients, conn)
		ws.lock.Unlock()
		conn.Close()
	}()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		ws.broadcast <- message
	}
}

// Start 启动广播协程
func (ws *WebSocketServer) Start() {
	go func() {
		for {
			msg := <-ws.broadcast
			ws.lock.RLock()
			for client := range ws.clients {
				err := client.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					client.Close()
					delete(ws.clients, client)
				}
			}
			ws.lock.RUnlock()
		}
	}()
}

// SendToAll 主动广播消息
func (ws *WebSocketServer) SendToAll(msg []byte) {
	ws.broadcast <- msg
}

// ClientCount 获取当前连接数
func (ws *WebSocketServer) ClientCount() int {
	ws.lock.RLock()
	defer ws.lock.RUnlock()
	return len(ws.clients)
}
