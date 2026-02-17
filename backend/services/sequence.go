package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SequenceService generates sequential numbers for POs and transactions.
type SequenceService struct {
	db *gorm.DB
}

// NewSequenceService creates a new sequence service.
func NewSequenceService(db *gorm.DB) *SequenceService {
	return &SequenceService{db: db}
}

// GeneratePONumber generates the next PO number in format PO-YYYY-NNNN.
func (s *SequenceService) GeneratePONumber() (string, error) {
	year := time.Now().Year()
	prefix := fmt.Sprintf("PO-%d-", year)

	var lastNumber string
	err := s.db.Raw(
		"SELECT po_number FROM purchase_orders WHERE po_number LIKE ? ORDER BY po_number DESC LIMIT 1",
		prefix+"%",
	).Scan(&lastNumber).Error
	if err != nil {
		return "", err
	}

	nextSeq := 1
	if lastNumber != "" {
		parts := strings.Split(lastNumber, "-")
		if len(parts) == 3 {
			if n, err := strconv.Atoi(parts[2]); err == nil {
				nextSeq = n + 1
			}
		}
	}

	return formatPONumber(year, nextSeq), nil
}

// GenerateTrxNumber generates the next transaction number in format TRX-YYYY-NNNNNN.
func (s *SequenceService) GenerateTrxNumber() (string, error) {
	year := time.Now().Year()
	prefix := fmt.Sprintf("TRX-%d-", year)

	var lastNumber string
	err := s.db.Raw(
		"SELECT transaction_number FROM sales_transactions WHERE transaction_number LIKE ? ORDER BY transaction_number DESC LIMIT 1",
		prefix+"%",
	).Scan(&lastNumber).Error
	if err != nil {
		return "", err
	}

	nextSeq := 1
	if lastNumber != "" {
		parts := strings.Split(lastNumber, "-")
		if len(parts) == 3 {
			if n, err := strconv.Atoi(parts[2]); err == nil {
				nextSeq = n + 1
			}
		}
	}

	return formatTrxNumber(year, nextSeq), nil
}

func formatPONumber(year, seq int) string {
	return fmt.Sprintf("PO-%d-%04d", year, seq)
}

func formatTrxNumber(year, seq int) string {
	return fmt.Sprintf("TRX-%d-%06d", year, seq)
}
