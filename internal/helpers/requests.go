package helpers

type GetMessagesRequest struct {
	UserId uint `json:"userId"`
}

type SendMessageRequest struct {
	Text     string `json:"text"`
	ToUserId uint   `json:"toUserId"`
}
