package services

import (
	"errors"
	"sort"
)

// PricingTier represents a tier in the pricing structure.
type PricingTier struct {
	MinQty int
	Value  float64
}

// CalculateTieredPrice returns the per-base-unit price for the given quantity and unit conversion.
// quantity is in the selected unit, toBaseUnit is the conversion factor to base unit.
// It finds the highest tier where baseQty >= tier.MinQty.
func CalculateTieredPrice(tiers []PricingTier, quantity int, toBaseUnit int) (float64, error) {
	if len(tiers) == 0 {
		return 0, errors.New("no pricing tiers defined")
	}

	baseQty := quantity * toBaseUnit

	// Sort tiers by MinQty descending to find highest matching tier
	sorted := make([]PricingTier, len(tiers))
	copy(sorted, tiers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].MinQty > sorted[j].MinQty
	})

	for _, tier := range sorted {
		if baseQty >= tier.MinQty {
			return tier.Value, nil
		}
	}

	// Fallback to lowest tier
	return sorted[len(sorted)-1].Value, nil
}
