package models

import (
	"time"
)

type User struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	Name           string    `json:"name" gorm:"not null"`
	Email          string    `json:"email" gorm:"uniqueIndex;not null"`
	Phone          string    `json:"phone,omitempty"`
	Address        string    `json:"address,omitempty"`
	PasswordHash   string    `json:"-" gorm:"column:password_hash;not null"`
	ProfilePicture *string   `json:"profilePicture,omitempty" gorm:"column:profile_picture"`
	Status         string    `json:"status" gorm:"default:active;not null"`
	IsSuperAdmin   bool      `json:"isSuperAdmin" gorm:"column:is_super_admin;default:false"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Roles          []Role    `json:"roles,omitempty" gorm:"many2many:user_roles;"`
}
