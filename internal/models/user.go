package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Password string `json:"password,omitempty"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
}
