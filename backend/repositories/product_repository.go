package repositories

import (
	"strings"
	"time"

	"github.com/pointofsale/backend/models"
	"gorm.io/gorm"
)

// ProductListParams holds filter and pagination params for list endpoint.
type ProductListParams struct {
	PaginationParams
	Status     string
	CategoryID uint
	SupplierID uint
}

// ProductListItem is the lightweight product representation for list endpoint.
type ProductListItem struct {
	ID           uint                  `json:"id"`
	Name         string                `json:"name"`
	Description  string                `json:"description,omitempty"`
	CategoryID   uint                  `json:"categoryId"`
	Category     *models.Category      `json:"category,omitempty"`
	PriceSetting string                `json:"priceSetting"`
	MarkupType   *string               `json:"markupType,omitempty"`
	HasVariants  bool                  `json:"hasVariants"`
	Status       string                `json:"status"`
	Images       []models.ProductImage `json:"images"`
	Suppliers    []models.Supplier     `json:"suppliers"`
	VariantCount int64                 `json:"variantCount"`
	CreatedAt    time.Time             `json:"createdAt"`
}

// ProductRepository defines the interface for product data operations.
type ProductRepository interface {
	GetDB() *gorm.DB
	GetByID(id uint) (*models.Product, error)
	List(params ProductListParams) ([]ProductListItem, int64, error)
	CategoryExists(id uint) (bool, error)
	CountActiveSuppliers(ids []uint) (int64, error)
	CountActiveRacks(ids []uint) (int64, error)
	SKUExistsForOtherProducts(sku string, excludeProductID uint) (bool, error)
	BarcodeExistsForOtherProducts(barcode string, excludeProductID uint) (bool, error)
	CountVariantsWithStock(productID uint) (int64, error)
	CountPurchaseOrderReferences(productID uint) (int64, error)
	Delete(id uint) error
}

// ProductRepositoryImpl implements ProductRepository.
type ProductRepositoryImpl struct {
	db *gorm.DB
}

// NewProductRepository creates a new product repository instance.
func NewProductRepository(db *gorm.DB) *ProductRepositoryImpl {
	return &ProductRepositoryImpl{db: db}
}

func (r *ProductRepositoryImpl) GetDB() *gorm.DB {
	return r.db
}

func (r *ProductRepositoryImpl) CategoryExists(id uint) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Category{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ProductRepositoryImpl) CountActiveSuppliers(ids []uint) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	var count int64
	if err := r.db.Model(&models.Supplier{}).Where("id IN ? AND active = ?", ids, true).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ProductRepositoryImpl) CountActiveRacks(ids []uint) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	var count int64
	if err := r.db.Model(&models.Rack{}).Where("id IN ? AND active = ?", ids, true).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ProductRepositoryImpl) SKUExistsForOtherProducts(sku string, excludeProductID uint) (bool, error) {
	sku = strings.TrimSpace(sku)
	if sku == "" {
		return false, nil
	}

	var count int64
	query := r.db.Model(&models.ProductVariant{}).Where("LOWER(sku) = LOWER(?)", sku)
	if excludeProductID > 0 {
		query = query.Where("product_id <> ?", excludeProductID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ProductRepositoryImpl) BarcodeExistsForOtherProducts(barcode string, excludeProductID uint) (bool, error) {
	barcode = strings.TrimSpace(barcode)
	if barcode == "" {
		return false, nil
	}

	var count int64
	query := r.db.Model(&models.ProductVariant{}).Where("LOWER(barcode) = LOWER(?)", barcode)
	if excludeProductID > 0 {
		query = query.Where("product_id <> ?", excludeProductID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetByID loads the full product with all nested relations.
func (r *ProductRepositoryImpl) GetByID(id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.
		Preload("Category").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Suppliers").
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
		Preload("Variants.Racks").
		First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// List returns lightweight product rows with pagination and filters.
func (r *ProductRepositoryImpl) List(params ProductListParams) ([]ProductListItem, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.Model(&models.Product{})

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("products.name ILIKE ?", search)
	}

	if params.Status != "" {
		query = query.Where("products.status = ?", params.Status)
	}

	if params.CategoryID > 0 {
		query = query.Where("products.category_id = ?", params.CategoryID)
	}

	if params.SupplierID > 0 {
		query = query.Where(
			"EXISTS (SELECT 1 FROM product_suppliers ps WHERE ps.product_id = products.id AND ps.supplier_id = ?)",
			params.SupplierID,
		)
	}

	sortBy := "products.id"
	switch params.SortBy {
	case "name":
		sortBy = "products.name"
	case "status":
		sortBy = "products.status"
	case "category":
		sortBy = "categories.name"
		query = query.Joins("LEFT JOIN categories ON categories.id = products.category_id")
	default:
		sortBy = "products.id"
	}

	sortDir := "asc"
	if params.SortDir == "desc" {
		sortDir = "desc"
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PageSize
	if err := query.
		Preload("Category").
		Preload("Images", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).
		Preload("Suppliers").
		Order(sortBy + " " + sortDir).
		Offset(offset).
		Limit(params.PageSize).
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	if len(products) == 0 {
		return []ProductListItem{}, total, nil
	}

	productIDs := make([]uint, 0, len(products))
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}

	type countRow struct {
		ProductID uint
		Count     int64
	}
	var rows []countRow
	if err := r.db.Table("product_variants").
		Select("product_id, COUNT(*) as count").
		Where("product_id IN ?", productIDs).
		Group("product_id").
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	countMap := make(map[uint]int64, len(rows))
	for _, row := range rows {
		countMap[row.ProductID] = row.Count
	}

	items := make([]ProductListItem, 0, len(products))
	for _, product := range products {
		item := ProductListItem{
			ID:           product.ID,
			Name:         product.Name,
			Description:  product.Description,
			CategoryID:   product.CategoryID,
			Category:     product.Category,
			PriceSetting: product.PriceSetting,
			MarkupType:   product.MarkupType,
			HasVariants:  product.HasVariants,
			Status:       product.Status,
			Suppliers:    product.Suppliers,
			VariantCount: countMap[product.ID],
			CreatedAt:    product.CreatedAt,
		}
		if len(product.Images) > 0 {
			item.Images = []models.ProductImage{product.Images[0]}
		} else {
			item.Images = []models.ProductImage{}
		}
		items = append(items, item)
	}

	return items, total, nil
}

func (r *ProductRepositoryImpl) CountVariantsWithStock(productID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.ProductVariant{}).
		Where("product_id = ? AND current_stock > 0", productID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ProductRepositoryImpl) CountPurchaseOrderReferences(productID uint) (int64, error) {
	if !r.db.Migrator().HasTable("purchase_order_items") {
		return 0, nil
	}

	var count int64

	if r.db.Migrator().HasColumn("purchase_order_items", "product_id") {
		if err := r.db.Table("purchase_order_items").
			Select("COUNT(DISTINCT purchase_order_id)").
			Where("product_id = ?", productID).
			Scan(&count).Error; err != nil {
			return 0, err
		}
		return count, nil
	}

	if r.db.Migrator().HasColumn("purchase_order_items", "variant_id") {
		if err := r.db.Table("purchase_order_items poi").
			Joins("JOIN product_variants pv ON pv.id = poi.variant_id").
			Where("pv.product_id = ?", productID).
			Select("COUNT(DISTINCT poi.purchase_order_id)").
			Scan(&count).Error; err != nil {
			return 0, err
		}
		return count, nil
	}

	return 0, nil
}

func (r *ProductRepositoryImpl) Delete(id uint) error {
	result := r.db.Delete(&models.Product{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
