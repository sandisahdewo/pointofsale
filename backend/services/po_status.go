package services

import "fmt"

// validPOTransitions defines allowed PO status transitions
var validPOTransitions = map[string][]string{
	"draft":    {"sent", "cancelled"},
	"sent":     {"cancelled"},
	"received": {"completed"},
}

// ValidatePOStatusTransition checks if the transition from current to next status is allowed.
func ValidatePOStatusTransition(current, next string) error {
	allowed, exists := validPOTransitions[current]
	if !exists {
		return fmt.Errorf("invalid status transition from %s to %s", current, next)
	}
	for _, s := range allowed {
		if s == next {
			return nil
		}
	}
	return fmt.Errorf("invalid status transition from %s to %s", current, next)
}
