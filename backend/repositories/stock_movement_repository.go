package repositories

import (
	"github.com/pointofsale/backend/models"
	"gorm.io/gorm"
)

// StockMovementRepository defines the interface for stock movement data operations
type StockMovementRepository interface {
	Create(movement *models.StockMovement) error
	GetByVariant(variantID string) ([]models.StockMovement, error)
	GetByReference(referenceType string, referenceID uint) ([]models.StockMovement, error)
}

// StockMovementRepositoryImpl implements StockMovementRepository
type StockMovementRepositoryImpl struct {
	db *gorm.DB
}

// NewStockMovementRepository creates a new stock movement repository instance
func NewStockMovementRepository(db *gorm.DB) *StockMovementRepositoryImpl {
	return &StockMovementRepositoryImpl{db: db}
}

// Create creates a new stock movement record
func (r *StockMovementRepositoryImpl) Create(movement *models.StockMovement) error {
	return r.db.Create(movement).Error
}

// GetByVariant returns all stock movements for a specific variant in chronological order
func (r *StockMovementRepositoryImpl) GetByVariant(variantID string) ([]models.StockMovement, error) {
	var movements []models.StockMovement
	err := r.db.
		Where("variant_id = ?", variantID).
		Order("created_at ASC").
		Find(&movements).Error
	if err != nil {
		return nil, err
	}
	return movements, nil
}

// GetByReference returns all stock movements for a specific reference (e.g. purchase_order)
func (r *StockMovementRepositoryImpl) GetByReference(referenceType string, referenceID uint) ([]models.StockMovement, error) {
	var movements []models.StockMovement
	err := r.db.
		Where("reference_type = ? AND reference_id = ?", referenceType, referenceID).
		Order("created_at ASC").
		Find(&movements).Error
	if err != nil {
		return nil, err
	}
	return movements, nil
}
