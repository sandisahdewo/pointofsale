package models

import "time"

type PurchaseOrder struct {
	ID                    uint                `json:"id" gorm:"primaryKey"`
	PONumber              string              `json:"poNumber" gorm:"column:po_number;uniqueIndex"`
	SupplierID            uint                `json:"supplierId" gorm:"column:supplier_id"`
	Supplier              *Supplier           `json:"supplier,omitempty" gorm:"foreignKey:SupplierID"`
	Date                  string              `json:"date" gorm:"type:date"`
	Status                string              `json:"status" gorm:"default:draft"`
	Notes                 string              `json:"notes,omitempty"`
	ReceivedDate          *time.Time          `json:"receivedDate,omitempty" gorm:"column:received_date"`
	PaymentMethod         *string             `json:"paymentMethod,omitempty" gorm:"column:payment_method"`
	SupplierBankAccountID *string             `json:"supplierBankAccountId,omitempty" gorm:"column:supplier_bank_account_id;type:uuid"`
	Subtotal              *float64            `json:"subtotal,omitempty"`
	TotalItems            *int                `json:"totalItems,omitempty" gorm:"column:total_items"`
	Items                 []PurchaseOrderItem `json:"items,omitempty" gorm:"foreignKey:PurchaseOrderID"`
	CreatedAt             time.Time           `json:"createdAt"`
	UpdatedAt             time.Time           `json:"updatedAt"`
}

type PurchaseOrderItem struct {
	ID              string   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PurchaseOrderID uint     `json:"purchaseOrderId" gorm:"column:purchase_order_id"`
	ProductID       uint     `json:"productId" gorm:"column:product_id"`
	VariantID       string   `json:"variantId" gorm:"column:variant_id;type:uuid"`
	UnitID          uint     `json:"unitId" gorm:"column:unit_id"`
	UnitName        string   `json:"unitName" gorm:"column:unit_name"`
	ProductName     string   `json:"productName" gorm:"column:product_name"`
	VariantLabel    string   `json:"variantLabel" gorm:"column:variant_label"`
	SKU             string   `json:"sku,omitempty"`
	CurrentStock    int      `json:"currentStock" gorm:"column:current_stock;default:0"`
	OrderedQty      int      `json:"orderedQty" gorm:"column:ordered_qty"`
	Price           float64  `json:"price" gorm:"default:0"`
	ReceivedQty     *int     `json:"receivedQty,omitempty" gorm:"column:received_qty"`
	ReceivedPrice   *float64 `json:"receivedPrice,omitempty" gorm:"column:received_price"`
	IsVerified      bool     `json:"isVerified" gorm:"column:is_verified;default:false"`
}
