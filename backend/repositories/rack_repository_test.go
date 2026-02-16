package repositories

import (
	"testing"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateRack_Valid_Succeeds verifies rack creation with valid data
func TestCreateRack_Valid_Succeeds(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRackRepository(db)

	rack := &models.Rack{
		Name:        "Main Display",
		Code:        "R-001",
		Location:    "Store Front",
		Capacity:    100,
		Description: "Primary display shelf",
		Active:      true,
	}

	err := repo.Create(rack)
	require.NoError(t, err)
	assert.NotZero(t, rack.ID)

	// Verify in database
	var found models.Rack
	err = db.First(&found, rack.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "Main Display", found.Name)
	assert.Equal(t, "R-001", found.Code)
	assert.Equal(t, "Store Front", found.Location)
	assert.Equal(t, 100, found.Capacity)
	assert.Equal(t, "Primary display shelf", found.Description)
	assert.True(t, found.Active)
}

// TestCreateRack_DuplicateCode_ReturnsError verifies unique code constraint
func TestCreateRack_DuplicateCode_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRackRepository(db)

	rack1 := &models.Rack{
		Name:     "Rack 1",
		Code:     "R-001",
		Location: "Location 1",
		Capacity: 50,
		Active:   true,
	}
	err := repo.Create(rack1)
	require.NoError(t, err)

	rack2 := &models.Rack{
		Name:     "Rack 2",
		Code:     "R-001",
		Location: "Location 2",
		Capacity: 75,
		Active:   true,
	}
	err = repo.Create(rack2)
	assert.Error(t, err)
}

// TestCreateRack_DuplicateCodeDifferentCase_ReturnsError verifies case-insensitive unique code
func TestCreateRack_DuplicateCodeDifferentCase_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRackRepository(db)

	rack1 := &models.Rack{
		Name:     "Rack 1",
		Code:     "R-001",
		Location: "Location 1",
		Capacity: 50,
		Active:   true,
	}
	err := repo.Create(rack1)
	require.NoError(t, err)

	rack2 := &models.Rack{
		Name:     "Rack 2",
		Code:     "r-001",
		Location: "Location 2",
		Capacity: 75,
		Active:   true,
	}
	err = repo.Create(rack2)
	assert.Error(t, err)
}

// TestListRacks_SearchByNameCodeLocation_Works verifies search across name, code, and location
func TestListRacks_SearchByNameCodeLocation_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRackRepository(db)

	// Create test racks
	repo.Create(&models.Rack{Name: "Main Display", Code: "MD-001", Location: "Store Front", Capacity: 100, Active: true})
	repo.Create(&models.Rack{Name: "Electronics Shelf", Code: "ES-001", Location: "Store Front", Capacity: 50, Active: true})
	repo.Create(&models.Rack{Name: "Cold Storage", Code: "CS-001", Location: "Warehouse Zone A", Capacity: 200, Active: true})

	// Search by name
	racks, total, err := repo.List(1, 10, "Main", "", "id", "asc")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, racks, 1)
	assert.Equal(t, "Main Display", racks[0].Name)

	// Search by code
	racks, total, err = repo.List(1, 10, "ES-001", "", "id", "asc")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Electronics Shelf", racks[0].Name)

	// Search by location
	racks, total, err = repo.List(1, 10, "Warehouse", "", "id", "asc")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Cold Storage", racks[0].Name)

	// Search matching multiple fields
	racks, total, err = repo.List(1, 10, "Store Front", "", "id", "asc")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, racks, 2)
}

// TestListRacks_FilterByActive_Works verifies active status filtering
func TestListRacks_FilterByActive_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRackRepository(db)

	repo.Create(&models.Rack{Name: "Active Rack", Code: "AR-001", Location: "Location 1", Capacity: 50, Active: true})
	inactiveRack := &models.Rack{Name: "Inactive Rack", Code: "IR-001", Location: "Location 2", Capacity: 30, Active: true}
	repo.Create(inactiveRack)
	db.Model(inactiveRack).Update("active", false)
	repo.Create(&models.Rack{Name: "Active Rack 2", Code: "AR-002", Location: "Location 3", Capacity: 75, Active: true})

	// Filter active only
	racks, total, err := repo.List(1, 10, "", "true", "id", "asc")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, racks, 2)
	for _, r := range racks {
		assert.True(t, r.Active)
	}

	// Filter inactive only
	racks, total, err = repo.List(1, 10, "", "false", "id", "asc")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, racks, 1)
	assert.False(t, racks[0].Active)
}

// TestListRacks_Pagination_Works verifies pagination
func TestListRacks_Pagination_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRackRepository(db)

	for i := 1; i <= 5; i++ {
		repo.Create(&models.Rack{
			Name:     "Rack",
			Code:     "R-" + string(rune('0'+i)),
			Location: "Location",
			Capacity: 50,
			Active:   true,
		})
	}

	racks, total, err := repo.List(1, 2, "", "", "id", "asc")
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, racks, 2)

	racks2, _, err := repo.List(2, 2, "", "", "id", "asc")
	require.NoError(t, err)
	assert.Len(t, racks2, 2)
	assert.NotEqual(t, racks[0].ID, racks2[0].ID)
}

// TestFindRackByID_Exists_ReturnsRack verifies finding rack by ID
func TestFindRackByID_Exists_ReturnsRack(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRackRepository(db)

	rack := &models.Rack{Name: "Test Rack", Code: "TR-001", Location: "Location", Capacity: 50, Active: true}
	err := repo.Create(rack)
	require.NoError(t, err)

	found, err := repo.FindByID(rack.ID)
	require.NoError(t, err)
	assert.Equal(t, rack.ID, found.ID)
	assert.Equal(t, "Test Rack", found.Name)
	assert.Equal(t, "TR-001", found.Code)
}

// TestFindRackByID_NotFound_ReturnsError verifies error for non-existent rack
func TestFindRackByID_NotFound_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRackRepository(db)

	_, err := repo.FindByID(99999)
	assert.Error(t, err)
}

// TestFindRackByCode_Exists_ReturnsRack verifies finding rack by code (case-insensitive)
func TestFindRackByCode_Exists_ReturnsRack(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRackRepository(db)

	rack := &models.Rack{Name: "Test Rack", Code: "TR-001", Location: "Location", Capacity: 50, Active: true}
	err := repo.Create(rack)
	require.NoError(t, err)

	found, err := repo.FindByCode("TR-001")
	require.NoError(t, err)
	assert.Equal(t, rack.ID, found.ID)

	// Case-insensitive
	found, err = repo.FindByCode("tr-001")
	require.NoError(t, err)
	assert.Equal(t, rack.ID, found.ID)
}

// TestFindRackByCodeExcluding_Works verifies code uniqueness check excluding self
func TestFindRackByCodeExcluding_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRackRepository(db)

	rack1 := &models.Rack{Name: "Rack 1", Code: "R-001", Location: "Loc 1", Capacity: 50, Active: true}
	repo.Create(rack1)
	rack2 := &models.Rack{Name: "Rack 2", Code: "R-002", Location: "Loc 2", Capacity: 75, Active: true}
	repo.Create(rack2)

	// Find "R-001" excluding rack1.ID should not find anything
	_, err := repo.FindByCodeExcluding("R-001", rack1.ID)
	assert.Error(t, err)

	// Find "R-001" excluding rack2.ID should find rack1
	found, err := repo.FindByCodeExcluding("R-001", rack2.ID)
	require.NoError(t, err)
	assert.Equal(t, rack1.ID, found.ID)
}

// TestUpdateRack_Valid_Succeeds verifies rack update
func TestUpdateRack_Valid_Succeeds(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRackRepository(db)

	rack := &models.Rack{Name: "OldName", Code: "OLD-001", Location: "Old Loc", Capacity: 50, Active: true}
	err := repo.Create(rack)
	require.NoError(t, err)

	rack.Name = "NewName"
	rack.Location = "New Loc"
	rack.Capacity = 100
	err = repo.Update(rack)
	require.NoError(t, err)

	found, err := repo.FindByID(rack.ID)
	require.NoError(t, err)
	assert.Equal(t, "NewName", found.Name)
	assert.Equal(t, "New Loc", found.Location)
	assert.Equal(t, 100, found.Capacity)
}

// TestDeleteRack_Succeeds verifies rack deletion
func TestDeleteRack_Succeeds(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRackRepository(db)

	rack := &models.Rack{Name: "ToDelete", Code: "DEL-001", Location: "Location", Capacity: 50, Active: true}
	err := repo.Create(rack)
	require.NoError(t, err)

	err = repo.Delete(rack.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(rack.ID)
	assert.Error(t, err)
}

// TestDeleteRack_NotFound_ReturnsError verifies error for deleting non-existent rack
func TestDeleteRack_NotFound_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRackRepository(db)

	err := repo.Delete(99999)
	assert.Error(t, err)
}
