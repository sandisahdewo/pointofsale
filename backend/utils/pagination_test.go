package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePaginationParams_Defaults_ReturnsDefaults(t *testing.T) {
	// Arrange: Create request with no query parameters
	req := httptest.NewRequest("GET", "/test", nil)
	allowedSortFields := []string{"id", "name", "email"}

	// Act
	params, err := ParsePaginationParams(req, allowedSortFields)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 10, params.PageSize)
	assert.Equal(t, "", params.Search)
	assert.Equal(t, "id", params.SortBy)
	assert.Equal(t, "asc", params.SortDir)
}

func TestParsePaginationParams_ValidValues_ParsesCorrectly(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test?page=2&pageSize=25&search=test&sortBy=name&sortDir=desc", nil)
	allowedSortFields := []string{"id", "name", "email"}

	// Act
	params, err := ParsePaginationParams(req, allowedSortFields)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 2, params.Page)
	assert.Equal(t, 25, params.PageSize)
	assert.Equal(t, "test", params.Search)
	assert.Equal(t, "name", params.SortBy)
	assert.Equal(t, "desc", params.SortDir)
}

func TestParsePaginationParams_PageSizeExceedsMax_CapsAt100(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test?pageSize=500", nil)
	allowedSortFields := []string{"id"}

	// Act
	params, err := ParsePaginationParams(req, allowedSortFields)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 100, params.PageSize)
}

func TestParsePaginationParams_PageSizeBelowMin_SetsTo1(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test?pageSize=0", nil)
	allowedSortFields := []string{"id"}

	// Act
	params, err := ParsePaginationParams(req, allowedSortFields)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, params.PageSize)
}

func TestParsePaginationParams_InvalidSortDir_DefaultsToAsc(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test?sortDir=invalid", nil)
	allowedSortFields := []string{"id"}

	// Act
	params, err := ParsePaginationParams(req, allowedSortFields)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "asc", params.SortDir)
}

func TestParsePaginationParams_InvalidSortBy_ReturnsError(t *testing.T) {
	// Arrange: sortBy field not in allowlist
	req := httptest.NewRequest("GET", "/test?sortBy=malicious_field", nil)
	allowedSortFields := []string{"id", "name", "email"}

	// Act
	params, err := ParsePaginationParams(req, allowedSortFields)

	// Assert
	require.Error(t, err)
	assert.Nil(t, params)
	assert.Contains(t, err.Error(), "invalid sort field")
}

func TestParsePaginationParams_CalculatesOffsetCorrectly(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test?page=3&pageSize=20", nil)
	allowedSortFields := []string{"id"}

	// Act
	params, err := ParsePaginationParams(req, allowedSortFields)

	// Assert
	require.NoError(t, err)
	offset := params.GetOffset()
	assert.Equal(t, 40, offset) // (3-1) * 20 = 40
}

func TestPaginationMeta_CalculatesTotalPages_RoundUp(t *testing.T) {
	// Test that 25 items with pageSize 10 gives 3 pages
	meta := CalculatePaginationMeta(1, 10, 25)
	assert.Equal(t, 3, meta.TotalPages)

	// Test exact division: 30 items with pageSize 10 gives 3 pages
	meta = CalculatePaginationMeta(1, 10, 30)
	assert.Equal(t, 3, meta.TotalPages)

	// Test 0 items gives 0 pages
	meta = CalculatePaginationMeta(1, 10, 0)
	assert.Equal(t, 0, meta.TotalPages)
}

func TestPaginationMeta_StructureIsCorrect(t *testing.T) {
	// Arrange & Act
	meta := CalculatePaginationMeta(2, 15, 47)

	// Assert
	assert.Equal(t, 2, meta.Page)
	assert.Equal(t, 15, meta.PageSize)
	assert.Equal(t, 47, meta.TotalItems)
	assert.Equal(t, 4, meta.TotalPages) // ceil(47/15) = 4
}
