package repository

import (
	"github.com/jinzhu/gorm"
	"project/cmd/database/model"
	"strings"
)

type Repository struct {
	database *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		database: db,
	}
}

func (r *Repository) CreateUser(user *model.User) (*model.User, error) {
	result := r.database.Create(user)
	if result.Error != nil {
		msg := result.Error
		return nil, msg
	}

	return user, nil
}

func (r *Repository) GetUser(user *model.User) (*model.User, error) {
	result := r.database.Model(&model.User{}).Where("username = ?", user.Username).Find(&user)
	if result.Error != nil {
		msg := result.Error
		if result.RecordNotFound() {
			return nil, nil
		}

		return nil, msg
	}

	return user, nil
}

func (r *Repository) GetConversations(userId uint) (*[]model.User, error) {
	var users []model.User

	if err := r.database.Raw(`
        SELECT DISTINCT u.*
        FROM users u
        JOIN conversations c ON (u.id = c.user1_id OR u.id = c.user2_id)
        WHERE (c.user1_id = ? OR c.user2_id = ?)
          AND u.id != ?`, userId, userId, userId).Scan(&users).Error; err != nil {
		return nil, err
	}

	return &users, nil
}

func (r *Repository) GetUsers(query string) (*[]model.User, error) {
	var users []model.User

	if err := r.database.Where("LOWER(username) LIKE ?", "%"+strings.ToLower(query)+"%").Find(&users).Error; err != nil {
		return nil, err
	}

	return &users, nil
}
