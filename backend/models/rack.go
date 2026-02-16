package models

import "time"

type Rack struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Code        string    `json:"code" gorm:"uniqueIndex"`
	Location    string    `json:"location"`
	Capacity    int       `json:"capacity"`
	Description string    `json:"description,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
