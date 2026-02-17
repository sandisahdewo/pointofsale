package repositories

import (
	"time"

	"github.com/pointofsale/backend/models"
	"gorm.io/gorm"
)

// SalesRepository defines the interface for sales transaction data operations.
type SalesRepository interface {
	Create(tx *models.SalesTransaction) error
	GetByID(id uint) (*models.SalesTransaction, error)
	List(params PaginationParams, dateFrom, dateTo string, paymentMethod string) ([]models.SalesTransaction, int64, error)
}

// SalesRepositoryImpl implements SalesRepository.
type SalesRepositoryImpl struct {
	db *gorm.DB
}

// NewSalesRepository creates a new sales repository instance.
func NewSalesRepository(db *gorm.DB) *SalesRepositoryImpl {
	return &SalesRepositoryImpl{db: db}
}

// Create persists a new sales transaction with its items.
func (r *SalesRepositoryImpl) Create(tx *models.SalesTransaction) error {
	return r.db.Create(tx).Error
}

// GetByID loads a sales transaction by ID, eagerly loading its items.
func (r *SalesRepositoryImpl) GetByID(id uint) (*models.SalesTransaction, error) {
	var tx models.SalesTransaction
	err := r.db.
		Preload("Items").
		First(&tx, id).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// List returns paginated sales transactions with optional filters.
func (r *SalesRepositoryImpl) List(params PaginationParams, dateFrom, dateTo string, paymentMethod string) ([]models.SalesTransaction, int64, error) {
	var transactions []models.SalesTransaction
	var total int64

	query := r.db.Model(&models.SalesTransaction{})

	// Search by transaction number
	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		query = query.Where("transaction_number ILIKE ?", searchPattern)
	}

	// Filter by date range
	if dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			query = query.Where("date >= ?", t)
		}
	}
	if dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			// Include the entire end day
			query = query.Where("date < ?", t.AddDate(0, 0, 1))
		}
	}

	// Filter by payment method
	if paymentMethod != "" {
		query = query.Where("payment_method = ?", paymentMethod)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sort (default: date desc)
	sortBy := "date"
	switch params.SortBy {
	case "transaction_number":
		sortBy = "transaction_number"
	case "grand_total":
		sortBy = "grand_total"
	default:
		sortBy = "date"
	}

	sortDir := "desc"
	if params.SortDir == "asc" {
		sortDir = "asc"
	}

	// Apply pagination
	offset := (params.Page - 1) * params.PageSize
	if err := query.
		Order(sortBy + " " + sortDir).
		Offset(offset).
		Limit(params.PageSize).
		Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}
