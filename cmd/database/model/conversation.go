package model

import "time"

type Conversation struct {
	ID        uint `gorm:"primaryKey"`
	IsGroup   bool `gorm:"not null"`
	GroupID   *uint
	User1ID   *uint
	User2ID   *uint
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`

	Group Group `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`
	User1 User  `gorm:"foreignKey:User1ID;constraint:OnDelete:CASCADE"`
	User2 User  `gorm:"foreignKey:User2ID;constraint:OnDelete:CASCADE"`
}

func (c *Conversation) TableName() string {
	return "conversations"
}
