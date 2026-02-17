package services

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"gorm.io/gorm"
)

// ProductServiceRepository defines repository methods needed by ProductService.
type ProductServiceRepository interface {
	repositories.ProductRepository
}

// ProductService handles product business logic.
type ProductService struct {
	repo         ProductServiceRepository
	imageStorage ImageStorage
}

// NewProductService creates a new product service instance.
func NewProductService(repo ProductServiceRepository, imageStorage ...ImageStorage) *ProductService {
	var storage ImageStorage
	if len(imageStorage) > 0 {
		storage = imageStorage[0]
	}
	return &ProductService{repo: repo, imageStorage: storage}
}

// ListProducts returns paginated products with lightweight list payload.
func (s *ProductService) ListProducts(params repositories.ProductListParams) ([]repositories.ProductListItem, int64, *ServiceError) {
	products, total, err := s.repo.List(params)
	if err != nil {
		return nil, 0, &ServiceError{
			Err:     err,
			Message: "Failed to list products",
			Code:    "INTERNAL_ERROR",
		}
	}
	return products, total, nil
}

// GetProduct returns a full product by ID.
func (s *ProductService) GetProduct(id uint) (*models.Product, *ServiceError) {
	product, err := s.repo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{
				Err:     ErrNotFound,
				Message: "Product not found",
				Code:    "PRODUCT_NOT_FOUND",
			}
		}
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to fetch product",
			Code:    "INTERNAL_ERROR",
		}
	}
	return product, nil
}

// CreateProduct creates a product with nested units, variants, and relations.
func (s *ProductService) CreateProduct(input CreateProductInput) (*models.Product, *ServiceError) {
	if err := ValidateProductInput(input); err != nil {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: err.Error(),
			Code:    "VALIDATION_ERROR",
		}
	}

	if err := s.validateReferences(input); err != nil {
		return nil, err
	}

	if err := s.validateGlobalVariantUniqueness(input.Variants, 0); err != nil {
		return nil, err
	}

	status := normalizeStatus(input.Status)
	product := &models.Product{
		Name:         strings.TrimSpace(input.Name),
		Description:  strings.TrimSpace(input.Description),
		CategoryID:   input.CategoryID,
		PriceSetting: input.PriceSetting,
		MarkupType:   input.MarkupType,
		HasVariants:  input.HasVariants,
		Status:       status,
	}

	err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(product).Error; err != nil {
			return err
		}

		if err := syncProductSuppliers(tx, product.ID, input.SupplierIDs); err != nil {
			return err
		}

		if err := s.syncProductImages(tx, product.ID, input.Images); err != nil {
			return err
		}

		if err := recreateUnits(tx, product.ID, input.Units); err != nil {
			return err
		}

		if err := s.syncVariants(tx, product.ID, nil, input.Variants); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if serviceErr, ok := err.(*ServiceError); ok {
			return nil, serviceErr
		}
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to create product",
			Code:    "INTERNAL_ERROR",
		}
	}

	created, err := s.repo.GetByID(product.ID)
	if err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to load created product",
			Code:    "INTERNAL_ERROR",
		}
	}

	return created, nil
}

// UpdateProduct updates a product and syncs nested relations.
func (s *ProductService) UpdateProduct(id uint, input UpdateProductInput) (*models.Product, *ServiceError) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{
				Err:     ErrNotFound,
				Message: "Product not found",
				Code:    "PRODUCT_NOT_FOUND",
			}
		}
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to fetch product",
			Code:    "INTERNAL_ERROR",
		}
	}

	if err := ValidateProductInput(input); err != nil {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: err.Error(),
			Code:    "VALIDATION_ERROR",
		}
	}

	if err := s.validateReferences(input); err != nil {
		return nil, err
	}

	if err := s.validateGlobalVariantUniqueness(input.Variants, id); err != nil {
		return nil, err
	}

	unitsChanged := hasUnitChanges(existing.Units, input.Units)
	if unitsChanged {
		stockCount, err := s.repo.CountVariantsWithStock(id)
		if err != nil {
			return nil, &ServiceError{
				Err:     err,
				Message: "Failed to check existing stock",
				Code:    "INTERNAL_ERROR",
			}
		}
		if stockCount > 0 {
			return nil, &ServiceError{
				Err:     ErrConflict,
				Message: "Cannot modify units while stock exists.",
				Code:    "UNITS_LOCKED_BY_STOCK",
			}
		}
	}

	err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"name":          strings.TrimSpace(input.Name),
			"description":   strings.TrimSpace(input.Description),
			"category_id":   input.CategoryID,
			"price_setting": input.PriceSetting,
			"markup_type":   input.MarkupType,
			"has_variants":  input.HasVariants,
			"status":        normalizeStatus(input.Status),
		}

		if err := tx.Model(&models.Product{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}

		if err := syncProductSuppliers(tx, id, input.SupplierIDs); err != nil {
			return err
		}

		if err := s.syncProductImages(tx, id, input.Images); err != nil {
			return err
		}

		if unitsChanged {
			if err := tx.Where("product_id = ?", id).Delete(&models.ProductUnit{}).Error; err != nil {
				return err
			}
			if err := recreateUnits(tx, id, input.Units); err != nil {
				return err
			}
		}

		if err := s.syncVariants(tx, id, existing.Variants, input.Variants); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if serviceErr, ok := err.(*ServiceError); ok {
			return nil, serviceErr
		}
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to update product",
			Code:    "INTERNAL_ERROR",
		}
	}

	updated, err := s.repo.GetByID(id)
	if err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to load updated product",
			Code:    "INTERNAL_ERROR",
		}
	}

	return updated, nil
}

// DeleteProduct deletes a product if it has no stock and no purchase order references.
func (s *ProductService) DeleteProduct(id uint) *ServiceError {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &ServiceError{
				Err:     ErrNotFound,
				Message: "Product not found",
				Code:    "PRODUCT_NOT_FOUND",
			}
		}
		return &ServiceError{
			Err:     err,
			Message: "Failed to fetch product",
			Code:    "INTERNAL_ERROR",
		}
	}

	stockCount, err := s.repo.CountVariantsWithStock(id)
	if err != nil {
		return &ServiceError{
			Err:     err,
			Message: "Failed to check product stock",
			Code:    "INTERNAL_ERROR",
		}
	}
	if stockCount > 0 {
		return &ServiceError{
			Err:     ErrConflict,
			Message: "Cannot delete product with existing stock.",
			Code:    "PRODUCT_HAS_STOCK",
		}
	}

	poRefCount, err := s.repo.CountPurchaseOrderReferences(id)
	if err != nil {
		return &ServiceError{
			Err:     err,
			Message: "Failed to check purchase order references",
			Code:    "INTERNAL_ERROR",
		}
	}
	if poRefCount > 0 {
		return &ServiceError{
			Err:     ErrConflict,
			Message: fmt.Sprintf("Cannot delete product. It is referenced by %d purchase order(s).", poRefCount),
			Code:    "PRODUCT_IN_USE",
		}
	}

	if err := s.repo.Delete(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return &ServiceError{
				Err:     ErrNotFound,
				Message: "Product not found",
				Code:    "PRODUCT_NOT_FOUND",
			}
		}
		return &ServiceError{
			Err:     err,
			Message: "Failed to delete product",
			Code:    "INTERNAL_ERROR",
		}
	}
	return nil
}

func (s *ProductService) validateReferences(input CreateProductInput) *ServiceError {
	categoryExists, err := s.repo.CategoryExists(input.CategoryID)
	if err != nil {
		return &ServiceError{
			Err:     err,
			Message: "Failed to validate category",
			Code:    "INTERNAL_ERROR",
		}
	}
	if !categoryExists {
		return &ServiceError{
			Err:     ErrValidation,
			Message: "Invalid categoryId",
			Code:    "VALIDATION_ERROR",
		}
	}

	supplierIDs := uniqueUintSlice(input.SupplierIDs)
	if len(supplierIDs) > 0 {
		count, err := s.repo.CountActiveSuppliers(supplierIDs)
		if err != nil {
			return &ServiceError{
				Err:     err,
				Message: "Failed to validate suppliers",
				Code:    "INTERNAL_ERROR",
			}
		}
		if int(count) != len(supplierIDs) {
			return &ServiceError{
				Err:     ErrValidation,
				Message: "One or more supplierIds are invalid or inactive",
				Code:    "VALIDATION_ERROR",
			}
		}
	}

	rackIDs := collectRackIDs(input.Variants)
	if len(rackIDs) > 0 {
		count, err := s.repo.CountActiveRacks(rackIDs)
		if err != nil {
			return &ServiceError{
				Err:     err,
				Message: "Failed to validate racks",
				Code:    "INTERNAL_ERROR",
			}
		}
		if int(count) != len(rackIDs) {
			return &ServiceError{
				Err:     ErrValidation,
				Message: "One or more rackIds are invalid or inactive",
				Code:    "VALIDATION_ERROR",
			}
		}
	}

	return nil
}

func (s *ProductService) validateGlobalVariantUniqueness(variants []CreateProductVariantInput, excludeProductID uint) *ServiceError {
	for _, variant := range variants {
		sku := strings.TrimSpace(variant.SKU)
		if sku != "" {
			exists, err := s.repo.SKUExistsForOtherProducts(sku, excludeProductID)
			if err != nil {
				return &ServiceError{
					Err:     err,
					Message: "Failed to validate sku",
					Code:    "INTERNAL_ERROR",
				}
			}
			if exists {
				return &ServiceError{
					Err:     ErrConflict,
					Message: "SKU already exists",
					Code:    "SKU_EXISTS",
				}
			}
		}

		barcode := strings.TrimSpace(variant.Barcode)
		if barcode != "" {
			exists, err := s.repo.BarcodeExistsForOtherProducts(barcode, excludeProductID)
			if err != nil {
				return &ServiceError{
					Err:     err,
					Message: "Failed to validate barcode",
					Code:    "INTERNAL_ERROR",
				}
			}
			if exists {
				return &ServiceError{
					Err:     ErrConflict,
					Message: "Barcode already exists",
					Code:    "BARCODE_EXISTS",
				}
			}
		}
	}

	return nil
}

func normalizeStatus(status string) string {
	trimmed := strings.TrimSpace(status)
	if trimmed == "" {
		return "active"
	}
	return trimmed
}

func syncProductSuppliers(tx *gorm.DB, productID uint, supplierIDs []uint) error {
	product := models.Product{ID: productID}
	ids := uniqueUintSlice(supplierIDs)

	if len(ids) == 0 {
		return tx.Model(&product).Association("Suppliers").Clear()
	}

	var suppliers []models.Supplier
	if err := tx.Where("id IN ?", ids).Find(&suppliers).Error; err != nil {
		return err
	}
	return tx.Model(&product).Association("Suppliers").Replace(&suppliers)
}

func (s *ProductService) syncProductImages(tx *gorm.DB, productID uint, images []CreateProductImageInput) error {
	if err := tx.Where("product_id = ?", productID).Delete(&models.ProductImage{}).Error; err != nil {
		return err
	}

	if len(images) == 0 {
		return nil
	}

	toCreate := make([]models.ProductImage, 0, len(images))
	for i, image := range images {
		url := strings.TrimSpace(image.ImageURL)
		if url == "" {
			continue
		}
		uploadedURL, err := s.resolveImageURL(url, fmt.Sprintf("products/%d/%s", productID, uuid.NewString()))
		if err != nil {
			return fmt.Errorf("process product image at index %d: %w", i, err)
		}
		sortOrder := image.SortOrder
		if sortOrder < 0 {
			sortOrder = i
		}
		toCreate = append(toCreate, models.ProductImage{
			ProductID: productID,
			ImageURL:  uploadedURL,
			SortOrder: sortOrder,
		})
	}

	if len(toCreate) == 0 {
		return nil
	}

	return tx.Create(&toCreate).Error
}

func recreateUnits(tx *gorm.DB, productID uint, units []CreateProductUnitInput) error {
	orderedUnits, err := ResolveUnitDependencyOrder(units)
	if err != nil {
		return err
	}

	createdByName := make(map[string]models.ProductUnit, len(orderedUnits))
	for _, unit := range orderedUnits {
		name := strings.TrimSpace(unit.Name)
		key := strings.ToLower(name)

		model := models.ProductUnit{
			ProductID: productID,
			Name:      name,
			IsBase:    unit.IsBase,
		}

		if unit.IsBase {
			model.ConversionFactor = 1
			model.ToBaseUnit = 1
			model.ConvertsToID = nil
		} else {
			parentKey := strings.ToLower(strings.TrimSpace(unit.ConvertsToName))
			parent, ok := createdByName[parentKey]
			if !ok {
				return fmt.Errorf("convertsToName must reference another unit in the same request")
			}
			model.ConversionFactor = unit.ConversionFactor
			model.ToBaseUnit = unit.ConversionFactor * parent.ToBaseUnit
			model.ConvertsToID = &parent.ID
		}

		if err := tx.Create(&model).Error; err != nil {
			return err
		}

		createdByName[key] = model
	}

	return nil
}

func (s *ProductService) syncVariants(tx *gorm.DB, productID uint, existing []models.ProductVariant, inputs []CreateProductVariantInput) error {
	existingByID := make(map[string]models.ProductVariant, len(existing))
	for _, variant := range existing {
		existingByID[variant.ID] = variant
	}

	incomingIDs := make(map[string]struct{}, len(inputs))
	for _, in := range inputs {
		id := strings.TrimSpace(in.ID)
		if id != "" {
			incomingIDs[id] = struct{}{}
		}
	}

	// Delete removed variants, but block deletion when stock exists.
	for _, variant := range existing {
		if _, keep := incomingIDs[variant.ID]; keep {
			continue
		}
		if variant.CurrentStock > 0 {
			return &ServiceError{
				Err:     ErrConflict,
				Message: "Cannot delete variant with existing stock.",
				Code:    "VARIANT_HAS_STOCK",
			}
		}
		if err := tx.Delete(&models.ProductVariant{}, "id = ?", variant.ID).Error; err != nil {
			return err
		}
	}

	// Upsert variants and nested data.
	for _, in := range inputs {
		trimmedID := strings.TrimSpace(in.ID)
		if existingVariant, ok := existingByID[trimmedID]; ok {
			updates := map[string]interface{}{
				"sku":     strings.TrimSpace(in.SKU),
				"barcode": strings.TrimSpace(in.Barcode),
			}
			if err := tx.Model(&models.ProductVariant{}).Where("id = ?", existingVariant.ID).Updates(updates).Error; err != nil {
				return err
			}
			if err := s.replaceVariantDetails(tx, productID, existingVariant.ID, in); err != nil {
				return err
			}
			continue
		}

		newVariant := models.ProductVariant{
			ProductID: productID,
			SKU:       strings.TrimSpace(in.SKU),
			Barcode:   strings.TrimSpace(in.Barcode),
		}
		if trimmedID != "" {
			if _, err := uuid.Parse(trimmedID); err == nil {
				newVariant.ID = trimmedID
			}
		}
		if err := tx.Create(&newVariant).Error; err != nil {
			return err
		}
		if err := s.replaceVariantDetails(tx, productID, newVariant.ID, in); err != nil {
			return err
		}
	}

	return nil
}

func (s *ProductService) replaceVariantDetails(tx *gorm.DB, productID uint, variantID string, input CreateProductVariantInput) error {
	if err := tx.Where("variant_id = ?", variantID).Delete(&models.VariantAttribute{}).Error; err != nil {
		return err
	}
	if err := tx.Where("variant_id = ?", variantID).Delete(&models.VariantImage{}).Error; err != nil {
		return err
	}
	if err := tx.Where("variant_id = ?", variantID).Delete(&models.VariantPricingTier{}).Error; err != nil {
		return err
	}

	if len(input.Attributes) > 0 {
		attributes := make([]models.VariantAttribute, 0, len(input.Attributes))
		for _, attr := range input.Attributes {
			name := strings.TrimSpace(attr.AttributeName)
			value := strings.TrimSpace(attr.AttributeValue)
			if name == "" || value == "" {
				continue
			}
			attributes = append(attributes, models.VariantAttribute{
				VariantID:      variantID,
				AttributeName:  name,
				AttributeValue: value,
			})
		}
		if len(attributes) > 0 {
			if err := tx.Create(&attributes).Error; err != nil {
				return err
			}
		}
	}

	if len(input.Images) > 0 {
		images := make([]models.VariantImage, 0, len(input.Images))
		for i, image := range input.Images {
			url := strings.TrimSpace(image.ImageURL)
			if url == "" {
				continue
			}
			uploadedURL, err := s.resolveImageURL(url, fmt.Sprintf("products/%d/variants/%s/%s", productID, variantID, uuid.NewString()))
			if err != nil {
				return fmt.Errorf("process variant image at index %d: %w", i, err)
			}
			sortOrder := image.SortOrder
			if sortOrder < 0 {
				sortOrder = i
			}
			images = append(images, models.VariantImage{
				VariantID: variantID,
				ImageURL:  uploadedURL,
				SortOrder: sortOrder,
			})
		}
		if len(images) > 0 {
			if err := tx.Create(&images).Error; err != nil {
				return err
			}
		}
	}

	if len(input.PricingTiers) > 0 {
		pricing := make([]models.VariantPricingTier, 0, len(input.PricingTiers))
		for _, tier := range input.PricingTiers {
			pricing = append(pricing, models.VariantPricingTier{
				VariantID: variantID,
				MinQty:    tier.MinQty,
				Value:     tier.Value,
			})
		}
		if err := tx.Create(&pricing).Error; err != nil {
			return err
		}
	}

	variant := models.ProductVariant{ID: variantID}
	rackIDs := uniqueUintSlice(input.RackIDs)
	if len(rackIDs) == 0 {
		return tx.Model(&variant).Association("Racks").Clear()
	}

	var racks []models.Rack
	if err := tx.Where("id IN ?", rackIDs).Find(&racks).Error; err != nil {
		return err
	}
	return tx.Model(&variant).Association("Racks").Replace(&racks)
}

func uniqueUintSlice(ids []uint) []uint {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func collectRackIDs(variants []CreateProductVariantInput) []uint {
	ids := make([]uint, 0)
	for _, variant := range variants {
		ids = append(ids, variant.RackIDs...)
	}
	return uniqueUintSlice(ids)
}

type normalizedUnit struct {
	Name       string
	Conversion float64
	ConvertsTo string
	IsBase     bool
}

func hasUnitChanges(existing []models.ProductUnit, requested []CreateProductUnitInput) bool {
	if len(existing) != len(requested) {
		return true
	}

	existingNorm := normalizeExistingUnits(existing)
	requestNorm := normalizeRequestUnits(requested)

	if len(existingNorm) != len(requestNorm) {
		return true
	}

	sort.Slice(existingNorm, func(i, j int) bool { return existingNorm[i].Name < existingNorm[j].Name })
	sort.Slice(requestNorm, func(i, j int) bool { return requestNorm[i].Name < requestNorm[j].Name })

	for i := range existingNorm {
		a := existingNorm[i]
		b := requestNorm[i]
		if a.Name != b.Name || a.IsBase != b.IsBase || a.ConvertsTo != b.ConvertsTo {
			return true
		}
		if math.Abs(a.Conversion-b.Conversion) > 0.0001 {
			return true
		}
	}

	return false
}

func normalizeExistingUnits(units []models.ProductUnit) []normalizedUnit {
	idToName := make(map[uint]string, len(units))
	for _, unit := range units {
		idToName[unit.ID] = strings.ToLower(strings.TrimSpace(unit.Name))
	}

	normalized := make([]normalizedUnit, 0, len(units))
	for _, unit := range units {
		convertsTo := ""
		if unit.ConvertsToID != nil {
			convertsTo = idToName[*unit.ConvertsToID]
		}
		normalized = append(normalized, normalizedUnit{
			Name:       strings.ToLower(strings.TrimSpace(unit.Name)),
			Conversion: unit.ConversionFactor,
			ConvertsTo: convertsTo,
			IsBase:     unit.IsBase,
		})
	}
	return normalized
}

func normalizeRequestUnits(units []CreateProductUnitInput) []normalizedUnit {
	normalized := make([]normalizedUnit, 0, len(units))
	for _, unit := range units {
		conversion := unit.ConversionFactor
		if unit.IsBase {
			conversion = 1
		}
		normalized = append(normalized, normalizedUnit{
			Name:       strings.ToLower(strings.TrimSpace(unit.Name)),
			Conversion: conversion,
			ConvertsTo: strings.ToLower(strings.TrimSpace(unit.ConvertsToName)),
			IsBase:     unit.IsBase,
		})
	}
	return normalized
}
