package repositories

import (
	"strings"

	"github.com/pointofsale/backend/models"
	"gorm.io/gorm"
)

// SupplierRepository defines the interface for supplier data operations
type SupplierRepository interface {
	Create(supplier *models.Supplier) error
	FindByID(id uint) (*models.Supplier, error)
	List(params PaginationParams, active *bool) ([]models.Supplier, int64, error)
	Update(supplier *models.Supplier, bankAccounts []models.SupplierBankAccount) error
	Delete(id uint) error
	CountPurchaseOrdersBySupplierID(supplierID uint) (int64, error)
	CleanupProductSuppliers(supplierID uint) error
}

// SupplierRepositoryImpl implements SupplierRepository interface
type SupplierRepositoryImpl struct {
	db *gorm.DB
}

// NewSupplierRepository creates a new supplier repository instance
func NewSupplierRepository(db *gorm.DB) *SupplierRepositoryImpl {
	return &SupplierRepositoryImpl{db: db}
}

// Create creates a new supplier with bank accounts in a single transaction
func (r *SupplierRepositoryImpl) Create(supplier *models.Supplier) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create supplier with explicit active field to handle false values
		if err := tx.Omit("BankAccounts").Select("Name", "Address", "Phone", "Email", "Website", "Active", "CreatedAt", "UpdatedAt").Create(supplier).Error; err != nil {
			return err
		}

		// Create bank accounts
		for i := range supplier.BankAccounts {
			supplier.BankAccounts[i].SupplierID = supplier.ID
		}
		if len(supplier.BankAccounts) > 0 {
			if err := tx.Create(&supplier.BankAccounts).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// FindByID finds a supplier by ID with bank accounts eager-loaded
func (r *SupplierRepositoryImpl) FindByID(id uint) (*models.Supplier, error) {
	var supplier models.Supplier
	err := r.db.Preload("BankAccounts").First(&supplier, id).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

// List returns paginated suppliers with optional active filter and search
func (r *SupplierRepositoryImpl) List(params PaginationParams, active *bool) ([]models.Supplier, int64, error) {
	var suppliers []models.Supplier
	var total int64

	query := r.db.Model(&models.Supplier{})

	// Apply active filter
	if active != nil {
		query = query.Where("active = ?", *active)
	}

	// Apply search filter (name, address, email)
	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		query = query.Where(
			"name ILIKE ? OR address ILIKE ? OR email ILIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderClause := params.SortBy + " " + params.SortDir
	query = query.Order(orderClause)

	// Apply pagination
	offset := (params.Page - 1) * params.PageSize
	query = query.Offset(offset).Limit(params.PageSize)

	// Preload bank accounts
	query = query.Preload("BankAccounts")

	// Execute query
	if err := query.Find(&suppliers).Error; err != nil {
		return nil, 0, err
	}

	return suppliers, total, nil
}

// Update updates a supplier and syncs bank accounts (full replace strategy)
func (r *SupplierRepositoryImpl) Update(supplier *models.Supplier, bankAccounts []models.SupplierBankAccount) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update supplier fields
		if err := tx.Save(supplier).Error; err != nil {
			return err
		}

		// Delete existing bank accounts
		if err := tx.Where("supplier_id = ?", supplier.ID).Delete(&models.SupplierBankAccount{}).Error; err != nil {
			return err
		}

		// Insert new bank accounts
		for i := range bankAccounts {
			bankAccounts[i].SupplierID = supplier.ID
			bankAccounts[i].ID = "" // Let DB generate UUID
		}
		if len(bankAccounts) > 0 {
			if err := tx.Create(&bankAccounts).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// Delete removes a supplier by ID (CASCADE deletes bank accounts)
func (r *SupplierRepositoryImpl) Delete(id uint) error {
	result := r.db.Delete(&models.Supplier{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CountPurchaseOrdersBySupplierID counts purchase orders referencing a supplier.
// Returns 0 if the purchase_orders table does not exist yet.
func (r *SupplierRepositoryImpl) CountPurchaseOrdersBySupplierID(supplierID uint) (int64, error) {
	var count int64
	// Use savepoint so a missing table doesn't poison the current transaction
	r.db.SavePoint("sp_count_po")
	err := r.db.Table("purchase_orders").Where("supplier_id = ?", supplierID).Count(&count).Error
	if err != nil {
		if isTableNotExistsError(err) {
			r.db.RollbackTo("sp_count_po")
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

// CleanupProductSuppliers removes product_suppliers junction entries for a supplier.
// Returns nil if the product_suppliers table does not exist yet.
func (r *SupplierRepositoryImpl) CleanupProductSuppliers(supplierID uint) error {
	r.db.SavePoint("sp_cleanup_ps")
	err := r.db.Exec("DELETE FROM product_suppliers WHERE supplier_id = ?", supplierID).Error
	if err != nil && isTableNotExistsError(err) {
		r.db.RollbackTo("sp_cleanup_ps")
		return nil
	}
	return err
}

// isTableNotExistsError checks if the error is a "relation does not exist" PostgreSQL error
func isTableNotExistsError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "42P01") || strings.Contains(err.Error(), "does not exist")
}
