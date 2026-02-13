'use client';

import React, { useMemo } from 'react';
import { useSalesStore } from '@/stores/useSalesStore';
import { useProductStore } from '@/stores/useProductStore';
import { formatCurrency } from '@/utils/currency';

interface CartSummaryProps {
  sessionId: number;
}

export default function CartSummary({ sessionId }: CartSummaryProps) {
  const sessions = useSalesStore((state) => state.sessions);
  const products = useProductStore((state) => state.products);

  const session = sessions.find((s) => s.id === sessionId);

  const summary = useMemo(() => {
    if (!session) {
      return {
        totalItems: 0,
        subtotal: 0,
        grandTotal: 0,
      };
    }

    let subtotal = 0;

    for (const cartItem of session.cart) {
      const product = products.find((p) => p.id === cartItem.productId);
      if (!product) continue;

      const variant = product.variants.find((v) => v.id === cartItem.variantId);
      if (!variant) continue;

      const selectedUnit = product.units.find((u) => u.id === cartItem.selectedUnitId);
      if (!selectedUnit) continue;

      // Calculate base quantity for pricing
      const baseQty = cartItem.quantity * selectedUnit.toBaseUnit;

      // Find the highest tier where baseQty >= tier.minQty
      let applicableTier = variant.pricingTiers[0];
      for (const tier of variant.pricingTiers) {
        if (baseQty >= tier.minQty) {
          applicableTier = tier;
        }
      }

      // Calculate per-unit price and total
      const perUnitPrice = applicableTier.value * selectedUnit.toBaseUnit;
      const itemTotal = cartItem.quantity * perUnitPrice;

      subtotal += itemTotal;
    }

    return {
      totalItems: session.cart.length,
      subtotal,
      grandTotal: subtotal, // No tax/discount for now
    };
  }, [session, products]);

  if (!session) {
    return null;
  }

  return (
    <div className="bg-white rounded-lg border border-gray-200 p-6 space-y-4">
      <h3 className="text-lg font-semibold text-gray-900 border-b border-gray-200 pb-3">
        Cart Summary
      </h3>

      <div className="space-y-3">
        <div className="flex justify-between items-center">
          <span className="text-sm text-gray-600">Total Items:</span>
          <span className="text-sm font-medium text-gray-900">
            {summary.totalItems}
          </span>
        </div>

        <div className="flex justify-between items-center">
          <span className="text-sm text-gray-600">Subtotal:</span>
          <span className="text-sm font-medium text-gray-900">
            {formatCurrency(summary.subtotal)}
          </span>
        </div>

        <div className="border-t border-gray-200 pt-3">
          <div className="flex justify-between items-center">
            <span className="text-base font-semibold text-gray-900">Grand Total:</span>
            <span className="text-xl font-bold text-blue-600">
              {formatCurrency(summary.grandTotal)}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
