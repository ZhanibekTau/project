package repository

import (
	"errors"
	"github.com/jinzhu/gorm"
	"log"
	"project/cmd/database/model"
	"project/internal/helpers"
	"strings"
	"time"
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

func (r *Repository) GetConversationsUsers(userId uint) (*[]model.User, error) {
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

func (r *Repository) GetGroups(userId uint) (*[]model.Group, error) {
	var groups []model.Group

	if err := r.database.Raw(`
        SELECT DISTINCT g.*
        FROM groups g
        JOIN group_members gm ON g.id = gm.group_id
        WHERE gm.user_id = ?`, userId).Scan(&groups).Error; err != nil {
		return nil, err
	}

	return &groups, nil
}

func (r *Repository) GetUsers(query string, userId uint) (*[]model.User, error) {
	var users []model.User

	if err := r.database.Where("LOWER(username) LIKE ? AND id != ?", "%"+strings.ToLower(query)+"%", userId).Find(&users).Error; err != nil {
		return nil, err
	}

	return &users, nil
}

func (r *Repository) GetPrivateMessages(user1ID uint, user2ID uint) (*[]model.Message, error) {
	// Step 1: Find the conversation between user1 and user2
	var conversation model.Conversation
	err := r.database.Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)", user1ID, user2ID, user2ID, user1ID).First(&conversation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Println("Error fetching conversation:", err)
		return nil, err
	}

	// Step 2: Fetch messages for this conversation
	var messages []model.Message
	err = r.database.Where("conversation_id = ?", conversation.ID).Order("created_at asc").Preload("Sender").Find(&messages).Error
	if err != nil {
		log.Println("Error fetching messages:", err)
		return nil, err
	}

	return &messages, nil
}

func (r *Repository) GetGroupMessages(groupID uint) (*[]model.Message, error) {
	var conversation model.Conversation
	err := r.database.Where("group_id = ?", groupID).First(&conversation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Println("Error fetching conversation:", err)
		return nil, err
	}

	var messages []model.Message
	err = r.database.Where("conversation_id = ?", conversation.ID).Order("created_at asc").Preload("Sender").Find(&messages).Error
	if err != nil {
		log.Println("Error fetching messages:", err)
		return nil, err
	}

	return &messages, nil
}

func (r *Repository) CreateMessage(message *model.Message) (bool, error) {
	result := r.database.Create(&message)
	if result.Error != nil {
		return false, result.Error
	}

	return true, nil
}

func (r *Repository) CreateConversation(conv *model.Conversation) (*model.Conversation, error) {
	result := r.database.Create(conv)
	if result.Error != nil {
		msg := result.Error
		return nil, msg
	}

	return conv, nil
}

func (r *Repository) CheckPrivateConversation(user1Id, user2Id uint) (*model.Conversation, error) {
	var conv model.Conversation

	result := r.database.Model(&model.Conversation{}).
		Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)", user1Id, user2Id, user2Id, user1Id).
		Where("is_group = ?", false).
		First(&conv)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			user1 := model.User{ID: user1Id}
			user2 := model.User{ID: user2Id}

			conv = model.Conversation{
				User1ID: &user1.ID,
				User2ID: &user2.ID,
			}

			return r.CreateConversation(&conv)
		}
		return nil, result.Error
	}

	return &conv, nil
}

func (r *Repository) CheckGroupConversation(groupId uint) (*model.Conversation, error) {
	var conv model.Conversation

	result := r.database.Model(&model.Conversation{}).
		Where("group_id = ?", groupId).
		Where("is_group = ?", true).
		First(&conv)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			conv = model.Conversation{
				GroupID: &groupId,
				IsGroup: true,
			}

			return r.CreateConversation(&conv)
		}
		return nil, result.Error
	}

	return &conv, nil
}

func (r *Repository) CreateGroup(payload *helpers.CreateGroupRequest, userId uint) (*model.Group, error) {
	var group model.Group

	group.CreatedBy = userId
	group.Name = payload.GroupName
	group.GroupPhotoURL = payload.GroupPhotoPath

	result := r.database.Create(&group)
	if result.Error != nil {
		msg := result.Error
		return nil, msg
	}

	return &group, nil
}

func (r *Repository) CreateGroupMembers(userId, addedById, groupId uint) (bool, error) {
	var groupMember model.GroupMember
	groupMember.GroupID = groupId
	groupMember.UserID = userId
	groupMember.AddedBy = addedById
	groupMember.AddedAt = time.Now()

	result := r.database.Create(&groupMember)
	if result.Error != nil {
		msg := result.Error
		return false, msg
	}

	return true, nil
}

func (r *Repository) GetGroupMembers(groupId uint) (*[]model.GroupMember, error) {
	var groupMember []model.GroupMember
	result := r.database.Model(&model.GroupMember{}).
		Where("group_id = ?", groupId).Preload("User").
		Find(&groupMember)
	if result.Error != nil {
		msg := result.Error
		return nil, msg
	}

	return &groupMember, nil
}

func (r *Repository) UpdateUserProfile(user *model.User) (bool, error) {
	result := r.database.Model(&user).Where("id = ?", user.ID).Update("profile_photo_url", user.ProfilePhotoURL)
	if result.Error != nil {
		msg := result.Error
		return false, msg
	}

	return true, nil
}

func (r *Repository) DeleteGroupMember(userId uint, group model.Group) (bool, error) {
	result := r.database.Model(&model.GroupMember{}).
		Where("user_id = ? AND group_id = ?", userId, group.ID).
		Delete(&model.GroupMember{})

	if result.Error != nil {
		return false, result.Error
	}

	return true, nil
}
