package model

import "time"

type Reaction struct {
	ID        uint      `gorm:"primaryKey"`
	MessageID uint      `gorm:"not null"`
	UserID    uint      `gorm:"not null"`
	Reaction  string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	Message Message `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE"`
	User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (r *Reaction) TableName() string {
	return "reactions"
}
