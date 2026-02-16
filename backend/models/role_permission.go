package models

import "github.com/lib/pq"

type RolePermission struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	RoleID       uint           `json:"roleId" gorm:"column:role_id;not null"`
	PermissionID uint           `json:"permissionId" gorm:"column:permission_id;not null"`
	Actions      pq.StringArray `json:"actions" gorm:"type:text[];not null"`
	Role         Role           `json:"-" gorm:"foreignKey:RoleID"`
	Permission   Permission     `json:"-" gorm:"foreignKey:PermissionID"`
}
