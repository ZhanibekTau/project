package service

import (
	"fmt"
	"project/cmd/database/model"
	"project/internal/helpers"
	repository "project/internal/repositories"
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
