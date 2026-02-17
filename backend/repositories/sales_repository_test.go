package repositories

import (
	"testing"
	"time"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestCreateSalesTransaction_Valid_CreatesWithItems(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewSalesRepository(db)

	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	tx := &models.SalesTransaction{
		TransactionNumber: "TRX-2026-000001",
		Date:              time.Now(),
		Subtotal:          50000,
		GrandTotal:        50000,
		TotalItems:        2,
		PaymentMethod:     "cash",
		Items: []models.SalesTransactionItem{
			{
				ProductID:    product.ID,
				VariantID:    variant.ID,
				UnitID:       unit.ID,
				ProductName:  product.Name,
				VariantLabel: "Default",
				SKU:          variant.SKU,
				UnitName:     unit.Name,
				Quantity:     2,
				BaseQty:      2,
				UnitPrice:    25000,
				TotalPrice:   50000,
			},
		},
	}

	err := repo.Create(tx)
	require.NoError(t, err)
	assert.NotZero(t, tx.ID)
	assert.Len(t, tx.Items, 1)
	assert.NotZero(t, tx.Items[0].ID)
}

func TestGetSalesTransaction_EagerLoadsItems(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewSalesRepository(db)

	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	tx := &models.SalesTransaction{
		TransactionNumber: "TRX-2026-000002",
		Date:              time.Now(),
		Subtotal:          30000,
		GrandTotal:        30000,
		TotalItems:        1,
		PaymentMethod:     "card",
		Items: []models.SalesTransactionItem{
			{
				ProductID:    product.ID,
				VariantID:    variant.ID,
				UnitID:       unit.ID,
				ProductName:  "Test Product B",
				VariantLabel: "Default",
				UnitName:     unit.Name,
				Quantity:     3,
				BaseQty:      3,
				UnitPrice:    10000,
				TotalPrice:   30000,
			},
		},
	}

	require.NoError(t, repo.Create(tx))

	loaded, err := repo.GetByID(tx.ID)
	require.NoError(t, err)
	assert.Equal(t, tx.ID, loaded.ID)
	assert.Equal(t, "TRX-2026-000002", loaded.TransactionNumber)
	assert.Len(t, loaded.Items, 1)
	assert.Equal(t, "Test Product B", loaded.Items[0].ProductName)
}

func TestGetSalesTransaction_NotFound_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewSalesRepository(db)

	_, err := repo.GetByID(99999)
	require.Error(t, err)
}

func TestListSalesTransactions_Pagination_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewSalesRepository(db)

	// Create 3 transactions (no items needed for list tests)
	numbers := []string{"TRX-2026-PAG001", "TRX-2026-PAG002", "TRX-2026-PAG003"}
	for i, num := range numbers {
		tx := &models.SalesTransaction{
			TransactionNumber: num,
			Date:              time.Now(),
			Subtotal:          float64((i + 1) * 10000),
			GrandTotal:        float64((i + 1) * 10000),
			TotalItems:        1,
			PaymentMethod:     "cash",
		}
		require.NoError(t, repo.Create(tx))
	}

	params := PaginationParams{Page: 1, PageSize: 2, SortBy: "date", SortDir: "desc"}
	list, total, err := repo.List(params, "", "", "")
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, list, 2)
}

func TestListSalesTransactions_FilterByDate_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewSalesRepository(db)

	yesterday := time.Now().AddDate(0, 0, -1)
	today := time.Now()

	txYesterday := &models.SalesTransaction{
		TransactionNumber: "TRX-2026-YEST01",
		Date:              yesterday,
		Subtotal:          10000,
		GrandTotal:        10000,
		TotalItems:        1,
		PaymentMethod:     "cash",
	}
	txToday := &models.SalesTransaction{
		TransactionNumber: "TRX-2026-TODAY1",
		Date:              today,
		Subtotal:          20000,
		GrandTotal:        20000,
		TotalItems:        1,
		PaymentMethod:     "cash",
	}

	require.NoError(t, repo.Create(txYesterday))
	require.NoError(t, repo.Create(txToday))

	dateFrom := today.Format("2006-01-02")
	params := PaginationParams{Page: 1, PageSize: 10, SortBy: "date", SortDir: "desc"}
	list, total, err := repo.List(params, dateFrom, "", "")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
	assert.Equal(t, "TRX-2026-TODAY1", list[0].TransactionNumber)
}

func TestListSalesTransactions_FilterByPaymentMethod_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewSalesRepository(db)

	txCash := &models.SalesTransaction{
		TransactionNumber: "TRX-2026-CASH01",
		Date:              time.Now(),
		Subtotal:          10000,
		GrandTotal:        10000,
		TotalItems:        1,
		PaymentMethod:     "cash",
	}
	txCard := &models.SalesTransaction{
		TransactionNumber: "TRX-2026-CARD01",
		Date:              time.Now(),
		Subtotal:          20000,
		GrandTotal:        20000,
		TotalItems:        1,
		PaymentMethod:     "card",
	}

	require.NoError(t, repo.Create(txCash))
	require.NoError(t, repo.Create(txCard))

	params := PaginationParams{Page: 1, PageSize: 10, SortBy: "date", SortDir: "desc"}
	list, total, err := repo.List(params, "", "", "card")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
	assert.Equal(t, "card", list[0].PaymentMethod)
}

func TestListSalesTransactions_SearchByNumber_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewSalesRepository(db)

	tx1 := &models.SalesTransaction{
		TransactionNumber: "TRX-2026-SRCH01",
		Date:              time.Now(),
		Subtotal:          10000,
		GrandTotal:        10000,
		TotalItems:        1,
		PaymentMethod:     "cash",
	}
	tx2 := &models.SalesTransaction{
		TransactionNumber: "TRX-2026-OTHER1",
		Date:              time.Now(),
		Subtotal:          20000,
		GrandTotal:        20000,
		TotalItems:        1,
		PaymentMethod:     "cash",
	}

	require.NoError(t, repo.Create(tx1))
	require.NoError(t, repo.Create(tx2))

	params := PaginationParams{Page: 1, PageSize: 10, Search: "SRCH01", SortBy: "date", SortDir: "desc"}
	list, total, err := repo.List(params, "", "", "")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
	assert.Contains(t, list[0].TransactionNumber, "SRCH01")
}
