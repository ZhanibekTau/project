package helpers

import "time"

type MessagesResponse struct {
	Message        string    `json:"message"`
	IsPhoto        bool      `json:"is_photo"`
	UserId         uint      `json:"user_id"`
	ConversationId uint      `json:"conversation_id"`
	Username       string    `json:"username"`
	CreatedAt      time.Time `json:"createdAt"`
}
