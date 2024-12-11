package model

import "time"

type User struct {
	Id        int        `gorm:"primary_key" json:"id"`
	Firstname string     `json:"firstname"`
	Lastname  string     `json:"lastname"`
	Token     string     `json:"token"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func (u *User) TableName() string {
	return "users"
}
