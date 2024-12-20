package api

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"sync"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Хранилище соединений и групп
var (
	connections = make(map[uint]*websocket.Conn)          // userID -> WebSocket connection
	groupConn   = make(map[uint]map[uint]*websocket.Conn) // groupID -> userID -> WebSocket connection
	mu          sync.Mutex                                // Mutex для безопасного доступа
)

func (h *Handler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	var groupIdUint uint
	userId := r.Context().Value("userId").(uint) // Получаем userId из контекста
	groupIdStr := r.URL.Query().Get("groupId")
	if groupIdStr != "" {
		groupId, err := strconv.ParseUint(groupIdStr, 10, 32) // Преобразуем строку в uint
		if err != nil {
			log.Println("Невозможно преобразовать groupId в uint:", err)
			http.Error(w, "Invalid groupId", http.StatusBadRequest)
			return
		}

		// Преобразуем groupId в uint
		groupIdUint = uint(groupId)
	}

	conn, err := upgrader.Upgrade(w, r, nil) // Обновляем соединение до WebSocket
	if err != nil {
		log.Println("Ошибка обновления соединения:", err)
		http.Error(w, "Could not upgrade connection", http.StatusInternalServerError)
		return
	}

	mu.Lock()
	connections[userId] = conn // Добавляем пользователя в глобальное хранилище
	if groupIdStr != "" {
		// Инициализация группы, если не существует
		if groupConn[groupIdUint] == nil {
			groupConn[groupIdUint] = make(map[uint]*websocket.Conn)
		}
		groupConn[groupIdUint][userId] = conn // Добавляем соединение в группу
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(connections, userId) // Удаляем пользователя из глобального хранилища
		if groupIdStr != "" {
			delete(groupConn[groupIdUint], userId) // Удаляем соединение из группы
		}
		mu.Unlock()
		conn.Close()
	}()

	// Чтение сообщений от клиента
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Ошибка чтения сообщения: %v\n", err)
			break
		}

		// Отправка сообщения всем участникам группы, если это групповое сообщение
		if groupIdStr != "" {
			h.broadcastToGroup(groupIdUint, userId, message)
		}
	}
}

// Функция для отправки сообщений всем участникам группы
func (h *Handler) broadcastToGroup(groupId uint, senderId uint, message []byte) {
	mu.Lock()
	defer mu.Unlock()

	// Перебираем всех участников группы и отправляем им сообщение
	for userId, conn := range groupConn[groupId] {
		// Не отправляем сообщение самому себе
		if userId == senderId {
			continue
		}
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Ошибка отправки сообщения пользователю %d в группе %s: %v\n", userId, groupId, err)
			delete(groupConn[groupId], userId) // Если ошибка, удаляем соединение
		}
	}
}
