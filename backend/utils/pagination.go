package utils

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

// PaginationParams holds query parameters for pagination
type PaginationParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Search   string `json:"search"`
	SortBy   string `json:"sortBy"`
	SortDir  string `json:"sortDir"`
}

// PaginationMeta holds metadata for paginated responses
type PaginationMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

// PaginatedResponse wraps data with pagination metadata
type PaginatedResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// ParsePaginationParams extracts and validates pagination parameters from request query string.
// Returns error if sortBy field is not in the allowlist (prevents SQL injection).
func ParsePaginationParams(r *http.Request, allowedSortFields []string) (*PaginationParams, error) {
	query := r.URL.Query()

	// Parse page (default: 1)
	page := 1
	if p := query.Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	// Parse pageSize (default: 10, min: 1, max: 100)
	pageSize := 10
	if ps := query.Get("pageSize"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil {
			pageSize = val
		}
	}
	if pageSize < 1 {
		pageSize = 1
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Parse search (optional)
	search := query.Get("search")

	// Parse sortBy (default: first allowed field or "id")
	sortBy := "id"
	if len(allowedSortFields) > 0 {
		sortBy = allowedSortFields[0]
	}
	if sb := query.Get("sortBy"); sb != "" {
		// Validate against allowlist to prevent SQL injection
		if !contains(allowedSortFields, sb) {
			return nil, fmt.Errorf("invalid sort field: %s", sb)
		}
		sortBy = sb
	}

	// Parse sortDir (default: "asc", allowed: "asc" or "desc")
	sortDir := "asc"
	if sd := query.Get("sortDir"); sd != "" {
		if sd == "asc" || sd == "desc" {
			sortDir = sd
		}
	}

	return &PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
		SortBy:   sortBy,
		SortDir:  sortDir,
	}, nil
}

// GetOffset calculates the database offset for the current page
func (p *PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// ApplyToQuery applies pagination, sorting to a GORM query
func (p *PaginationParams) ApplyToQuery(query *gorm.DB) *gorm.DB {
	return query.
		Offset(p.GetOffset()).
		Limit(p.PageSize).
		Order(p.SortBy + " " + p.SortDir)
}

// CalculatePaginationMeta calculates pagination metadata from total count
func CalculatePaginationMeta(page, pageSize, totalItems int) PaginationMeta {
	totalPages := 0
	if totalItems > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(pageSize)))
	}

	return PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

// contains checks if a string slice contains a specific value
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
