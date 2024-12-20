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
	GroupId  uint   `json:"groupId"`
}

type CreateGroupRequest struct {
	GroupName      string       `json:"groupName"`
	GroupPhotoPath string       `json:"groupPhotoPath"`
	Users          []model.User `json:"users"`
}

type GroupRequest struct {
	Group model.Group `json:"group"`
}
