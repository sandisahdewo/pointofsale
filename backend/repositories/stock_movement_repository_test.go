package repositories

import (
	"testing"

	"github.com/pointofsale/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateStockMovement_PurchaseReceive_PositiveQty(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewStockMovementRepository(db)

	product := testutil.CreateTestProduct(t, db)
	variantID := product.Variants[0].ID

	refID := uint(1)
	movement := testutil.NewStockMovement(variantID, "purchase_receive", 50, "purchase_order", &refID, "Received 50 units")

	err := repo.Create(movement)
	require.NoError(t, err)
	assert.NotZero(t, movement.ID)
}

func TestCreateStockMovement_Sales_NegativeQty(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewStockMovementRepository(db)

	product := testutil.CreateTestProduct(t, db)
	variantID := product.Variants[0].ID

	refID := uint(2)
	movement := testutil.NewStockMovement(variantID, "sales", -10, "sales_transaction", &refID, "Sold 10 units")

	err := repo.Create(movement)
	require.NoError(t, err)
	assert.NotZero(t, movement.ID)
}

func TestGetStockMovementsByVariant_ReturnsChronological(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewStockMovementRepository(db)

	product := testutil.CreateTestProduct(t, db)
	variantID := product.Variants[0].ID

	otherProduct := testutil.CreateTestProduct(t, db)
	otherVariantID := otherProduct.Variants[0].ID

	// Create movements for target variant
	for i := 0; i < 3; i++ {
		refID := uint(i + 1)
		movement := testutil.NewStockMovement(variantID, "purchase_receive", 10*(i+1), "purchase_order", &refID, "")
		require.NoError(t, repo.Create(movement))
	}

	// Create movement for other variant (should not appear)
	otherRefID := uint(99)
	require.NoError(t, repo.Create(testutil.NewStockMovement(otherVariantID, "purchase_receive", 5, "purchase_order", &otherRefID, "")))

	movements, err := repo.GetByVariant(variantID)
	require.NoError(t, err)
	assert.Len(t, movements, 3)

	// Should be in chronological order (ascending by created_at)
	for i := 1; i < len(movements); i++ {
		assert.True(t, movements[i].CreatedAt.After(movements[i-1].CreatedAt) || movements[i].CreatedAt.Equal(movements[i-1].CreatedAt))
	}
}

func TestGetStockMovementsByReference_ReturnsMatching(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewStockMovementRepository(db)

	poID := uint(42)
	otherPoID := uint(99)

	// Create movements for target PO
	for i := 0; i < 2; i++ {
		product := testutil.CreateTestProduct(t, db)
		variantID := product.Variants[0].ID
		movement := testutil.NewStockMovement(variantID, "purchase_receive", 10, "purchase_order", &poID, "")
		require.NoError(t, repo.Create(movement))
	}

	// Create movement for other PO
	otherProduct := testutil.CreateTestProduct(t, db)
	otherVariantID := otherProduct.Variants[0].ID
	require.NoError(t, repo.Create(testutil.NewStockMovement(otherVariantID, "purchase_receive", 5, "purchase_order", &otherPoID, "")))

	movements, err := repo.GetByReference("purchase_order", poID)
	require.NoError(t, err)
	assert.Len(t, movements, 2)

	for _, m := range movements {
		assert.Equal(t, "purchase_order", m.ReferenceType)
		assert.Equal(t, poID, *m.ReferenceID)
	}
}
