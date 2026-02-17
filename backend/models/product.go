package models

import "time"

type Product struct {
	ID           uint             `json:"id" gorm:"primaryKey"`
	Name         string           `json:"name"`
	Description  string           `json:"description,omitempty"`
	CategoryID   uint             `json:"categoryId" gorm:"column:category_id"`
	Category     *Category        `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	PriceSetting string           `json:"priceSetting" gorm:"column:price_setting;default:fixed"`
	MarkupType   *string          `json:"markupType,omitempty" gorm:"column:markup_type"`
	HasVariants  bool             `json:"hasVariants" gorm:"column:has_variants;default:false"`
	Status       string           `json:"status" gorm:"default:active"`
	Images       []ProductImage   `json:"images" gorm:"foreignKey:ProductID"`
	Suppliers    []Supplier       `json:"suppliers,omitempty" gorm:"many2many:product_suppliers;"`
	Units        []ProductUnit    `json:"units,omitempty" gorm:"foreignKey:ProductID"`
	Variants     []ProductVariant `json:"variants,omitempty" gorm:"foreignKey:ProductID"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
}

type ProductImage struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	ProductID uint   `json:"productId" gorm:"column:product_id"`
	ImageURL  string `json:"imageUrl" gorm:"column:image_url"`
	SortOrder int    `json:"sortOrder" gorm:"column:sort_order;default:0"`
}

type ProductUnit struct {
	ID               uint    `json:"id" gorm:"primaryKey"`
	ProductID        uint    `json:"productId" gorm:"column:product_id"`
	Name             string  `json:"name"`
	ConversionFactor float64 `json:"conversionFactor" gorm:"column:conversion_factor;default:1"`
	ConvertsToID     *uint   `json:"convertsToId,omitempty" gorm:"column:converts_to_id"`
	ToBaseUnit       float64 `json:"toBaseUnit" gorm:"column:to_base_unit;default:1"`
	IsBase           bool    `json:"isBase" gorm:"column:is_base;default:false"`
}

type ProductVariant struct {
	ID           string               `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProductID    uint                 `json:"productId" gorm:"column:product_id"`
	SKU          string               `json:"sku,omitempty"`
	Barcode      string               `json:"barcode,omitempty"`
	CurrentStock int                  `json:"currentStock" gorm:"column:current_stock;default:0"`
	Attributes   []VariantAttribute   `json:"attributes" gorm:"foreignKey:VariantID"`
	Images       []VariantImage       `json:"images" gorm:"foreignKey:VariantID"`
	PricingTiers []VariantPricingTier `json:"pricingTiers" gorm:"foreignKey:VariantID"`
	Racks        []Rack               `json:"racks,omitempty" gorm:"many2many:variant_racks;foreignKey:ID;joinForeignKey:VariantID;references:ID;joinReferences:RackID"`
	CreatedAt    time.Time            `json:"createdAt"`
	UpdatedAt    time.Time            `json:"updatedAt"`
}

type VariantAttribute struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	VariantID      string `json:"variantId" gorm:"column:variant_id;type:uuid"`
	AttributeName  string `json:"attributeName" gorm:"column:attribute_name"`
	AttributeValue string `json:"attributeValue" gorm:"column:attribute_value"`
}

type VariantImage struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	VariantID string `json:"variantId" gorm:"column:variant_id;type:uuid"`
	ImageURL  string `json:"imageUrl" gorm:"column:image_url"`
	SortOrder int    `json:"sortOrder" gorm:"column:sort_order;default:0"`
}

type VariantPricingTier struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	VariantID string  `json:"variantId" gorm:"column:variant_id;type:uuid"`
	MinQty    int     `json:"minQty" gorm:"column:min_qty"`
	Value     float64 `json:"value"`
}
