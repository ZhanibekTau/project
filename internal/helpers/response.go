package helpers

type MessagesResponse struct {
	Message        string `json:"message"`
	UserId         uint   `json:"user_id"`
	ConversationId uint   `json:"conversation_id"`
}
