package repositories

import (
	"github.com/pointofsale/backend/models"
	"gorm.io/gorm"
)

// PORepository defines the interface for purchase order data operations.
type PORepository interface {
	Create(po *models.PurchaseOrder) error
	GetByID(id uint) (*models.PurchaseOrder, error)
	List(params PaginationParams, status string, supplierID uint) ([]models.PurchaseOrder, int64, error)
	StatusCounts() (map[string]int64, error)
	Update(po *models.PurchaseOrder) error
	Delete(id uint) error
	ReplaceItems(poID uint, items []models.PurchaseOrderItem) error
	GetProductsForPO(supplierID uint, search string) ([]models.Product, error)
}

// PORepositoryImpl implements PORepository.
type PORepositoryImpl struct {
	db *gorm.DB
}

// NewPORepository creates a new purchase order repository instance.
func NewPORepository(db *gorm.DB) *PORepositoryImpl {
	return &PORepositoryImpl{db: db}
}

// Create persists a new purchase order with its items.
func (r *PORepositoryImpl) Create(po *models.PurchaseOrder) error {
	return r.db.Create(po).Error
}

// GetByID loads a purchase order by ID, eagerly loading supplier and items.
func (r *PORepositoryImpl) GetByID(id uint) (*models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	err := r.db.
		Preload("Supplier").
		Preload("Items").
		First(&po, id).Error
	if err != nil {
		return nil, err
	}
	return &po, nil
}

// List returns paginated purchase orders with optional filters.
func (r *PORepositoryImpl) List(params PaginationParams, status string, supplierID uint) ([]models.PurchaseOrder, int64, error) {
	var pos []models.PurchaseOrder
	var total int64

	query := r.db.Model(&models.PurchaseOrder{})

	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		query = query.Where(
			"po_number ILIKE ? OR EXISTS (SELECT 1 FROM suppliers s WHERE s.id = purchase_orders.supplier_id AND s.name ILIKE ?)",
			searchPattern, searchPattern,
		)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if supplierID > 0 {
		query = query.Where("supplier_id = ?", supplierID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortBy := "date"
	switch params.SortBy {
	case "po_number":
		sortBy = "po_number"
	case "status":
		sortBy = "status"
	}

	sortDir := "desc"
	if params.SortDir == "asc" {
		sortDir = "asc"
	}

	offset := (params.Page - 1) * params.PageSize
	if err := query.
		Preload("Supplier").
		Order(sortBy + " " + sortDir).
		Offset(offset).
		Limit(params.PageSize).
		Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	return pos, total, nil
}

// StatusCounts returns counts per status plus "all".
func (r *PORepositoryImpl) StatusCounts() (map[string]int64, error) {
	type countRow struct {
		Status string
		Count  int64
	}

	var rows []countRow
	if err := r.db.Model(&models.PurchaseOrder{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	var all int64
	for _, row := range rows {
		counts[row.Status] = row.Count
		all += row.Count
	}
	counts["all"] = all

	return counts, nil
}

// Update saves changes to an existing purchase order.
func (r *PORepositoryImpl) Update(po *models.PurchaseOrder) error {
	return r.db.Save(po).Error
}

// Delete removes a purchase order from the database.
func (r *PORepositoryImpl) Delete(id uint) error {
	result := r.db.Delete(&models.PurchaseOrder{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ReplaceItems replaces all items for a PO atomically.
func (r *PORepositoryImpl) ReplaceItems(poID uint, items []models.PurchaseOrderItem) error {
	if err := r.db.Where("purchase_order_id = ?", poID).Delete(&models.PurchaseOrderItem{}).Error; err != nil {
		return err
	}
	for i := range items {
		items[i].PurchaseOrderID = poID
	}
	if len(items) > 0 {
		return r.db.Create(&items).Error
	}
	return nil
}

// GetProductsForPO returns active products that belong to the specified supplier
// (or have no supplier) so the PO form can show eligible items.
func (r *PORepositoryImpl) GetProductsForPO(supplierID uint, search string) ([]models.Product, error) {
	var products []models.Product

	query := r.db.Model(&models.Product{}).Where("products.status = ?", "active")

	if supplierID > 0 {
		query = query.Where(
			"EXISTS (SELECT 1 FROM product_suppliers ps WHERE ps.product_id = products.id AND ps.supplier_id = ?) "+
				"OR NOT EXISTS (SELECT 1 FROM product_suppliers ps2 WHERE ps2.product_id = products.id)",
			supplierID,
		)
	}

	if search != "" {
		query = query.Where("products.name ILIKE ?", "%"+search+"%")
	}

	err := query.
		Preload("Units").
		Preload("Variants").
		Preload("Variants.Attributes").
		Find(&products).Error
	if err != nil {
		return nil, err
	}

	return products, nil
}
