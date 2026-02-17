package services

// CreateProductInput represents the payload for creating a product.
type CreateProductInput struct {
	Name         string                      `json:"name"`
	Description  string                      `json:"description"`
	CategoryID   uint                        `json:"categoryId"`
	PriceSetting string                      `json:"priceSetting"`
	MarkupType   *string                     `json:"markupType"`
	HasVariants  bool                        `json:"hasVariants"`
	Status       string                      `json:"status"`
	SupplierIDs  []uint                      `json:"supplierIds"`
	Images       []CreateProductImageInput   `json:"images"`
	Units        []CreateProductUnitInput    `json:"units"`
	Variants     []CreateProductVariantInput `json:"variants"`
}

// UpdateProductInput reuses create input shape for full replacement updates.
type UpdateProductInput = CreateProductInput

type CreateProductImageInput struct {
	ImageURL  string `json:"imageUrl"`
	SortOrder int    `json:"sortOrder"`
}

type CreateProductUnitInput struct {
	Name             string  `json:"name"`
	ConversionFactor float64 `json:"conversionFactor"`
	ConvertsToName   string  `json:"convertsToName"`
	IsBase           bool    `json:"isBase"`
}

type CreateProductVariantInput struct {
	ID           string                          `json:"id,omitempty"`
	SKU          string                          `json:"sku"`
	Barcode      string                          `json:"barcode"`
	Attributes   []CreateVariantAttributeInput   `json:"attributes"`
	Images       []CreateVariantImageInput       `json:"images"`
	PricingTiers []CreateVariantPricingTierInput `json:"pricingTiers"`
	RackIDs      []uint                          `json:"rackIds"`
}

type CreateVariantAttributeInput struct {
	AttributeName  string `json:"attributeName"`
	AttributeValue string `json:"attributeValue"`
}

type CreateVariantImageInput struct {
	ImageURL  string `json:"imageUrl"`
	SortOrder int    `json:"sortOrder"`
}

type CreateVariantPricingTierInput struct {
	MinQty int     `json:"minQty"`
	Value  float64 `json:"value"`
}
