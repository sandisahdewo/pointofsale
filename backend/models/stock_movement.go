package models

import "time"

type StockMovement struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	VariantID     string    `json:"variantId" gorm:"column:variant_id;type:uuid"`
	MovementType  string    `json:"movementType" gorm:"column:movement_type"`
	Quantity      int       `json:"quantity"`
	ReferenceType string    `json:"referenceType,omitempty" gorm:"column:reference_type"`
	ReferenceID   *uint     `json:"referenceId,omitempty" gorm:"column:reference_id"`
	Notes         string    `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}
