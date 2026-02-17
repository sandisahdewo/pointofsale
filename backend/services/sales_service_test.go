package services

import (
	"sync"
	"testing"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"github.com/pointofsale/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestCheckout_Valid_DeductsStock(t *testing.T) {
	db := testutil.SetupTestDB(t)
	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]
	initialStock := variant.CurrentStock // 100

	input := CheckoutInput{
		PaymentMethod: "cash",
		Items: []CheckoutItemInput{
			{
				ProductID: product.ID,
				VariantID: variant.ID,
				UnitID:    unit.ID,
				Quantity:  2,
			},
		},
	}

	result, err := svc.Checkout(input)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotZero(t, result.ID)

	// Verify stock deducted
	var updatedVariant models.ProductVariant
	require.NoError(t, db.First(&updatedVariant, "id = ?", variant.ID).Error)
	// baseQty = 2 * 1 (toBaseUnit=1) = 2
	assert.Equal(t, initialStock-2, updatedVariant.CurrentStock)
}

func TestCheckout_InsufficientStock_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	product := testutil.CreateTestProduct(t, db, func(p *models.Product) {
		// stock will be 100 by default; we'll manually set it to 1
	})
	variant := product.Variants[0]
	unit := product.Units[0]

	// Set stock to 1
	require.NoError(t, db.Model(&models.ProductVariant{}).Where("id = ?", variant.ID).Update("current_stock", 1).Error)

	input := CheckoutInput{
		PaymentMethod: "cash",
		Items: []CheckoutItemInput{
			{
				ProductID: product.ID,
				VariantID: variant.ID,
				UnitID:    unit.ID,
				Quantity:  5, // needs 5, only 1 available
			},
		},
	}

	_, err := svc.Checkout(input)
	require.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrValidation, serviceErr.Err)
}

func TestCheckout_MultipleItems_DeductsAll(t *testing.T) {
	db := testutil.SetupTestDB(t)
	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	product1 := testutil.CreateTestProduct(t, db)
	product2 := testutil.CreateTestProduct(t, db)
	variant1 := product1.Variants[0]
	variant2 := product2.Variants[0]
	unit1 := product1.Units[0]
	unit2 := product2.Units[0]
	init1 := variant1.CurrentStock // 100
	init2 := variant2.CurrentStock // 100

	input := CheckoutInput{
		PaymentMethod: "card",
		Items: []CheckoutItemInput{
			{ProductID: product1.ID, VariantID: variant1.ID, UnitID: unit1.ID, Quantity: 3},
			{ProductID: product2.ID, VariantID: variant2.ID, UnitID: unit2.ID, Quantity: 5},
		},
	}

	result, err := svc.Checkout(input)
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)

	var v1, v2 models.ProductVariant
	require.NoError(t, db.First(&v1, "id = ?", variant1.ID).Error)
	require.NoError(t, db.First(&v2, "id = ?", variant2.ID).Error)
	assert.Equal(t, init1-3, v1.CurrentStock)
	assert.Equal(t, init2-5, v2.CurrentStock)
}

func TestCheckout_UnitConversion_DeductsCorrectBaseQty(t *testing.T) {
	db := testutil.SetupTestDB(t)
	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	// Product with units: Pcs (base) + Dozen (12 Pcs)
	product := testutil.CreateTestProductWithUnits(t, db)
	variant := product.Variants[0]
	initialStock := variant.CurrentStock // 200

	// Find the dozen unit
	var dozenUnit models.ProductUnit
	for _, u := range product.Units {
		if u.Name == "Dozen" {
			dozenUnit = u
			break
		}
	}
	require.NotZero(t, dozenUnit.ID, "Dozen unit not found")

	input := CheckoutInput{
		PaymentMethod: "cash",
		Items: []CheckoutItemInput{
			{
				ProductID: product.ID,
				VariantID: variant.ID,
				UnitID:    dozenUnit.ID,
				Quantity:  2, // 2 dozen = 24 base units
			},
		},
	}

	result, err := svc.Checkout(input)
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, 24, result.Items[0].BaseQty)

	var updated models.ProductVariant
	require.NoError(t, db.First(&updated, "id = ?", variant.ID).Error)
	assert.Equal(t, initialStock-24, updated.CurrentStock)
}

func TestCheckout_TieredPricing_AppliesCorrectTier(t *testing.T) {
	db := testutil.SetupTestDB(t)
	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	// Product has pricing tiers: 1+ pcs = 75000, 12+ pcs = 70000
	product := testutil.CreateTestProductWithUnits(t, db)
	variant := product.Variants[0]

	// Find the base unit (Pcs, ToBaseUnit=1) by name since IsBase might not preload correctly
	var baseUnit models.ProductUnit
	for _, u := range product.Units {
		if u.Name == "Pcs" {
			baseUnit = u
			break
		}
	}
	require.NotZero(t, baseUnit.ID, "base unit (Pcs) not found")
	require.Equal(t, float64(1), baseUnit.ToBaseUnit, "base unit should have toBaseUnit=1")

	// Buy 15 pcs → 15 base units → should match 12+ tier (70000)
	input := CheckoutInput{
		PaymentMethod: "cash",
		Items: []CheckoutItemInput{
			{ProductID: product.ID, VariantID: variant.ID, UnitID: baseUnit.ID, Quantity: 15},
		},
	}

	result, err := svc.Checkout(input)
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	// unitPrice = tier.value * unit.toBaseUnit = 70000 * 1 = 70000
	assert.Equal(t, float64(70000), result.Items[0].UnitPrice)
	// totalPrice = 15 * 70000 = 1050000
	assert.Equal(t, float64(1050000), result.Items[0].TotalPrice)
}

func TestCheckout_TieredPricingWithUnitConversion_CalculatesCorrectly(t *testing.T) {
	db := testutil.SetupTestDB(t)
	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	// Product has pricing tiers: 1+ pcs = 75000, 12+ pcs = 70000
	product := testutil.CreateTestProductWithUnits(t, db)
	variant := product.Variants[0]

	// Find dozen unit (toBaseUnit=12)
	var dozenUnit models.ProductUnit
	for _, u := range product.Units {
		if u.Name == "Dozen" {
			dozenUnit = u
			break
		}
	}
	require.NotZero(t, dozenUnit.ID)

	// Buy 2 dozen = 24 base units → matches 12+ tier (70000 per base unit)
	input := CheckoutInput{
		PaymentMethod: "cash",
		Items: []CheckoutItemInput{
			{ProductID: product.ID, VariantID: variant.ID, UnitID: dozenUnit.ID, Quantity: 2},
		},
	}

	result, err := svc.Checkout(input)
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	// unitPrice = tier.value * toBaseUnit = 70000 * 12 = 840000
	assert.Equal(t, float64(840000), result.Items[0].UnitPrice)
	// totalPrice = 2 * 840000 = 1680000
	assert.Equal(t, float64(1680000), result.Items[0].TotalPrice)
}

func TestCheckout_CalculatesSubtotalAndGrandTotal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	product := testutil.CreateTestProduct(t, db) // price 10000/pcs
	variant := product.Variants[0]
	unit := product.Units[0]

	input := CheckoutInput{
		PaymentMethod: "cash",
		Items: []CheckoutItemInput{
			{ProductID: product.ID, VariantID: variant.ID, UnitID: unit.ID, Quantity: 3},
		},
	}

	result, err := svc.Checkout(input)
	require.NoError(t, err)
	// total = 3 * 10000 = 30000
	assert.Equal(t, float64(30000), result.Subtotal)
	assert.Equal(t, float64(30000), result.GrandTotal)
	assert.Equal(t, 1, result.TotalItems)
}

func TestCheckout_GeneratesTransactionNumber(t *testing.T) {
	db := testutil.SetupTestDB(t)
	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	input := CheckoutInput{
		PaymentMethod: "cash",
		Items: []CheckoutItemInput{
			{ProductID: product.ID, VariantID: variant.ID, UnitID: unit.ID, Quantity: 1},
		},
	}

	result, err := svc.Checkout(input)
	require.NoError(t, err)
	assert.NotEmpty(t, result.TransactionNumber)
	assert.Contains(t, result.TransactionNumber, "TRX-")
}

func TestCheckout_CreatesStockMovements_NegativeQuantity(t *testing.T) {
	db := testutil.SetupTestDB(t)
	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	input := CheckoutInput{
		PaymentMethod: "cash",
		Items: []CheckoutItemInput{
			{ProductID: product.ID, VariantID: variant.ID, UnitID: unit.ID, Quantity: 3},
		},
	}

	result, err := svc.Checkout(input)
	require.NoError(t, err)

	// Verify stock movement was created
	var movements []models.StockMovement
	require.NoError(t, db.Where("variant_id = ? AND reference_type = ?", variant.ID, "sales_transaction").Find(&movements).Error)
	require.Len(t, movements, 1)
	assert.Equal(t, -3, movements[0].Quantity) // negative for sales
	assert.Equal(t, "sales", movements[0].MovementType)
	assert.Equal(t, result.ID, *movements[0].ReferenceID)
}

func TestCheckout_EmptyCart_ReturnsValidation(t *testing.T) {
	db := testutil.SetupTestDB(t)
	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	input := CheckoutInput{
		PaymentMethod: "cash",
		Items:         []CheckoutItemInput{},
	}

	_, err := svc.Checkout(input)
	require.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrValidation, serviceErr.Err)
}

func TestCheckout_InvalidPaymentMethod_ReturnsValidation(t *testing.T) {
	db := testutil.SetupTestDB(t)
	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	input := CheckoutInput{
		PaymentMethod: "bitcoin", // invalid
		Items: []CheckoutItemInput{
			{ProductID: product.ID, VariantID: variant.ID, UnitID: unit.ID, Quantity: 1},
		},
	}

	_, err := svc.Checkout(input)
	require.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrValidation, serviceErr.Err)
}

func TestCheckout_ZeroQuantity_ReturnsValidation(t *testing.T) {
	db := testutil.SetupTestDB(t)
	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	input := CheckoutInput{
		PaymentMethod: "cash",
		Items: []CheckoutItemInput{
			{ProductID: product.ID, VariantID: variant.ID, UnitID: unit.ID, Quantity: 0},
		},
	}

	_, err := svc.Checkout(input)
	require.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrValidation, serviceErr.Err)
}

func TestCheckout_ConcurrentCheckout_NoOverselling(t *testing.T) {
	// This test needs a real (non-transactional) DB connection because
	// concurrent goroutines each need their own DB connection for SELECT FOR UPDATE.
	db := testutil.SetupTestDBNoTx(t)

	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	// Set stock to exactly 1
	require.NoError(t, db.Model(&models.ProductVariant{}).Where("id = ?", variant.ID).Update("current_stock", 1).Error)

	input := CheckoutInput{
		PaymentMethod: "cash",
		Items: []CheckoutItemInput{
			{ProductID: product.ID, VariantID: variant.ID, UnitID: unit.ID, Quantity: 1},
		},
	}

	// Run 2 concurrent checkouts
	var wg sync.WaitGroup
	results := make([]error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, results[idx] = svc.Checkout(input)
		}(i)
	}
	wg.Wait()

	// Exactly one should succeed, one should fail
	successCount := 0
	errorCount := 0
	for _, err := range results {
		if err == nil {
			successCount++
		} else {
			errorCount++
		}
	}
	assert.Equal(t, 1, successCount, "exactly one checkout should succeed")
	assert.Equal(t, 1, errorCount, "exactly one checkout should fail")

	// Stock should be 0
	var finalVariant models.ProductVariant
	require.NoError(t, db.First(&finalVariant, "id = ?", variant.ID).Error)
	assert.Equal(t, 0, finalVariant.CurrentStock)
}

func TestProductSearch_ReturnsResults(t *testing.T) {
	db := testutil.SetupTestDB(t)
	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	product := testutil.CreateTestProduct(t, db, func(p *models.Product) {
		p.Name = "Unique SearchTest Product"
		p.Status = "active"
	})
	_ = product

	results, err := svc.ProductSearch("SearchTest")
	require.NoError(t, err)
	require.NotEmpty(t, results)
	assert.Equal(t, "Unique SearchTest Product", results[0].Name)
}

func TestProductSearch_TooShort_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	_, err := svc.ProductSearch("ab") // less than 3 chars
	require.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrValidation, serviceErr.Err)
}

func TestProductSearch_InactiveProducts_Excluded(t *testing.T) {
	db := testutil.SetupTestDB(t)
	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	testutil.CreateTestProduct(t, db, func(p *models.Product) {
		p.Name = "InactiveProduct XYZ"
		p.Status = "inactive"
	})

	results, err := svc.ProductSearch("InactiveProduct")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestProductSearch_Max10Results(t *testing.T) {
	db := testutil.SetupTestDB(t)
	salesRepo := repositories.NewSalesRepository(db)
	seqService := NewSequenceService(db)
	svc := NewSalesService(db, salesRepo, seqService)

	// Create 12 active products with similar name
	for i := 0; i < 12; i++ {
		testutil.CreateTestProduct(t, db, func(p *models.Product) {
			p.Name = "LimitTest Product Alpha"
			p.Status = "active"
		})
	}

	results, err := svc.ProductSearch("LimitTest")
	require.NoError(t, err)
	assert.LessOrEqual(t, len(results), 10)
}
