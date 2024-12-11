package service

import (
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

func (s *Service) GetString() (string, error) {
	return "Salem Jake!", nil
}
