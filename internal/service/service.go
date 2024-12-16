package service

import (
	"fmt"
	"project/cmd/database/model"
	"project/internal/consts"
	"project/internal/helpers"
	repository "project/internal/repositories"
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

func (s *Service) CreateOrUpdateUser(user *model.User) (string, error) {
	existUser, err := s.Repository.GetUser(user)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	if existUser == nil {
		createdUser, err := s.Repository.CreateUser(user)
		if err != nil {
			return "", err
		}

		return helpers.GenerateSessionToken(createdUser)
	}

	return helpers.GenerateSessionToken(existUser)
}

func (s *Service) GetConversations(userId uint) (*[]model.User, error) {
	return s.Repository.GetConversations(userId)
}

func (s *Service) GetUsers(query string) (*[]model.User, error) {
	return s.Repository.GetUsers(query)
}

func (s *Service) ValidateUser(user *model.User) (string, error) {
	token, err := helpers.GenerateSessionToken(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) GetMessages(user1ID uint, user2ID uint) (*[]helpers.MessagesResponse, error) {
	messages, err := s.Repository.GetMessages(user1ID, user2ID)
	if err != nil {
		return nil, err
	}

	var response []helpers.MessagesResponse
	for _, message := range messages {
		response = append(response, helpers.MessagesResponse{
			Message:        message.Content,        // Message text
			UserId:         message.SenderID,       // Sender user ID
			ConversationId: message.ConversationID, // Conversation ID
		})
	}

	return &response, nil
}

func (s *Service) SendMessage(userId uint, input *helpers.SendMessageRequest) (bool, error) {
	conv, err := s.Repository.CheckConversation(userId, input.ToUserId)
	if err != nil {
		return false, err
	}

	message := model.Message{
		ConversationID: conv.ID,
		SenderID:       userId,
		Content:        input.Text,
		MessageType:    consts.MessageTypeText,
		CreatedAt:      time.Now(),
	}

	return s.Repository.CreateMessage(&message)
}
