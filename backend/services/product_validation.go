package services

import (
	"fmt"
	"strings"
)

// ValidateProductInput validates product create/update payload rules that do not require database access.
func ValidateProductInput(input CreateProductInput) error {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if len(name) > 255 {
		return fmt.Errorf("name must be between 1 and 255 characters")
	}

	if input.CategoryID == 0 {
		return fmt.Errorf("categoryId is required")
	}

	switch input.PriceSetting {
	case "fixed":
		if input.MarkupType != nil && strings.TrimSpace(*input.MarkupType) != "" {
			return fmt.Errorf("markupType must be empty when priceSetting is fixed")
		}
	case "markup":
		if input.MarkupType == nil || strings.TrimSpace(*input.MarkupType) == "" {
			return fmt.Errorf("markupType is required when priceSetting is markup")
		}
		if *input.MarkupType != "percentage" && *input.MarkupType != "fixed_amount" {
			return fmt.Errorf("markupType must be percentage or fixed_amount")
		}
	default:
		return fmt.Errorf("priceSetting must be fixed or markup")
	}

	if input.Status == "" {
		input.Status = "active"
	}
	if input.Status != "active" && input.Status != "inactive" {
		return fmt.Errorf("status must be active or inactive")
	}

	if len(input.Units) == 0 {
		return fmt.Errorf("at least one unit is required")
	}
	if err := validateUnits(input.Units); err != nil {
		return err
	}

	if len(input.Variants) == 0 {
		return fmt.Errorf("at least one variant is required")
	}

	if !input.HasVariants && len(input.Variants) != 1 {
		return fmt.Errorf("hasVariants=false requires exactly one variant")
	}

	if input.HasVariants {
		hasAttributes := false
		for _, v := range input.Variants {
			if len(v.Attributes) > 0 {
				hasAttributes = true
				break
			}
		}
		if !hasAttributes {
			return fmt.Errorf("hasVariants=true requires at least one variant with attributes")
		}
	}

	if err := validateVariants(input.Variants); err != nil {
		return err
	}

	return nil
}

func validateUnits(units []CreateProductUnitInput) error {
	baseCount := 0
	seenNames := make(map[string]struct{}, len(units))
	lookup := make(map[string]CreateProductUnitInput, len(units))

	for _, unit := range units {
		name := strings.TrimSpace(unit.Name)
		if name == "" {
			return fmt.Errorf("unit name is required")
		}

		key := strings.ToLower(name)
		if _, exists := seenNames[key]; exists {
			return fmt.Errorf("unit name must be unique")
		}
		seenNames[key] = struct{}{}

		if unit.IsBase {
			baseCount++
			if unit.ConversionFactor != 0 && unit.ConversionFactor != 1 {
				return fmt.Errorf("base unit conversionFactor must be 1")
			}
			if strings.TrimSpace(unit.ConvertsToName) != "" {
				return fmt.Errorf("base unit convertsToName must be empty")
			}
		} else {
			if unit.ConversionFactor <= 0 {
				return fmt.Errorf("conversionFactor must be greater than 0")
			}
			if strings.TrimSpace(unit.ConvertsToName) == "" {
				return fmt.Errorf("convertsToName is required for non-base units")
			}
		}

		lookup[key] = unit
	}

	if baseCount != 1 {
		return fmt.Errorf("exactly one base unit is required")
	}

	for _, unit := range units {
		if unit.IsBase {
			continue
		}
		target := strings.ToLower(strings.TrimSpace(unit.ConvertsToName))
		if _, ok := lookup[target]; !ok {
			return fmt.Errorf("convertsToName must reference another unit in the same request")
		}
	}

	if err := ValidateUnitCircularRef(units); err != nil {
		return err
	}

	return nil
}

func validateVariants(variants []CreateProductVariantInput) error {
	skuSeen := make(map[string]struct{}, len(variants))
	barcodeSeen := make(map[string]struct{}, len(variants))

	for _, variant := range variants {
		sku := strings.TrimSpace(variant.SKU)
		if sku != "" {
			key := strings.ToLower(sku)
			if _, exists := skuSeen[key]; exists {
				return fmt.Errorf("duplicate sku")
			}
			skuSeen[key] = struct{}{}
		}

		barcode := strings.TrimSpace(variant.Barcode)
		if barcode != "" {
			key := strings.ToLower(barcode)
			if _, exists := barcodeSeen[key]; exists {
				return fmt.Errorf("duplicate barcode")
			}
			barcodeSeen[key] = struct{}{}
		}

		if len(variant.PricingTiers) == 0 {
			return fmt.Errorf("at least one pricing tier is required for each variant")
		}

		if variant.PricingTiers[0].MinQty != 1 {
			return fmt.Errorf("first pricing tier minQty must be 1")
		}

		prevMinQty := 0
		for i, tier := range variant.PricingTiers {
			if tier.MinQty <= 0 {
				return fmt.Errorf("pricing tier minQty must be greater than 0")
			}
			if tier.Value < 0 {
				return fmt.Errorf("pricing tier value must be non-negative")
			}
			if i > 0 && tier.MinQty <= prevMinQty {
				return fmt.Errorf("pricing tiers must be sorted by minQty ascending")
			}
			prevMinQty = tier.MinQty
		}
	}

	return nil
}
