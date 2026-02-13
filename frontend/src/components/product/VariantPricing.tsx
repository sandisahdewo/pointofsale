'use client';

import React from 'react';
import Button from '@/components/ui/Button';
import type { PriceSetting, MarkupType, PricingTier } from '@/stores/useProductStore';

interface VariantPricingProps {
  priceSetting: PriceSetting;
  markupType?: MarkupType;
  pricingTiers?: PricingTier[];
  disabled?: boolean;
  onChange: (fields: {
    pricingTiers?: PricingTier[];
  }) => void;
}

export default function VariantPricing({
  priceSetting,
  markupType,
  pricingTiers = [],
  disabled = false,
  onChange,
}: VariantPricingProps) {
  const tiers = pricingTiers.length > 0 ? pricingTiers : [{ minQty: 1, value: 0 }];

  const updateTier = (index: number, field: keyof PricingTier, raw: string) => {
    const updated = tiers.map((t, i) =>
      i === index ? { ...t, [field]: raw === '' ? 0 : Number(raw) } : t
    );
    onChange({ pricingTiers: updated });
  };

  const addTier = () => {
    const lastQty = tiers[tiers.length - 1]?.minQty ?? 0;
    onChange({
      pricingTiers: [...tiers, { minQty: lastQty + 1, value: 0 }],
    });
  };

  const removeTier = (index: number) => {
    if (index === 0) return; // cannot remove first tier
    onChange({ pricingTiers: tiers.filter((_, i) => i !== index) });
  };

  // Determine column header for value
  let valueLabel = 'Sell Price';
  if (priceSetting === 'markup') {
    valueLabel = markupType === 'percentage' ? 'Markup (%)' : 'Markup Amount';
  }

  // Check for non-descending prices warning
  const priceWarnings: Set<number> = new Set();
  if (priceSetting === 'fixed') {
    for (let i = 1; i < tiers.length; i++) {
      if (tiers[i].value >= tiers[i - 1].value && tiers[i - 1].value > 0) {
        priceWarnings.add(i);
      }
    }
  }

  // Check ascending qty order
  const qtyWarnings: Set<number> = new Set();
  for (let i = 1; i < tiers.length; i++) {
    if (tiers[i].minQty <= tiers[i - 1].minQty) {
      qtyWarnings.add(i);
    }
  }

  return (
    <div className="space-y-2">
      <p className="text-sm font-medium text-gray-700">Tiered Pricing</p>
      <div className="text-xs text-blue-600 bg-blue-50 border border-blue-200 rounded-md px-3 py-2">
        All pricing uses tiered structure. For retail (single price), set one tier with Min Qty = 1. Add more tiers for volume/wholesale pricing.
      </div>
      <table className="w-full text-sm border border-gray-200 rounded-md overflow-hidden">
        <thead>
          <tr className="bg-gray-50">
            <th className="text-left px-3 py-2 font-medium text-gray-600">Min Qty</th>
            <th className="text-left px-3 py-2 font-medium text-gray-600">{valueLabel}</th>
            <th className="w-16 px-3 py-2" />
          </tr>
        </thead>
        <tbody>
          {tiers.map((tier, idx) => (
            <tr key={idx} className="border-t border-gray-200">
              <td className="px-3 py-2">
                <input
                  type="number"
                  min={1}
                  className={`w-24 rounded border px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500 ${
                    disabled || idx === 0
                      ? 'bg-gray-100 text-gray-500 cursor-not-allowed border-gray-200'
                      : qtyWarnings.has(idx)
                        ? 'border-red-400'
                        : 'border-gray-300'
                  }`}
                  value={tier.minQty}
                  disabled={disabled}
                  readOnly={idx === 0}
                  onChange={(e) => updateTier(idx, 'minQty', e.target.value)}
                />
                {qtyWarnings.has(idx) && !disabled && (
                  <p className="text-xs text-red-500 mt-0.5">Must be greater than previous tier</p>
                )}
              </td>
              <td className="px-3 py-2">
                <input
                  type="number"
                  min={0}
                  step={priceSetting === 'markup' && markupType === 'percentage' ? '0.1' : '0.01'}
                  className={`w-32 rounded border px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500 ${
                    disabled
                      ? 'bg-gray-100 text-gray-500 cursor-not-allowed border-gray-200'
                      : priceWarnings.has(idx)
                        ? 'border-yellow-400'
                        : 'border-gray-300'
                  }`}
                  value={tier.value || ''}
                  placeholder="0"
                  disabled={disabled}
                  onChange={(e) => updateTier(idx, 'value', e.target.value)}
                />
                {priceWarnings.has(idx) && !disabled && (
                  <p className="text-xs text-yellow-600 mt-0.5">
                    Price usually decreases at higher quantities
                  </p>
                )}
              </td>
              <td className="px-3 py-2 text-center">
                {idx > 0 && !disabled && (
                  <button
                    type="button"
                    onClick={() => removeTier(idx)}
                    className="text-red-500 hover:text-red-700 cursor-pointer"
                    title="Remove tier"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {!disabled && (
        <Button type="button" variant="outline" size="sm" onClick={addTier}>
          + Add Tier
        </Button>
      )}
      {priceSetting === 'markup' && (
        <p className="text-xs text-gray-500">
          {markupType === 'percentage'
            ? 'Sell price will be calculated from purchase cost at transaction time.'
            : 'Sell price will be calculated by adding this amount to purchase cost.'}
        </p>
      )}
    </div>
  );
}
