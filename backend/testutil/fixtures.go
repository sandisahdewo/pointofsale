package testutil

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/utils"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// hashTestPassword returns a hashed version of the test password "Password@123".
func hashTestPassword(t *testing.T) string {
	t.Helper()
	hash, err := utils.HashPassword("Password@123")
	require.NoError(t, err, "failed to hash test password")
	return hash
}

// CreateTestUser creates a user in the test database with sensible defaults.
// Override fields using optional functions.
func CreateTestUser(t *testing.T, db *gorm.DB, overrides ...func(*models.User)) *models.User {
	t.Helper()

	user := &models.User{
		Name:         "Test User",
		Email:        fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8]),
		PasswordHash: hashTestPassword(t),
		Status:       "active",
		IsSuperAdmin: false,
	}

	// Apply overrides
	for _, override := range overrides {
		override(user)
	}

	err := db.Create(user).Error
	require.NoError(t, err, "failed to create test user")

	return user
}

// CreateTestRole creates a role in the test database.
func CreateTestRole(t *testing.T, db *gorm.DB, overrides ...func(*models.Role)) *models.Role {
	t.Helper()

	role := &models.Role{
		Name:        fmt.Sprintf("Test Role %s", uuid.New().String()[:8]),
		Description: "Test role description",
		IsSystem:    false,
	}

	// Apply overrides
	for _, override := range overrides {
		override(role)
	}

	err := db.Create(role).Error
	require.NoError(t, err, "failed to create test role")

	return role
}

// CreateTestPermission creates a permission in the test database.
func CreateTestPermission(t *testing.T, db *gorm.DB, overrides ...func(*models.Permission)) *models.Permission {
	t.Helper()

	permission := &models.Permission{
		Module:  "test",
		Feature: fmt.Sprintf("feature-%s", uuid.New().String()[:8]),
		Actions: []string{"view", "create", "edit", "delete"},
	}

	// Apply overrides
	for _, override := range overrides {
		override(permission)
	}

	err := db.Create(permission).Error
	require.NoError(t, err, "failed to create test permission")

	return permission
}

// CreateTestSupplier creates a supplier in the test database with sensible defaults.
func CreateTestSupplier(t *testing.T, db *gorm.DB, overrides ...func(*models.Supplier)) *models.Supplier {
	t.Helper()

	supplier := &models.Supplier{
		Name:    fmt.Sprintf("Test Supplier %s", uuid.New().String()[:8]),
		Address: "Test Address",
		Active:  true,
	}

	// Apply overrides
	for _, override := range overrides {
		override(supplier)
	}

	err := db.Create(supplier).Error
	require.NoError(t, err, "failed to create test supplier")

	return supplier
}

// CreateTestCategory creates a category in the test database with sensible defaults.
func CreateTestCategory(t *testing.T, db *gorm.DB, overrides ...func(*models.Category)) *models.Category {
	t.Helper()

	category := &models.Category{
		Name:        fmt.Sprintf("Test Category %s", uuid.New().String()[:8]),
		Description: "Test category description",
	}

	for _, override := range overrides {
		override(category)
	}

	err := db.Create(category).Error
	require.NoError(t, err, "failed to create test category")

	return category
}

// CreateTestRack creates a rack in the test database with sensible defaults.
func CreateTestRack(t *testing.T, db *gorm.DB, overrides ...func(*models.Rack)) *models.Rack {
	t.Helper()

	rack := &models.Rack{
		Name:     fmt.Sprintf("Test Rack %s", uuid.New().String()[:8]),
		Code:     fmt.Sprintf("TR-%s", uuid.New().String()[:6]),
		Location: "Test Location",
		Capacity: 100,
		Active:   true,
	}

	for _, override := range overrides {
		override(rack)
	}

	err := db.Create(rack).Error
	require.NoError(t, err, "failed to create test rack")

	return rack
}

// CreateTestProduct creates a product with a base unit, a single variant, and a pricing tier.
// Returns the product with its Units and Variants eagerly loaded.
func CreateTestProduct(t *testing.T, db *gorm.DB, overrides ...func(*models.Product)) *models.Product {
	t.Helper()

	category := CreateTestCategory(t, db)

	product := &models.Product{
		Name:         fmt.Sprintf("Test Product %s", uuid.New().String()[:8]),
		Description:  "Test product description",
		CategoryID:   category.ID,
		PriceSetting: "fixed",
		HasVariants:  false,
		Status:       "active",
	}

	for _, override := range overrides {
		override(product)
	}

	err := db.Create(product).Error
	require.NoError(t, err, "failed to create test product")

	// Create base unit
	unit := &models.ProductUnit{
		ProductID:        product.ID,
		Name:             "Pcs",
		ConversionFactor: 1,
		ToBaseUnit:       1,
		IsBase:           true,
	}
	err = db.Create(unit).Error
	require.NoError(t, err, "failed to create test product unit")

	// Create default variant
	variantID := uuid.New().String()
	variant := &models.ProductVariant{
		ID:           variantID,
		ProductID:    product.ID,
		SKU:          fmt.Sprintf("TST-%s", uuid.New().String()[:6]),
		Barcode:      fmt.Sprintf("890%s", uuid.New().String()[:10]),
		CurrentStock: 100,
	}
	err = db.Create(variant).Error
	require.NoError(t, err, "failed to create test product variant")

	// Create pricing tier
	tier := &models.VariantPricingTier{
		VariantID: variantID,
		MinQty:    1,
		Value:     10000,
	}
	err = db.Create(tier).Error
	require.NoError(t, err, "failed to create test pricing tier")

	// Reload product with associations
	var loaded models.Product
	err = db.Preload("Units").Preload("Variants").Preload("Variants.PricingTiers").Preload("Variants.Attributes").First(&loaded, product.ID).Error
	require.NoError(t, err, "failed to reload test product")

	return &loaded
}

// CreateTestProductWithUnits creates a product with multiple units (base + non-base).
// Returns the product with Units, Variants, and PricingTiers loaded.
func CreateTestProductWithUnits(t *testing.T, db *gorm.DB) *models.Product {
	t.Helper()

	category := CreateTestCategory(t, db)

	product := &models.Product{
		Name:         fmt.Sprintf("Test Product %s", uuid.New().String()[:8]),
		CategoryID:   category.ID,
		PriceSetting: "fixed",
		HasVariants:  false,
		Status:       "active",
	}
	require.NoError(t, db.Create(product).Error)

	// Create base unit (Pcs)
	baseUnit := &models.ProductUnit{
		ProductID:        product.ID,
		Name:             "Pcs",
		ConversionFactor: 1,
		ToBaseUnit:       1,
		IsBase:           true,
	}
	require.NoError(t, db.Create(baseUnit).Error)

	// Create non-base unit (Dozen = 12 Pcs)
	dozenUnit := &models.ProductUnit{
		ProductID:        product.ID,
		Name:             "Dozen",
		ConversionFactor: 12,
		ConvertsToID:     &baseUnit.ID,
		ToBaseUnit:       12,
		IsBase:           false,
	}
	require.NoError(t, db.Create(dozenUnit).Error)

	// Create variant with stock and tiered pricing
	variantID := uuid.New().String()
	variant := &models.ProductVariant{
		ID:           variantID,
		ProductID:    product.ID,
		SKU:          fmt.Sprintf("TST-%s", uuid.New().String()[:6]),
		CurrentStock: 200,
	}
	require.NoError(t, db.Create(variant).Error)

	// Pricing tiers: 1+ pcs = 75000, 12+ pcs = 70000
	require.NoError(t, db.Create(&models.VariantPricingTier{VariantID: variantID, MinQty: 1, Value: 75000}).Error)
	require.NoError(t, db.Create(&models.VariantPricingTier{VariantID: variantID, MinQty: 12, Value: 70000}).Error)

	// Reload
	var loaded models.Product
	err := db.Preload("Units").Preload("Variants").Preload("Variants.PricingTiers").Preload("Variants.Attributes").First(&loaded, product.ID).Error
	require.NoError(t, err)

	return &loaded
}

// NewStockMovement creates an in-memory StockMovement (does NOT save to DB).
func NewStockMovement(variantID string, movementType string, quantity int, referenceType string, referenceID *uint, notes string) *models.StockMovement {
	return &models.StockMovement{
		VariantID:     variantID,
		MovementType:  movementType,
		Quantity:      quantity,
		ReferenceType: referenceType,
		ReferenceID:   referenceID,
		Notes:         notes,
	}
}

// CreateTestSuperAdmin creates a super admin user with the Super Admin role.
func CreateTestSuperAdmin(t *testing.T, db *gorm.DB) *models.User {
	t.Helper()

	// Create super admin role
	superAdminRole := CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Super Admin"
		r.Description = "Super administrator with all permissions"
	})

	// Create super admin user
	superAdmin := CreateTestUser(t, db, func(u *models.User) {
		u.Name = "Super Admin"
		u.Email = fmt.Sprintf("superadmin-%s@example.com", uuid.New().String()[:8])
		u.IsSuperAdmin = true
		u.Roles = []models.Role{*superAdminRole}
	})

	return superAdmin
}
