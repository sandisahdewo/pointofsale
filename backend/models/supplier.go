package models

import "time"

type Supplier struct {
	ID           uint                  `json:"id" gorm:"primaryKey"`
	Name         string                `json:"name"`
	Address      string                `json:"address"`
	Phone        string                `json:"phone,omitempty"`
	Email        string                `json:"email,omitempty"`
	Website      string                `json:"website,omitempty"`
	Active       bool                  `json:"active"`
	BankAccounts []SupplierBankAccount `json:"bankAccounts" gorm:"foreignKey:SupplierID"`
	CreatedAt    time.Time             `json:"createdAt"`
	UpdatedAt    time.Time             `json:"updatedAt"`
}

type SupplierBankAccount struct {
	ID            string `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SupplierID    uint   `json:"supplierId" gorm:"column:supplier_id"`
	AccountName   string `json:"accountName" gorm:"column:account_name"`
	AccountNumber string `json:"accountNumber" gorm:"column:account_number"`
}
