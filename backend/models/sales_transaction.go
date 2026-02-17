package models

import "time"

type SalesTransaction struct {
	ID                uint                     `json:"id" gorm:"primaryKey"`
	TransactionNumber string                   `json:"transactionNumber" gorm:"column:transaction_number;uniqueIndex"`
	Date              time.Time                `json:"date"`
	Subtotal          float64                  `json:"subtotal"`
	GrandTotal        float64                  `json:"grandTotal" gorm:"column:grand_total"`
	TotalItems        int                      `json:"totalItems" gorm:"column:total_items"`
	PaymentMethod     string                   `json:"paymentMethod" gorm:"column:payment_method"`
	Items             []SalesTransactionItem   `json:"items,omitempty" gorm:"foreignKey:TransactionID"`
	CreatedAt         time.Time                `json:"createdAt"`
}

type SalesTransactionItem struct {
	ID            uint    `json:"id" gorm:"primaryKey"`
	TransactionID uint    `json:"transactionId" gorm:"column:transaction_id"`
	ProductID     uint    `json:"productId" gorm:"column:product_id"`
	VariantID     string  `json:"variantId" gorm:"column:variant_id;type:uuid"`
	UnitID        uint    `json:"unitId" gorm:"column:unit_id"`
	ProductName   string  `json:"productName" gorm:"column:product_name"`
	VariantLabel  string  `json:"variantLabel" gorm:"column:variant_label"`
	SKU           string  `json:"sku,omitempty"`
	UnitName      string  `json:"unitName" gorm:"column:unit_name"`
	Quantity      int     `json:"quantity"`
	BaseQty       int     `json:"baseQty" gorm:"column:base_qty"`
	UnitPrice     float64 `json:"unitPrice" gorm:"column:unit_price"`
	TotalPrice    float64 `json:"totalPrice" gorm:"column:total_price"`
}
