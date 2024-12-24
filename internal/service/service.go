package service

import (
	"fmt"
	"project/cmd/database/model"
	"project/internal/consts"
	"project/internal/helpers"
	repository "project/internal/repositories"
	"slices"
	"time"
)

type Service struct {
	Repository *repository.Repository
}

func NewService(repository *repository.Repository) *Service {
	return &Service{
		Repository: repository,
	}
}

func (s *Service) CreateOrGetUser(user *model.User) (*model.User, error) {
	existUser, err := s.Repository.GetUser(user)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if existUser == nil {
		createdUser, err := s.Repository.CreateUser(user)
		if err != nil {
			return nil, err
		}

		return createdUser, nil
	}

	return existUser, nil
}

func (s *Service) GetConversations(userId uint) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	users, err := s.Repository.GetConversationsUsers(userId)
	if err != nil {
		return nil, err
	}

	groups, err := s.Repository.GetGroups(userId)
	if err != nil {
		return nil, err
	}
	res["users"] = users
	res["groups"] = groups

	return res, nil
}

func (s *Service) GetUsers(query string, userId uint) (*[]model.User, error) {
	return s.Repository.GetUsers(query, userId)
}

func (s *Service) ValidateUser(user *model.User) (string, error) {
	token, err := helpers.GenerateSessionToken(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) GetMessages(user1ID uint, payload *helpers.GetMessagesRequest) (*[]helpers.MessagesResponse, error) {
	var messages *[]model.Message
	var err error

	if !payload.IsGroup {
		messages, err = s.Repository.GetPrivateMessages(user1ID, payload.UserId)
		if err != nil {
			return nil, err
		}
	} else {
		messages, err = s.Repository.GetGroupMessages(payload.UserId)
		if err != nil {
			return nil, err
		}
	}

	var response []helpers.MessagesResponse
	if messages != nil {
		for _, message := range *messages {
			if message.MessageType == consts.MessageTypeText {
				response = append(response, helpers.MessagesResponse{
					Message:        message.Content,         // Message text
					UserId:         message.SenderID,        // Sender user ID
					ConversationId: message.ConversationID,  // Conversation ID
					Username:       message.Sender.Username, // Conversation ID
					CreatedAt:      message.CreatedAt,       // Conversation ID
					IsPhoto:        false,
					IsRead:         message.IsRead,
				})
			} else {
				response = append(response, helpers.MessagesResponse{
					Message:        message.Content,         // Message text
					UserId:         message.SenderID,        // Sender user ID
					ConversationId: message.ConversationID,  // Conversation ID
					Username:       message.Sender.Username, // Conversation ID
					CreatedAt:      message.CreatedAt,       // Conversation ID
					IsPhoto:        true,
					IsRead:         message.IsRead,
				})
			}
		}
	}

	return &response, nil
}

func (s *Service) SendMessage(userId uint, input *helpers.SendMessageRequest) (bool, error) {
	var conv *model.Conversation
	var err error

	if !input.IsGroup {
		conv, err = s.Repository.CheckPrivateConversation(userId, input.ToUserId)
		if err != nil {
			return false, err
		}
	} else {
		conv, err = s.Repository.CheckGroupConversation(input.GroupId)
		if err != nil {
			return false, err
		}
	}
	var message model.Message

	if input.PhotoPath == "" {
		message = model.Message{
			ConversationID: conv.ID,
			SenderID:       userId,
			Content:        input.Text,
			MessageType:    consts.MessageTypeText,
			CreatedAt:      time.Now(),
		}
	} else {
		message = model.Message{
			ConversationID: conv.ID,
			SenderID:       userId,
			Content:        input.PhotoPath,
			MessageType:    consts.MessageTypePhoto,
			CreatedAt:      time.Now(),
		}
	}

	return s.Repository.CreateMessage(&message)
}

func (s *Service) CreateGroups(payload *helpers.CreateGroupRequest, addedById uint) (bool, error) {
	group, err := s.Repository.CreateGroup(payload, addedById)
	if err != nil {
		return false, err
	}

	_, err = s.Repository.CreateGroupMembers(addedById, addedById, group.ID)
	if err != nil {
		return false, err
	}

	for _, user := range payload.Users {
		_, err = s.Repository.CreateGroupMembers(user.ID, addedById, group.ID)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func (s *Service) GetGroupMembers(groupId uint) (*[]model.GroupMember, error) {
	return s.Repository.GetGroupMembers(groupId)
}

func (s *Service) UpdateUserProfile(userId uint, filepath string) (bool, error) {
	var user model.User
	user.ID = userId
	user.ProfilePhotoURL = filepath
	return s.Repository.UpdateUserProfile(&user)
}

func (s *Service) LeaveGroup(userId uint, payload *helpers.GroupRequest) (bool, error) {
	return s.Repository.DeleteGroupMember(userId, payload.Group)
}

func (s *Service) AddUsersToGroup(userId uint, payload *helpers.AddUsersToGroup) (bool, error) {
	users, err := s.Repository.GetGroupMembers(payload.GroupId)
	if err != nil {
		return false, err
	}
	var userIds = make([]uint, len(*users))

	for _, user := range *users {
		userIds = append(userIds, user.UserID)
	}

	for _, v := range payload.Users {
		if !slices.Contains(userIds, v.ID) {
			_, err = s.Repository.CreateGroupMembers(v.ID, userId, payload.GroupId)
			if err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

func (s *Service) MarkAsRead(userId uint, input *helpers.SendMessageRequest) (bool, error) {
	var conv *model.Conversation
	var err error

	if !input.IsGroup {
		conv, err = s.Repository.CheckPrivateConversation(userId, input.ToUserId)
		if err != nil {
			return false, err
		}
	} else {
		conv, err = s.Repository.CheckGroupConversation(input.GroupId)
		if err != nil {
			return false, err
		}
	}

	return s.Repository.MarkAsRead(conv.ID, userId)
}
