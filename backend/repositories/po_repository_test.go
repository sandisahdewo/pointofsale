package repositories

import (
	"testing"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePO_WithItems_CreatesAll(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewPORepository(db)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	po := &models.PurchaseOrder{
		SupplierID: supplier.ID,
		Date:       "2026-01-15",
		Status:     "draft",
		Notes:      "Test PO",
		Items: []models.PurchaseOrderItem{
			{
				ProductID:    product.ID,
				VariantID:    variant.ID,
				UnitID:       unit.ID,
				UnitName:     unit.Name,
				ProductName:  product.Name,
				VariantLabel: "Default",
				SKU:          variant.SKU,
				CurrentStock: variant.CurrentStock,
				OrderedQty:   10,
				Price:        15000,
			},
		},
	}

	err := repo.Create(po)
	require.NoError(t, err)
	assert.NotZero(t, po.ID)
	assert.NotEmpty(t, po.Items[0].ID)
}

func TestGetPO_EagerLoadsItems(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewPORepository(db)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	po := &models.PurchaseOrder{
		SupplierID: supplier.ID,
		Date:       "2026-01-15",
		Status:     "draft",
		Items: []models.PurchaseOrderItem{
			{
				ProductID:    product.ID,
				VariantID:    variant.ID,
				UnitID:       unit.ID,
				UnitName:     unit.Name,
				ProductName:  product.Name,
				VariantLabel: "Default",
				OrderedQty:   5,
				Price:        10000,
			},
		},
	}
	require.NoError(t, repo.Create(po))

	loaded, err := repo.GetByID(po.ID)
	require.NoError(t, err)
	assert.Equal(t, po.ID, loaded.ID)
	assert.Equal(t, supplier.ID, loaded.SupplierID)
	assert.NotNil(t, loaded.Supplier)
	assert.Equal(t, supplier.Name, loaded.Supplier.Name)
	assert.Len(t, loaded.Items, 1)
	assert.Equal(t, 5, loaded.Items[0].OrderedQty)
}

func TestListPOs_FilterByStatus_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewPORepository(db)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	item := models.PurchaseOrderItem{
		ProductID:    product.ID,
		VariantID:    variant.ID,
		UnitID:       unit.ID,
		UnitName:     unit.Name,
		ProductName:  product.Name,
		VariantLabel: "Default",
		OrderedQty:   1,
		Price:        100,
	}

	draftPO := &models.PurchaseOrder{PONumber: "PO-2026-0001", SupplierID: supplier.ID, Date: "2026-01-15", Status: "draft", Items: []models.PurchaseOrderItem{item}}
	sentPO := &models.PurchaseOrder{PONumber: "PO-2026-0002", SupplierID: supplier.ID, Date: "2026-01-16", Status: "sent", Items: []models.PurchaseOrderItem{item}}

	require.NoError(t, repo.Create(draftPO))
	require.NoError(t, repo.Create(sentPO))

	params := PaginationParams{Page: 1, PageSize: 10, SortBy: "date", SortDir: "asc"}

	drafts, total, err := repo.List(params, "draft", 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "draft", drafts[0].Status)
}

func TestListPOs_FilterBySupplier_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewPORepository(db)

	supplier1 := testutil.CreateTestSupplier(t, db)
	supplier2 := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	item := models.PurchaseOrderItem{
		ProductID:    product.ID,
		VariantID:    variant.ID,
		UnitID:       unit.ID,
		UnitName:     unit.Name,
		ProductName:  product.Name,
		VariantLabel: "Default",
		OrderedQty:   1,
		Price:        100,
	}

	po1 := &models.PurchaseOrder{PONumber: "PO-2026-0011", SupplierID: supplier1.ID, Date: "2026-01-15", Status: "draft", Items: []models.PurchaseOrderItem{item}}
	po2 := &models.PurchaseOrder{PONumber: "PO-2026-0012", SupplierID: supplier2.ID, Date: "2026-01-16", Status: "draft", Items: []models.PurchaseOrderItem{item}}

	require.NoError(t, repo.Create(po1))
	require.NoError(t, repo.Create(po2))

	params := PaginationParams{Page: 1, PageSize: 10, SortBy: "date", SortDir: "asc"}

	results, total, err := repo.List(params, "", supplier1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, supplier1.ID, results[0].SupplierID)
}

func TestListPOs_SearchByPONumberOrSupplier_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewPORepository(db)

	supplier := testutil.CreateTestSupplier(t, db, func(s *models.Supplier) {
		s.Name = "ACME Corp"
	})
	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	item := models.PurchaseOrderItem{
		ProductID:    product.ID,
		VariantID:    variant.ID,
		UnitID:       unit.ID,
		UnitName:     unit.Name,
		ProductName:  product.Name,
		VariantLabel: "Default",
		OrderedQty:   1,
		Price:        100,
	}

	po1 := &models.PurchaseOrder{PONumber: "PO-2026-0021", SupplierID: supplier.ID, Date: "2026-01-15", Status: "draft", Items: []models.PurchaseOrderItem{item}}
	po2 := &models.PurchaseOrder{PONumber: "PO-2026-0022", SupplierID: supplier.ID, Date: "2026-01-16", Status: "draft", Items: []models.PurchaseOrderItem{item}}
	require.NoError(t, repo.Create(po1))
	require.NoError(t, repo.Create(po2))

	// Search by PO number
	params := PaginationParams{Page: 1, PageSize: 10, Search: "0021", SortBy: "date", SortDir: "asc"}
	results, total, err := repo.List(params, "", 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "PO-2026-0021", results[0].PONumber)

	// Search by supplier name
	params2 := PaginationParams{Page: 1, PageSize: 10, Search: "ACME", SortBy: "date", SortDir: "asc"}
	results2, total2, err := repo.List(params2, "", 0)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total2)
	_ = results2
}

func TestListPOs_StatusCounts_ReturnsCorrectCounts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewPORepository(db)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	item := models.PurchaseOrderItem{
		ProductID:    product.ID,
		VariantID:    variant.ID,
		UnitID:       unit.ID,
		UnitName:     unit.Name,
		ProductName:  product.Name,
		VariantLabel: "Default",
		OrderedQty:   1,
		Price:        100,
	}

	drafts := []*models.PurchaseOrder{
		{PONumber: "PO-2026-0031", SupplierID: supplier.ID, Date: "2026-01-15", Status: "draft", Items: []models.PurchaseOrderItem{item}},
		{PONumber: "PO-2026-0032", SupplierID: supplier.ID, Date: "2026-01-16", Status: "draft", Items: []models.PurchaseOrderItem{item}},
	}
	for _, po := range drafts {
		require.NoError(t, repo.Create(po))
	}

	sent := &models.PurchaseOrder{PONumber: "PO-2026-0033", SupplierID: supplier.ID, Date: "2026-01-17", Status: "sent", Items: []models.PurchaseOrderItem{item}}
	require.NoError(t, repo.Create(sent))

	counts, err := repo.StatusCounts()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, counts["draft"], int64(2))
	assert.GreaterOrEqual(t, counts["sent"], int64(1))
	assert.GreaterOrEqual(t, counts["all"], int64(3))
}

func TestUpdatePO_DraftOnly_Succeeds(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewPORepository(db)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	po := &models.PurchaseOrder{
		PONumber:   "PO-2026-0041",
		SupplierID: supplier.ID,
		Date:       "2026-01-15",
		Status:     "draft",
		Notes:      "Original notes",
		Items: []models.PurchaseOrderItem{
			{
				ProductID:    product.ID,
				VariantID:    variant.ID,
				UnitID:       unit.ID,
				UnitName:     unit.Name,
				ProductName:  product.Name,
				VariantLabel: "Default",
				OrderedQty:   5,
				Price:        10000,
			},
		},
	}
	require.NoError(t, repo.Create(po))

	po.Notes = "Updated notes"
	err := repo.Update(po)
	require.NoError(t, err)

	loaded, err := repo.GetByID(po.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated notes", loaded.Notes)
}

func TestDeletePO_DraftOnly_Succeeds(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewPORepository(db)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	po := &models.PurchaseOrder{
		PONumber:   "PO-2026-0051",
		SupplierID: supplier.ID,
		Date:       "2026-01-15",
		Status:     "draft",
		Items: []models.PurchaseOrderItem{
			{
				ProductID:    product.ID,
				VariantID:    variant.ID,
				UnitID:       unit.ID,
				UnitName:     unit.Name,
				ProductName:  product.Name,
				VariantLabel: "Default",
				OrderedQty:   5,
				Price:        10000,
			},
		},
	}
	require.NoError(t, repo.Create(po))

	err := repo.Delete(po.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(po.ID)
	assert.Error(t, err)
}
