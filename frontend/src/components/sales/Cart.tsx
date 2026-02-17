'use client';

import React, { useMemo } from 'react';
import { useSalesStore } from '@/stores/useSalesStore';
import { formatCurrency } from '@/utils/currency';
import Button from '@/components/ui/Button';
import Badge from '@/components/ui/Badge';

interface CartProps {
  sessionId: number;
}

export default function Cart({ sessionId }: CartProps) {
  const sessions = useSalesStore((state) => state.sessions);
  const productCache = useSalesStore((state) => state.productCache);
  const updateCartItemQuantity = useSalesStore((state) => state.updateCartItemQuantity);
  const updateCartItemUnit = useSalesStore((state) => state.updateCartItemUnit);
  const removeFromCart = useSalesStore((state) => state.removeFromCart);

  const session = sessions.find((s) => s.id === sessionId);

  const cartItems = useMemo(() => {
    if (!session) return [];

    return session.cart.map((cartItem) => {
      const product = productCache[cartItem.productId];
      if (!product) return null;

      const variant = product.variants.find((v) => v.id === cartItem.variantId);
      if (!variant) return null;

      const selectedUnit = product.units.find(
        (u) => String(u.id) === cartItem.selectedUnitId
      );
      if (!selectedUnit) return null;

      const baseUnit = product.units.find((u) => u.isBase);
      if (!baseUnit) return null;

      // Calculate base quantity for pricing and stock validation
      const baseQty = cartItem.quantity * selectedUnit.toBaseUnit;

      // Find the highest tier where baseQty >= tier.minQty
      let applicableTier = variant.pricingTiers[0];
      for (const tier of variant.pricingTiers) {
        if (baseQty >= tier.minQty) {
          applicableTier = tier;
        }
      }

      // Calculate per-unit price
      const perUnitPrice = applicableTier.value * selectedUnit.toBaseUnit;

      // Calculate total
      const total = cartItem.quantity * perUnitPrice;

      // Check if tier pricing is active (not the base tier)
      const isTierPricing = applicableTier.minQty > variant.pricingTiers[0].minQty;

      // Stock validation
      const hasStockError = baseQty > variant.currentStock;

      // Format attributes from array → "Value1, Value2" for display
      const attributesList = variant.attributes
        .map((a) => `${a.attributeName}: ${a.attributeValue}`)
        .join(', ');

      // Get image URL from variant images array (sorted by sortOrder)
      let imageUrl: string | null = null;
      if (variant.images.length > 0) {
        const sorted = [...variant.images].sort((a, b) => a.sortOrder - b.sortOrder);
        imageUrl = sorted[0].imageUrl;
      }

      return {
        cartItem,
        product,
        variant,
        selectedUnit,
        baseUnit,
        baseQty,
        applicableTier,
        perUnitPrice,
        total,
        isTierPricing,
        hasStockError,
        attributesList,
        imageUrl,
      };
    }).filter(Boolean);
  }, [session, productCache]);

  if (!session) {
    return <div className="text-center text-gray-500 py-8">Session not found</div>;
  }

  if (session.cart.length === 0) {
    return (
      <div className="text-center text-gray-500 py-8">
        Cart is empty. Search and add products to get started.
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full border-collapse">
        <thead>
          <tr className="border-b border-gray-200 bg-gray-50">
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-700 uppercase">Image</th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-700 uppercase">SKU</th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-700 uppercase">Name</th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-700 uppercase">Attributes</th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-700 uppercase">Quantity</th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-700 uppercase">Unit</th>
            <th className="px-4 py-3 text-right text-xs font-medium text-gray-700 uppercase">Price</th>
            <th className="px-4 py-3 text-right text-xs font-medium text-gray-700 uppercase">Total</th>
            <th className="px-4 py-3 text-center text-xs font-medium text-gray-700 uppercase">Action</th>
          </tr>
        </thead>
        <tbody>
          {cartItems.map((item) => {
            if (!item) return null;

            const {
              cartItem,
              product,
              variant,
              selectedUnit,
              baseUnit,
              perUnitPrice,
              total,
              isTierPricing,
              hasStockError,
              attributesList,
              imageUrl,
            } = item;

            return (
              <React.Fragment key={variant.id}>
                <tr className="hover:bg-gray-50">
                  {/* Image */}
                  <td className="px-4 py-3" rowSpan={2}>
                    {imageUrl ? (
                      <img
                        src={imageUrl}
                        alt={variant.sku}
                        className="w-12 h-12 object-cover rounded"
                      />
                    ) : (
                      <div className="w-12 h-12 bg-gray-100 rounded flex items-center justify-center">
                        <svg
                          className="w-6 h-6 text-gray-400"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                          />
                        </svg>
                      </div>
                    )}
                  </td>

                  {/* SKU */}
                  <td className="px-4 pt-3 pb-0 text-sm font-medium text-gray-900">
                    {variant.sku}
                  </td>

                  {/* Name */}
                  <td className="px-4 pt-3 pb-0 text-sm text-gray-900">
                    {product.name}
                  </td>

                  {/* Attributes */}
                  <td className="px-4 pt-3 pb-0 text-sm text-gray-600">
                    {attributesList || '-'}
                  </td>

                  {/* Quantity Input */}
                  <td className="px-4 pt-3 pb-0">
                    <input
                      type="number"
                      min="1"
                      value={cartItem.quantity}
                      onChange={(e) => {
                        const value = parseInt(e.target.value);
                        if (!isNaN(value) && value >= 1) {
                          updateCartItemQuantity(sessionId, variant.id, value);
                        }
                      }}
                      className="w-20 rounded-md border border-gray-300 px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </td>

                  {/* Unit Selector */}
                  <td className="px-4 pt-3 pb-0">
                    <select
                      value={cartItem.selectedUnitId}
                      onChange={(e) => updateCartItemUnit(sessionId, variant.id, e.target.value)}
                      className="rounded-md border border-gray-300 px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      {product.units.map((unit) => (
                        <option key={unit.id} value={String(unit.id)}>
                          {unit.name}
                        </option>
                      ))}
                    </select>
                  </td>

                  {/* Price */}
                  <td className="px-4 pt-3 pb-0 text-right text-sm">
                    <div className="flex flex-col items-end">
                      {isTierPricing && (
                        <>
                          <span className="text-gray-400 line-through text-xs">
                            {formatCurrency(variant.pricingTiers[0].value * selectedUnit.toBaseUnit)}
                          </span>
                          <Badge variant="green" className="mb-1">Tier price</Badge>
                        </>
                      )}
                      <span className="font-medium text-gray-900">
                        {formatCurrency(perUnitPrice)}
                      </span>
                    </div>
                  </td>

                  {/* Total */}
                  <td className="px-4 pt-3 pb-0 text-right text-sm font-semibold text-gray-900">
                    {formatCurrency(total)}
                  </td>

                  {/* Remove Button */}
                  <td className="px-4 pt-3 pb-0 text-center" rowSpan={2}>
                    <Button
                      variant="danger"
                      size="sm"
                      onClick={() => removeFromCart(sessionId, variant.id)}
                    >
                      <svg
                        className="w-4 h-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                        />
                      </svg>
                    </Button>
                  </td>
                </tr>

                {/* Second row: description under SKU/Name, stock under Unit */}
                <tr className="border-b border-gray-200">
                  {/* Description spanning SKU + Name + Attributes columns */}
                  <td colSpan={3} className="px-4 pt-0 pb-3 text-xs text-gray-500">
                    {product.description}
                  </td>

                  {/* Spacer for Quantity */}
                  <td className="px-4 pt-0 pb-3"></td>

                  {/* Stock under Unit column */}
                  <td className="px-4 pt-0 pb-3 text-xs">
                    <span className={hasStockError ? 'text-red-600 font-medium' : 'text-gray-500'}>
                      Stock: {variant.currentStock} {baseUnit.name}
                    </span>
                    {hasStockError && (
                      <div className="text-red-600 text-xs mt-0.5">
                        Insufficient stock
                      </div>
                    )}
                  </td>

                  {/* Spacers for Price + Total */}
                  <td className="px-4 pt-0 pb-3"></td>
                  <td className="px-4 pt-0 pb-3"></td>
                </tr>
              </React.Fragment>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
