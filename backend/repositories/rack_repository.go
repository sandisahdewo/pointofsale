package repositories

import (
	"fmt"

	"github.com/pointofsale/backend/models"
	"gorm.io/gorm"
)

// RackRepository defines the interface for rack data operations
type RackRepository interface {
	List(page, pageSize int, search, active, sortBy, sortDir string) ([]models.Rack, int64, error)
	FindByID(id uint) (*models.Rack, error)
	FindByCode(code string) (*models.Rack, error)
	FindByCodeExcluding(code string, excludeID uint) (*models.Rack, error)
	Create(rack *models.Rack) error
	Update(rack *models.Rack) error
	Delete(id uint) error
}

// RackRepositoryImpl implements RackRepository interface
type RackRepositoryImpl struct {
	db *gorm.DB
}

// NewRackRepository creates a new rack repository instance
func NewRackRepository(db *gorm.DB) *RackRepositoryImpl {
	return &RackRepositoryImpl{db: db}
}

// List returns paginated racks with optional search and active filter
func (r *RackRepositoryImpl) List(page, pageSize int, search, active, sortBy, sortDir string) ([]models.Rack, int64, error) {
	var racks []models.Rack
	var total int64

	query := r.db.Model(&models.Rack{})

	// Apply search filter (case-insensitive, partial match on name, code, location)
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR location ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	// Apply active filter
	if active == "true" {
		query = query.Where("active = ?", true)
	} else if active == "false" {
		query = query.Where("active = ?", false)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	if sortBy == "" {
		sortBy = "id"
	}
	if sortDir == "" {
		sortDir = "asc"
	}
	orderClause := fmt.Sprintf("%s %s", sortBy, sortDir)

	// Apply pagination
	offset := (page - 1) * pageSize

	err := query.
		Order(orderClause).
		Offset(offset).
		Limit(pageSize).
		Find(&racks).Error

	if err != nil {
		return nil, 0, err
	}

	return racks, total, nil
}

// FindByID finds a rack by ID
func (r *RackRepositoryImpl) FindByID(id uint) (*models.Rack, error) {
	var rack models.Rack
	err := r.db.First(&rack, id).Error
	if err != nil {
		return nil, err
	}
	return &rack, nil
}

// FindByCode finds a rack by code (case-insensitive)
func (r *RackRepositoryImpl) FindByCode(code string) (*models.Rack, error) {
	var rack models.Rack
	err := r.db.Where("LOWER(code) = LOWER(?)", code).First(&rack).Error
	if err != nil {
		return nil, err
	}
	return &rack, nil
}

// FindByCodeExcluding finds a rack by code excluding a specific ID (for update uniqueness check)
func (r *RackRepositoryImpl) FindByCodeExcluding(code string, excludeID uint) (*models.Rack, error) {
	var rack models.Rack
	err := r.db.Where("LOWER(code) = LOWER(?) AND id != ?", code, excludeID).First(&rack).Error
	if err != nil {
		return nil, err
	}
	return &rack, nil
}

// Create creates a new rack
func (r *RackRepositoryImpl) Create(rack *models.Rack) error {
	return r.db.Create(rack).Error
}

// Update saves changes to an existing rack
func (r *RackRepositoryImpl) Update(rack *models.Rack) error {
	return r.db.Save(rack).Error
}

// Delete deletes a rack by ID
func (r *RackRepositoryImpl) Delete(id uint) error {
	result := r.db.Delete(&models.Rack{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CleanupVariantRacks removes variant_racks junction entries for a given rack ID.
// This is a no-op if the variant_racks table does not exist yet.
func (r *RackRepositoryImpl) CleanupVariantRacks(rackID uint) error {
	if r.db.Migrator().HasTable("variant_racks") {
		return r.db.Exec("DELETE FROM variant_racks WHERE rack_id = ?", rackID).Error
	}
	return nil
}
