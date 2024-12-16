package api

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var mu sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Убедитесь, что правильно обрабатываете CORS
	},
}

// Хранилище активных соединений
var connections = make(map[uint]*websocket.Conn) // Map userID -> WebSocket connection

func (h *Handler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(uint) // Получаем userId из контекста

	conn, err := upgrader.Upgrade(w, r, nil) // Обновляем соединение до WebSocket
	if err != nil {
		log.Println("Ошибка обновления соединения:", err)
		http.Error(w, "Could not upgrade connection", http.StatusInternalServerError)
		return
	}

	mu.Lock()
	connections[userId] = conn
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(connections, userId)
		mu.Unlock()
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Ошибка чтения сообщения: %v\n", err)
			break
		}
	}
}
