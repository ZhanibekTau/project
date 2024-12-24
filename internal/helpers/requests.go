package helpers

import "project/cmd/database/model"

type GetMessagesRequest struct {
	UserId  uint `json:"id"`
	IsGroup bool `json:"isGroup"`
}

type SendMessageRequest struct {
	Text      string `json:"text"`
	ToUserId  uint   `json:"toUserId"`
	IsGroup   bool   `json:"isGroup"`
	GroupId   uint   `json:"groupId"`
	PhotoPath string `json:"photoPath"`
}

type CreateGroupRequest struct {
	GroupName      string       `json:"groupName"`
	GroupPhotoPath string       `json:"groupPhotoPath"`
	Users          []model.User `json:"users"`
}

type GroupRequest struct {
	Group model.Group `json:"group"`
}

type AddUsersToGroup struct {
	GroupId uint         `json:"groupId"`
	Users   []model.User `json:"users"`
}

type DeleteMessage struct {
	MessageId uint `json:"id"`
}
