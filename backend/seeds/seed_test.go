package seeds

import (
	"testing"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeedCategories_CreatesExpectedData(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	// Act
	err := seedCategories(db)

	// Assert
	require.NoError(t, err)

	var categories []models.Category
	err = db.Order("id").Find(&categories).Error
	require.NoError(t, err)

	assert.Len(t, categories, 4)
	assert.Equal(t, "Clothing", categories[0].Name)
	assert.Equal(t, "Apparel and garments", categories[0].Description)
	assert.Equal(t, "Food & Beverages", categories[1].Name)
	assert.Equal(t, "Food items and drinks", categories[1].Description)
	assert.Equal(t, "Stationery", categories[2].Name)
	assert.Equal(t, "Office and school supplies", categories[2].Description)
	assert.Equal(t, "Household", categories[3].Name)
	assert.Equal(t, "Home and kitchen essentials", categories[3].Description)
}

func TestSeedSuppliers_CreatesWithBankAccounts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	// Act
	err := seedSuppliers(db)

	// Assert
	require.NoError(t, err)

	var suppliers []models.Supplier
	err = db.Preload("BankAccounts").Order("id").Find(&suppliers).Error
	require.NoError(t, err)

	assert.Len(t, suppliers, 4)

	// PT Sumber Makmur
	assert.Equal(t, "PT Sumber Makmur", suppliers[0].Name)
	assert.Equal(t, "Jl. Industri No. 45, Jakarta", suppliers[0].Address)
	assert.Equal(t, "+62-21-5550001", suppliers[0].Phone)
	assert.Equal(t, "order@sumbermakmur.co.id", suppliers[0].Email)
	assert.Equal(t, "sumbermakmur.co.id", suppliers[0].Website)
	assert.True(t, suppliers[0].Active)
	assert.Len(t, suppliers[0].BankAccounts, 2)
	assert.Equal(t, "BCA - Main Account", suppliers[0].BankAccounts[0].AccountName)
	assert.Equal(t, "1234567890", suppliers[0].BankAccounts[0].AccountNumber)
	assert.Equal(t, "Mandiri - Operations", suppliers[0].BankAccounts[1].AccountName)
	assert.Equal(t, "0987654321", suppliers[0].BankAccounts[1].AccountNumber)

	// CV Jaya Abadi
	assert.Equal(t, "CV Jaya Abadi", suppliers[1].Name)
	assert.Equal(t, "Jl. Perdagangan No. 12, Surabaya", suppliers[1].Address)
	assert.True(t, suppliers[1].Active)
	assert.Len(t, suppliers[1].BankAccounts, 1)
	assert.Equal(t, "BCA - Main Account", suppliers[1].BankAccounts[0].AccountName)
	assert.Equal(t, "1122334455", suppliers[1].BankAccounts[0].AccountNumber)

	// UD Berkah Sentosa (no bank accounts)
	assert.Equal(t, "UD Berkah Sentosa", suppliers[2].Name)
	assert.True(t, suppliers[2].Active)
	assert.Len(t, suppliers[2].BankAccounts, 0)

	// PT Global Supplies (inactive, 2 bank accounts)
	assert.Equal(t, "PT Global Supplies", suppliers[3].Name)
	assert.False(t, suppliers[3].Active)
	assert.Len(t, suppliers[3].BankAccounts, 2)
	assert.Equal(t, "BNI - Main Account", suppliers[3].BankAccounts[0].AccountName)
	assert.Equal(t, "5566778899", suppliers[3].BankAccounts[0].AccountNumber)
	assert.Equal(t, "BRI - Operations", suppliers[3].BankAccounts[1].AccountName)
	assert.Equal(t, "9988776655", suppliers[3].BankAccounts[1].AccountNumber)
}

func TestSeedRacks_CreatesExpectedData(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	// Act
	err := seedRacks(db)

	// Assert
	require.NoError(t, err)

	var racks []models.Rack
	err = db.Order("id").Find(&racks).Error
	require.NoError(t, err)

	assert.Len(t, racks, 5)

	// Main Display
	assert.Equal(t, "Main Display", racks[0].Name)
	assert.Equal(t, "R-001", racks[0].Code)
	assert.Equal(t, "Store Front", racks[0].Location)
	assert.Equal(t, 100, racks[0].Capacity)
	assert.Equal(t, "Primary display shelf near entrance", racks[0].Description)
	assert.True(t, racks[0].Active)

	// Electronics Shelf
	assert.Equal(t, "Electronics Shelf", racks[1].Name)
	assert.Equal(t, "R-002", racks[1].Code)
	assert.True(t, racks[1].Active)

	// Cold Storage
	assert.Equal(t, "Cold Storage", racks[2].Name)
	assert.Equal(t, "R-003", racks[2].Code)
	assert.Equal(t, 200, racks[2].Capacity)
	assert.True(t, racks[2].Active)

	// Bulk Storage
	assert.Equal(t, "Bulk Storage", racks[3].Name)
	assert.Equal(t, "R-004", racks[3].Code)
	assert.Equal(t, 500, racks[3].Capacity)
	assert.True(t, racks[3].Active)

	// Clearance Rack (inactive)
	assert.Equal(t, "Clearance Rack", racks[4].Name)
	assert.Equal(t, "R-005", racks[4].Code)
	assert.Equal(t, 30, racks[4].Capacity)
	assert.False(t, racks[4].Active)
}

func TestSeedIdempotent_RunTwice_NoErrors(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	// Act - Run seeds twice
	err := seedCategories(db)
	require.NoError(t, err)
	err = seedCategories(db)
	require.NoError(t, err)

	err = seedSuppliers(db)
	require.NoError(t, err)
	err = seedSuppliers(db)
	require.NoError(t, err)

	err = seedRacks(db)
	require.NoError(t, err)
	err = seedRacks(db)
	require.NoError(t, err)

	// Assert - Still only one set of data
	var categoryCount, supplierCount, rackCount int64
	db.Model(&models.Category{}).Count(&categoryCount)
	db.Model(&models.Supplier{}).Count(&supplierCount)
	db.Model(&models.Rack{}).Count(&rackCount)

	assert.Equal(t, int64(4), categoryCount)
	assert.Equal(t, int64(4), supplierCount)
	assert.Equal(t, int64(5), rackCount)
}
