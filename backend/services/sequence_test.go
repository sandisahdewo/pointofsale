package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/pointofsale/backend/testutil"
)

func TestGeneratePONumber_FirstOfYear_ReturnsPO_YYYY_0001(t *testing.T) {
	db := testutil.SetupTestDB(t)

	seq := NewSequenceService(db)
	poNumber, err := seq.GeneratePONumber()
	require.NoError(t, err)

	year := time.Now().Year()
	expected := formatPONumber(year, 1)
	assert.Equal(t, expected, poNumber)
}

func TestGeneratePONumber_Increment_ReturnsPO_YYYY_0002(t *testing.T) {
	db := testutil.SetupTestDB(t)

	// Create a PO with number 0001
	year := time.Now().Year()
	createPOWithNumber(t, db, formatPONumber(year, 1))

	seq := NewSequenceService(db)
	poNumber, err := seq.GeneratePONumber()
	require.NoError(t, err)

	expected := formatPONumber(year, 2)
	assert.Equal(t, expected, poNumber)
}

func TestGeneratePONumber_NewYear_ResetsSequence(t *testing.T) {
	db := testutil.SetupTestDB(t)

	// Create a PO from last year
	lastYear := time.Now().Year() - 1
	createPOWithNumber(t, db, formatPONumber(lastYear, 5))

	seq := NewSequenceService(db)
	poNumber, err := seq.GeneratePONumber()
	require.NoError(t, err)

	thisYear := time.Now().Year()
	expected := formatPONumber(thisYear, 1)
	assert.Equal(t, expected, poNumber)
}

func TestGenerateTrxNumber_FirstEver_ReturnsTRX_YYYY_000001(t *testing.T) {
	db := testutil.SetupTestDB(t)

	seq := NewSequenceService(db)
	trxNumber, err := seq.GenerateTrxNumber()
	require.NoError(t, err)

	year := time.Now().Year()
	expected := formatTrxNumber(year, 1)
	assert.Equal(t, expected, trxNumber)
}

func TestGenerateTrxNumber_Increment_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)

	// Create a transaction with number 000001
	year := time.Now().Year()
	createTrxWithNumber(t, db, formatTrxNumber(year, 1))

	seq := NewSequenceService(db)
	trxNumber, err := seq.GenerateTrxNumber()
	require.NoError(t, err)

	expected := formatTrxNumber(year, 2)
	assert.Equal(t, expected, trxNumber)
}

// helpers
func createPOWithNumber(t *testing.T, db *gorm.DB, poNumber string) {
	t.Helper()
	supplier := testutil.CreateTestSupplier(t, db)
	err := db.Exec(
		"INSERT INTO purchase_orders (po_number, supplier_id, date, status) VALUES (?, ?, CURRENT_DATE, 'draft')",
		poNumber, supplier.ID,
	).Error
	require.NoError(t, err)
}

func createTrxWithNumber(t *testing.T, db *gorm.DB, trxNumber string) {
	t.Helper()
	err := db.Exec(
		"INSERT INTO sales_transactions (transaction_number, date, subtotal, grand_total, total_items, payment_method) VALUES (?, NOW(), 0, 0, 0, 'cash')",
		trxNumber,
	).Error
	require.NoError(t, err)
}
