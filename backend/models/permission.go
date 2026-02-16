package models

import "github.com/lib/pq"

type Permission struct {
	ID      uint           `json:"id" gorm:"primaryKey"`
	Module  string         `json:"module" gorm:"not null"`
	Feature string         `json:"feature" gorm:"not null"`
	Actions pq.StringArray `json:"actions" gorm:"type:text[];not null"`
}
