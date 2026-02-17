package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"gorm.io/gorm"
)

// PORepositoryInterface is the service-layer interface for the PO repository
type PORepositoryInterface interface {
	Create(po *models.PurchaseOrder) error
	GetByID(id uint) (*models.PurchaseOrder, error)
	List(params repositories.PaginationParams, status string, supplierID uint) ([]models.PurchaseOrder, int64, error)
	StatusCounts() (map[string]int64, error)
	Update(po *models.PurchaseOrder) error
	Delete(id uint) error
	ReplaceItems(poID uint, items []models.PurchaseOrderItem) error
	GetProductsForPO(supplierID uint, search string) ([]models.Product, error)
}

// StockMovementRepositoryInterface is the service-layer interface for stock movements
type StockMovementRepositoryInterface interface {
	Create(movement *models.StockMovement) error
	GetByVariant(variantID string) ([]models.StockMovement, error)
	GetByReference(referenceType string, referenceID uint) ([]models.StockMovement, error)
}

// CreatePOInput holds the input for creating a purchase order
type CreatePOInput struct {
	SupplierID uint               `json:"supplierId"`
	Date       string             `json:"date"`
	Notes      string             `json:"notes"`
	Items      []CreatePOItemInput `json:"items"`
}

// CreatePOItemInput holds the input for a single PO line item
type CreatePOItemInput struct {
	ProductID  uint    `json:"productId"`
	VariantID  string  `json:"variantId"`
	UnitID     uint    `json:"unitId"`
	OrderedQty int     `json:"orderedQty"`
	Price      float64 `json:"price"`
}

// ReceivePOInput holds the input for receiving a purchase order
type ReceivePOInput struct {
	ReceivedDate          string             `json:"receivedDate"`
	PaymentMethod         string             `json:"paymentMethod"`
	SupplierBankAccountID *string            `json:"supplierBankAccountId"`
	Items                 []ReceivePOItemInput `json:"items"`
}

// ReceivePOItemInput holds per-item input for receiving
type ReceivePOItemInput struct {
	ItemID        string  `json:"itemId"`
	ReceivedQty   int     `json:"receivedQty"`
	ReceivedPrice float64 `json:"receivedPrice"`
	IsVerified    bool    `json:"isVerified"`
}

// POService handles purchase order business logic
type POService struct {
	db        *gorm.DB
	poRepo    PORepositoryInterface
	stockRepo StockMovementRepositoryInterface
	seqSvc    *SequenceService
}

// NewPOService creates a new PO service instance
func NewPOService(db *gorm.DB, poRepo PORepositoryInterface, stockRepo StockMovementRepositoryInterface, seqSvc *SequenceService) *POService {
	return &POService{
		db:        db,
		poRepo:    poRepo,
		stockRepo: stockRepo,
		seqSvc:    seqSvc,
	}
}

// CreatePO creates a new purchase order with denormalized item fields
func (s *POService) CreatePO(input CreatePOInput) (*models.PurchaseOrder, error) {
	// Validate items exist
	if len(input.Items) == 0 {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Purchase order must have at least one item",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate supplier exists and is active
	var supplier models.Supplier
	if err := s.db.First(&supplier, input.SupplierID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{
				Err:     ErrValidation,
				Message: "Supplier not found",
				Code:    "VALIDATION_ERROR",
			}
		}
		return nil, &ServiceError{Err: err, Message: "Failed to fetch supplier", Code: "INTERNAL_ERROR"}
	}
	if !supplier.Active {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Supplier is inactive",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Generate PO number
	poNumber, err := s.seqSvc.GeneratePONumber()
	if err != nil {
		return nil, &ServiceError{Err: err, Message: "Failed to generate PO number", Code: "INTERNAL_ERROR"}
	}

	// Build items with denormalized fields
	poItems := make([]models.PurchaseOrderItem, 0, len(input.Items))
	for _, itemInput := range input.Items {
		item, err := s.buildPOItem(itemInput)
		if err != nil {
			return nil, err
		}
		poItems = append(poItems, *item)
	}

	po := &models.PurchaseOrder{
		PONumber:   poNumber,
		SupplierID: input.SupplierID,
		Date:       input.Date,
		Status:     "draft",
		Notes:      input.Notes,
		Items:      poItems,
	}

	if err := s.poRepo.Create(po); err != nil {
		return nil, &ServiceError{Err: err, Message: "Failed to create purchase order", Code: "INTERNAL_ERROR"}
	}

	return po, nil
}

// buildPOItem loads product/variant/unit data to denormalize the PO item
func (s *POService) buildPOItem(input CreatePOItemInput) (*models.PurchaseOrderItem, error) {
	// Load product
	var product models.Product
	if err := s.db.First(&product, input.ProductID).Error; err != nil {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: fmt.Sprintf("Product %d not found", input.ProductID),
			Code:    "VALIDATION_ERROR",
		}
	}

	// Load variant
	var variant models.ProductVariant
	if err := s.db.Preload("Attributes").First(&variant, "id = ?", input.VariantID).Error; err != nil {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: fmt.Sprintf("Variant %s not found", input.VariantID),
			Code:    "VALIDATION_ERROR",
		}
	}

	// Load unit
	var unit models.ProductUnit
	if err := s.db.First(&unit, input.UnitID).Error; err != nil {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: fmt.Sprintf("Unit %d not found", input.UnitID),
			Code:    "VALIDATION_ERROR",
		}
	}

	// Build variant label from attributes
	variantLabel := buildVariantLabel(variant.Attributes)

	return &models.PurchaseOrderItem{
		ProductID:    input.ProductID,
		VariantID:    input.VariantID,
		UnitID:       input.UnitID,
		UnitName:     unit.Name,
		ProductName:  product.Name,
		VariantLabel: variantLabel,
		SKU:          variant.SKU,
		CurrentStock: variant.CurrentStock,
		OrderedQty:   input.OrderedQty,
		Price:        input.Price,
	}, nil
}

// buildVariantLabel builds a human-readable variant label from attributes
func buildVariantLabel(attributes []models.VariantAttribute) string {
	if len(attributes) == 0 {
		return "Default"
	}
	parts := make([]string, 0, len(attributes))
	for _, attr := range attributes {
		parts = append(parts, attr.AttributeValue)
	}
	return strings.Join(parts, " / ")
}

// GetPO returns a single purchase order by ID
func (s *POService) GetPO(id uint) (*models.PurchaseOrder, error) {
	po, err := s.poRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{
				Err:     ErrNotFound,
				Message: "Purchase order not found",
				Code:    "PO_NOT_FOUND",
			}
		}
		return nil, &ServiceError{Err: err, Message: "Failed to fetch purchase order", Code: "INTERNAL_ERROR"}
	}
	return po, nil
}

// ListPOs returns paginated purchase orders with status counts
func (s *POService) ListPOs(params repositories.PaginationParams, status string, supplierID uint) ([]models.PurchaseOrder, int64, map[string]int64, error) {
	pos, total, err := s.poRepo.List(params, status, supplierID)
	if err != nil {
		return nil, 0, nil, &ServiceError{Err: err, Message: "Failed to list purchase orders", Code: "INTERNAL_ERROR"}
	}

	counts, err := s.poRepo.StatusCounts()
	if err != nil {
		return nil, 0, nil, &ServiceError{Err: err, Message: "Failed to get status counts", Code: "INTERNAL_ERROR"}
	}

	return pos, total, counts, nil
}

// UpdatePO updates an existing draft purchase order
func (s *POService) UpdatePO(id uint, input CreatePOInput) (*models.PurchaseOrder, error) {
	po, err := s.poRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{Err: ErrNotFound, Message: "Purchase order not found", Code: "PO_NOT_FOUND"}
		}
		return nil, &ServiceError{Err: err, Message: "Failed to fetch purchase order", Code: "INTERNAL_ERROR"}
	}

	if po.Status != "draft" {
		return nil, &ServiceError{
			Err:     ErrForbidden,
			Message: "Only draft purchase orders can be updated",
			Code:    "PO_NOT_DRAFT",
		}
	}

	po.SupplierID = input.SupplierID
	po.Date = input.Date
	po.Notes = input.Notes

	if err := s.poRepo.Update(po); err != nil {
		return nil, &ServiceError{Err: err, Message: "Failed to update purchase order", Code: "INTERNAL_ERROR"}
	}

	// Replace items if provided
	if len(input.Items) > 0 {
		poItems := make([]models.PurchaseOrderItem, 0, len(input.Items))
		for _, itemInput := range input.Items {
			item, err := s.buildPOItem(itemInput)
			if err != nil {
				return nil, err
			}
			poItems = append(poItems, *item)
		}
		if err := s.poRepo.ReplaceItems(po.ID, poItems); err != nil {
			return nil, &ServiceError{Err: err, Message: "Failed to update items", Code: "INTERNAL_ERROR"}
		}
		po.Items = poItems
	}

	return po, nil
}

// DeletePO deletes a draft purchase order
func (s *POService) DeletePO(id uint) error {
	po, err := s.poRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &ServiceError{Err: ErrNotFound, Message: "Purchase order not found", Code: "PO_NOT_FOUND"}
		}
		return &ServiceError{Err: err, Message: "Failed to fetch purchase order", Code: "INTERNAL_ERROR"}
	}

	if po.Status != "draft" {
		return &ServiceError{
			Err:     ErrForbidden,
			Message: "Only draft purchase orders can be deleted",
			Code:    "PO_NOT_DRAFT",
		}
	}

	if err := s.poRepo.Delete(id); err != nil {
		return &ServiceError{Err: err, Message: "Failed to delete purchase order", Code: "INTERNAL_ERROR"}
	}

	return nil
}

// UpdatePOStatus transitions a PO to a new status
func (s *POService) UpdatePOStatus(id uint, newStatus string) (*models.PurchaseOrder, error) {
	po, err := s.poRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{Err: ErrNotFound, Message: "Purchase order not found", Code: "PO_NOT_FOUND"}
		}
		return nil, &ServiceError{Err: err, Message: "Failed to fetch purchase order", Code: "INTERNAL_ERROR"}
	}

	if err := ValidatePOStatusTransition(po.Status, newStatus); err != nil {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: err.Error(),
			Code:    "INVALID_STATUS_TRANSITION",
		}
	}

	po.Status = newStatus
	if err := s.poRepo.Update(po); err != nil {
		return nil, &ServiceError{Err: err, Message: "Failed to update purchase order status", Code: "INTERNAL_ERROR"}
	}

	return po, nil
}

// ReceivePO processes a received PO: updates stock and creates movements
func (s *POService) ReceivePO(id uint, input ReceivePOInput) (*models.PurchaseOrder, error) {
	po, err := s.poRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{Err: ErrNotFound, Message: "Purchase order not found", Code: "PO_NOT_FOUND"}
		}
		return nil, &ServiceError{Err: err, Message: "Failed to fetch purchase order", Code: "INTERNAL_ERROR"}
	}

	// Validate status
	if po.Status != "sent" && po.Status != "draft" {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Only sent or draft purchase orders can be received",
			Code:    "PO_INVALID_STATUS",
		}
	}

	// Validate bank account required for non-cash
	if input.PaymentMethod != "cash" && (input.SupplierBankAccountID == nil || *input.SupplierBankAccountID == "") {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Supplier bank account is required for non-cash payment",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Build item lookup map
	itemMap := make(map[string]*models.PurchaseOrderItem, len(po.Items))
	for i := range po.Items {
		itemMap[po.Items[i].ID] = &po.Items[i]
	}

	// Calculate totals
	var subtotal float64
	var totalItems int

	// Parse received date
	var receivedDate *time.Time
	if input.ReceivedDate != "" {
		t, err := time.Parse("2006-01-02", input.ReceivedDate)
		if err == nil {
			receivedDate = &t
		}
	}

	// Update each item and stock
	for _, itemInput := range input.Items {
		poItem, ok := itemMap[itemInput.ItemID]
		if !ok {
			continue
		}

		qty := itemInput.ReceivedQty
		price := itemInput.ReceivedPrice
		verified := itemInput.IsVerified

		poItem.ReceivedQty = &qty
		poItem.ReceivedPrice = &price
		poItem.IsVerified = verified

		subtotal += float64(qty) * price
		totalItems += qty

		// Load unit to get toBaseUnit factor
		var unit models.ProductUnit
		if err := s.db.First(&unit, poItem.UnitID).Error; err == nil {
			stockDelta := int(float64(qty) * unit.ToBaseUnit)
			// Update variant stock
			if err := s.db.Model(&models.ProductVariant{}).
				Where("id = ?", poItem.VariantID).
				Update("current_stock", gorm.Expr("current_stock + ?", stockDelta)).Error; err != nil {
				return nil, &ServiceError{Err: err, Message: "Failed to update stock", Code: "INTERNAL_ERROR"}
			}

			// Create stock movement
			movement := &models.StockMovement{
				VariantID:     poItem.VariantID,
				MovementType:  "purchase_receive",
				Quantity:      stockDelta,
				ReferenceType: "purchase_order",
				ReferenceID:   &po.ID,
				Notes:         fmt.Sprintf("Received %d %s via PO %s", qty, unit.Name, po.PONumber),
			}
			if err := s.stockRepo.Create(movement); err != nil {
				return nil, &ServiceError{Err: err, Message: "Failed to create stock movement", Code: "INTERNAL_ERROR"}
			}
		}
	}

	// Update PO
	po.Status = "received"
	po.ReceivedDate = receivedDate
	po.PaymentMethod = &input.PaymentMethod
	po.SupplierBankAccountID = input.SupplierBankAccountID
	po.Subtotal = &subtotal
	po.TotalItems = &totalItems

	if err := s.poRepo.Update(po); err != nil {
		return nil, &ServiceError{Err: err, Message: "Failed to update purchase order", Code: "INTERNAL_ERROR"}
	}

	// Replace items with updated receive data
	if err := s.poRepo.ReplaceItems(po.ID, po.Items); err != nil {
		return nil, &ServiceError{Err: err, Message: "Failed to update items", Code: "INTERNAL_ERROR"}
	}

	return po, nil
}

// GetProductsForPO returns products eligible for a PO
func (s *POService) GetProductsForPO(supplierID uint, search string) ([]models.Product, error) {
	products, err := s.poRepo.GetProductsForPO(supplierID, search)
	if err != nil {
		return nil, &ServiceError{Err: err, Message: "Failed to fetch products", Code: "INTERNAL_ERROR"}
	}
	return products, nil
}
