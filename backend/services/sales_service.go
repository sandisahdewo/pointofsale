package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SalesRepositoryInterface defines repository methods needed by SalesService.
type SalesRepositoryInterface interface {
	Create(tx *models.SalesTransaction) error
	GetByID(id uint) (*models.SalesTransaction, error)
	List(params repositories.PaginationParams, dateFrom, dateTo string, paymentMethod string) ([]models.SalesTransaction, int64, error)
}

// CheckoutInput is the input for creating a sales transaction.
type CheckoutInput struct {
	PaymentMethod string              `json:"paymentMethod"`
	Items         []CheckoutItemInput `json:"items"`
}

// CheckoutItemInput represents a single line item in the checkout.
type CheckoutItemInput struct {
	ProductID uint   `json:"productId"`
	VariantID string `json:"variantId"`
	UnitID    uint   `json:"unitId"`
	Quantity  int    `json:"quantity"`
}

// ProductSearchResult is the DTO returned by ProductSearch.
type ProductSearchResult struct {
	ID           uint                    `json:"id"`
	Name         string                  `json:"name"`
	Description  string                  `json:"description"`
	HasVariants  bool                    `json:"hasVariants"`
	PriceSetting string                  `json:"priceSetting"`
	MarkupType   *string                 `json:"markupType"`
	Images       []ProductImageResult    `json:"images"`
	Units        []ProductUnitResult     `json:"units"`
	Variants     []ProductVariantResult  `json:"variants"`
}

// ProductImageResult is a simplified product image DTO.
type ProductImageResult struct {
	ID        uint   `json:"id"`
	ImageURL  string `json:"imageUrl"`
	SortOrder int    `json:"sortOrder"`
}

// ProductUnitResult is a simplified unit DTO.
type ProductUnitResult struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	IsBase     bool    `json:"isBase"`
	ToBaseUnit float64 `json:"toBaseUnit"`
}

// ProductVariantResult is a simplified variant DTO for search results.
type ProductVariantResult struct {
	ID           string                    `json:"id"`
	SKU          string                    `json:"sku"`
	Barcode      string                    `json:"barcode"`
	CurrentStock int                       `json:"currentStock"`
	Attributes   []VariantAttributeResult  `json:"attributes"`
	Images       []VariantImageResult      `json:"images"`
	PricingTiers []VariantPricingTierResult `json:"pricingTiers"`
}

// VariantAttributeResult is a simplified attribute DTO.
type VariantAttributeResult struct {
	AttributeName  string `json:"attributeName"`
	AttributeValue string `json:"attributeValue"`
}

// VariantImageResult is a simplified variant image DTO.
type VariantImageResult struct {
	ImageURL  string `json:"imageUrl"`
	SortOrder int    `json:"sortOrder"`
}

// VariantPricingTierResult is a simplified pricing tier DTO.
type VariantPricingTierResult struct {
	MinQty int     `json:"minQty"`
	Value  float64 `json:"value"`
}

// SalesService handles sales transaction business logic.
type SalesService struct {
	db        *gorm.DB
	salesRepo SalesRepositoryInterface
	seqSvc    *SequenceService
}

// NewSalesService creates a new sales service instance.
func NewSalesService(db *gorm.DB, salesRepo SalesRepositoryInterface, seqSvc *SequenceService) *SalesService {
	return &SalesService{
		db:        db,
		salesRepo: salesRepo,
		seqSvc:    seqSvc,
	}
}

// validPaymentMethods is the allowlist for payment methods.
var validPaymentMethods = map[string]bool{
	"cash": true,
	"card": true,
	"qris": true,
}

// ProductSearch searches active products by name, SKU, or barcode.
// Returns at most 10 results. Query must be at least 3 characters.
func (s *SalesService) ProductSearch(query string) ([]ProductSearchResult, error) {
	query = strings.TrimSpace(query)
	if len(query) < 3 {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Search query must be at least 3 characters",
			Code:    "VALIDATION_ERROR",
		}
	}

	searchPattern := "%" + query + "%"

	var products []models.Product
	err := s.db.
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Units", func(db *gorm.DB) *gorm.DB {
			return db.Order("to_base_unit ASC")
		}).
		Preload("Variants", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("Variants.Attributes").
		Preload("Variants.Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Variants.PricingTiers", func(db *gorm.DB) *gorm.DB {
			return db.Order("min_qty ASC")
		}).
		Where("status = 'active'").
		Where(
			"name ILIKE ? OR EXISTS (SELECT 1 FROM product_variants pv WHERE pv.product_id = products.id AND (pv.sku ILIKE ? OR pv.barcode ILIKE ?))",
			searchPattern, searchPattern, searchPattern,
		).
		Limit(10).
		Find(&products).Error
	if err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to search products",
			Code:    "INTERNAL_ERROR",
		}
	}

	results := make([]ProductSearchResult, 0, len(products))
	for _, p := range products {
		results = append(results, toProductSearchResult(p))
	}

	return results, nil
}

// Checkout validates and processes a sales transaction.
// It deducts stock and creates stock movements within a DB transaction.
func (s *SalesService) Checkout(input CheckoutInput) (*models.SalesTransaction, error) {
	// Validate payment method
	if !validPaymentMethods[input.PaymentMethod] {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: fmt.Sprintf("Invalid payment method: %s. Must be one of: cash, card, qris", input.PaymentMethod),
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate items non-empty
	if len(input.Items) == 0 {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Cart is empty",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate each item quantity
	for _, item := range input.Items {
		if item.Quantity <= 0 {
			return nil, &ServiceError{
				Err:     ErrValidation,
				Message: "Item quantity must be greater than zero",
				Code:    "VALIDATION_ERROR",
			}
		}
	}

	var createdTx *models.SalesTransaction

	err := s.db.Transaction(func(tx *gorm.DB) error {
		txItems := make([]models.SalesTransactionItem, 0, len(input.Items))
		var subtotal float64

		for _, itemInput := range input.Items {
			// Load variant with pessimistic lock
			var variant models.ProductVariant
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ?", itemInput.VariantID).
				First(&variant).Error; err != nil {
				return &ServiceError{
					Err:     ErrValidation,
					Message: fmt.Sprintf("Variant %s not found", itemInput.VariantID),
					Code:    "VARIANT_NOT_FOUND",
				}
			}

			// Load pricing tiers
			var pricingTiers []models.VariantPricingTier
			if err := tx.Where("variant_id = ?", variant.ID).Find(&pricingTiers).Error; err != nil {
				return err
			}

			// Load unit
			var unit models.ProductUnit
			if err := tx.Where("id = ?", itemInput.UnitID).First(&unit).Error; err != nil {
				return &ServiceError{
					Err:     ErrValidation,
					Message: fmt.Sprintf("Unit %d not found", itemInput.UnitID),
					Code:    "UNIT_NOT_FOUND",
				}
			}

			// Load product for name/denormalization
			var product models.Product
			if err := tx.First(&product, itemInput.ProductID).Error; err != nil {
				return &ServiceError{
					Err:     ErrValidation,
					Message: fmt.Sprintf("Product %d not found", itemInput.ProductID),
					Code:    "PRODUCT_NOT_FOUND",
				}
			}

			// Calculate base quantity
			baseQty := itemInput.Quantity * int(unit.ToBaseUnit)

			// Stock check
			if baseQty > variant.CurrentStock {
				return &ServiceError{
					Err:     ErrValidation,
					Message: fmt.Sprintf("Insufficient stock for %s. Available: %d, requested: %d (base units)", product.Name, variant.CurrentStock, baseQty),
					Code:    "INSUFFICIENT_STOCK",
				}
			}

			// Calculate tiered price
			tiers := make([]PricingTier, 0, len(pricingTiers))
			for _, t := range pricingTiers {
				tiers = append(tiers, PricingTier{MinQty: t.MinQty, Value: t.Value})
			}

			tierValue, err := CalculateTieredPrice(tiers, itemInput.Quantity, int(unit.ToBaseUnit))
			if err != nil {
				return &ServiceError{
					Err:     err,
					Message: "Failed to calculate price",
					Code:    "PRICING_ERROR",
				}
			}

			// unitPrice = tier.value * toBaseUnit
			unitPrice := tierValue * unit.ToBaseUnit
			totalPrice := float64(itemInput.Quantity) * unitPrice

			// Build variant label
			var attributes []models.VariantAttribute
			if err := tx.Where("variant_id = ?", variant.ID).Find(&attributes).Error; err != nil {
				return err
			}
			variantLabel := buildSalesVariantLabel(attributes)

			txItems = append(txItems, models.SalesTransactionItem{
				ProductID:    product.ID,
				VariantID:    variant.ID,
				UnitID:       unit.ID,
				ProductName:  product.Name,
				VariantLabel: variantLabel,
				SKU:          variant.SKU,
				UnitName:     unit.Name,
				Quantity:     itemInput.Quantity,
				BaseQty:      baseQty,
				UnitPrice:    unitPrice,
				TotalPrice:   totalPrice,
			})

			subtotal += totalPrice

			// Deduct stock
			if err := tx.Model(&models.ProductVariant{}).
				Where("id = ?", variant.ID).
				Update("current_stock", gorm.Expr("current_stock - ?", baseQty)).Error; err != nil {
				return err
			}
		}

		// Generate transaction number
		trxNumber, err := s.seqSvc.GenerateTrxNumber()
		if err != nil {
			return err
		}

		// Create transaction record
		salesTx := &models.SalesTransaction{
			TransactionNumber: trxNumber,
			Date:              time.Now(),
			Subtotal:          subtotal,
			GrandTotal:        subtotal,
			TotalItems:        len(txItems),
			PaymentMethod:     input.PaymentMethod,
			Items:             txItems,
		}

		// Create the transaction
		if err := tx.Create(salesTx).Error; err != nil {
			return err
		}

		// Create stock movements
		for _, item := range salesTx.Items {
			movement := &models.StockMovement{
				VariantID:     item.VariantID,
				MovementType:  "sales",
				Quantity:      -item.BaseQty, // negative for deduction
				ReferenceType: "sales_transaction",
				ReferenceID:   &salesTx.ID,
				Notes:         fmt.Sprintf("Sales: %s", salesTx.TransactionNumber),
			}
			if err := tx.Create(movement).Error; err != nil {
				return err
			}
		}

		createdTx = salesTx
		return nil
	})

	if err != nil {
		if serviceErr, ok := err.(*ServiceError); ok {
			return nil, serviceErr
		}
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to process checkout",
			Code:    "INTERNAL_ERROR",
		}
	}

	return createdTx, nil
}

// GetTransaction retrieves a sales transaction by ID.
func (s *SalesService) GetTransaction(id uint) (*models.SalesTransaction, error) {
	tx, err := s.salesRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{
				Err:     ErrNotFound,
				Message: "Transaction not found",
				Code:    "TRANSACTION_NOT_FOUND",
			}
		}
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to fetch transaction",
			Code:    "INTERNAL_ERROR",
		}
	}
	return tx, nil
}

// ListTransactions returns paginated sales transactions.
func (s *SalesService) ListTransactions(params repositories.PaginationParams, dateFrom, dateTo string, paymentMethod string) ([]models.SalesTransaction, int64, error) {
	return s.salesRepo.List(params, dateFrom, dateTo, paymentMethod)
}

// buildSalesVariantLabel constructs a human-readable label from variant attributes.
func buildSalesVariantLabel(attributes []models.VariantAttribute) string {
	if len(attributes) == 0 {
		return "Default"
	}

	parts := make([]string, 0, len(attributes))
	for _, attr := range attributes {
		parts = append(parts, attr.AttributeValue)
	}

	return strings.Join(parts, " / ")
}

// toProductSearchResult converts a models.Product to ProductSearchResult.
func toProductSearchResult(p models.Product) ProductSearchResult {
	images := make([]ProductImageResult, 0, len(p.Images))
	for _, img := range p.Images {
		images = append(images, ProductImageResult{
			ID:        img.ID,
			ImageURL:  img.ImageURL,
			SortOrder: img.SortOrder,
		})
	}

	units := make([]ProductUnitResult, 0, len(p.Units))
	for _, u := range p.Units {
		units = append(units, ProductUnitResult{
			ID:         u.ID,
			Name:       u.Name,
			IsBase:     u.IsBase,
			ToBaseUnit: u.ToBaseUnit,
		})
	}

	variants := make([]ProductVariantResult, 0, len(p.Variants))
	for _, v := range p.Variants {
		attrs := make([]VariantAttributeResult, 0, len(v.Attributes))
		for _, a := range v.Attributes {
			attrs = append(attrs, VariantAttributeResult{
				AttributeName:  a.AttributeName,
				AttributeValue: a.AttributeValue,
			})
		}

		varImgs := make([]VariantImageResult, 0, len(v.Images))
		for _, i := range v.Images {
			varImgs = append(varImgs, VariantImageResult{
				ImageURL:  i.ImageURL,
				SortOrder: i.SortOrder,
			})
		}

		tiers := make([]VariantPricingTierResult, 0, len(v.PricingTiers))
		for _, t := range v.PricingTiers {
			tiers = append(tiers, VariantPricingTierResult{
				MinQty: t.MinQty,
				Value:  t.Value,
			})
		}

		variants = append(variants, ProductVariantResult{
			ID:           v.ID,
			SKU:          v.SKU,
			Barcode:      v.Barcode,
			CurrentStock: v.CurrentStock,
			Attributes:   attrs,
			Images:       varImgs,
			PricingTiers: tiers,
		})
	}

	return ProductSearchResult{
		ID:           p.ID,
		Name:         p.Name,
		Description:  p.Description,
		HasVariants:  p.HasVariants,
		PriceSetting: p.PriceSetting,
		MarkupType:   p.MarkupType,
		Images:       images,
		Units:        units,
		Variants:     variants,
	}
}
