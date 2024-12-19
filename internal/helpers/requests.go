package helpers

import "project/cmd/database/model"

type GetMessagesRequest struct {
	UserId  uint `json:"id"`
	IsGroup bool `json:"isGroup"`
}

type SendMessageRequest struct {
	Text     string `json:"text"`
	ToUserId uint   `json:"toUserId"`
	IsGroup  bool   `json:"isGroup"`
}

type CreateGroupRequest struct {
	GroupName string       `json:"groupName"`
	Users     []model.User `json:"users"`
}
