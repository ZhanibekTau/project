package model

import "time"

type Message struct {
	ID             uint      `gorm:"primaryKey"`
	ConversationID uint      `gorm:"not null"`
	SenderID       uint      `gorm:"not null"`
	Content        string    `gorm:"type:text"`
	MessageType    string    `gorm:"type:enum('text','photo');not null"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`

	Conversation Conversation `gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE"`
	Sender       User         `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`
}

func (m *Message) TableName() string {
	return "messages"
}