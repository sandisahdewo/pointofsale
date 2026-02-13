'use client';

import React from 'react';
import { Product } from '@/stores/useProductStore';
import Button from '@/components/ui/Button';

interface SearchResultsDropdownProps {
  products: Product[];
  onSelectVariant: (productId: number, variantId: string) => void;
  onClose: () => void;
}

export default function SearchResultsDropdown({
  products,
  onSelectVariant,
  onClose,
}: SearchResultsDropdownProps) {
  const getImageUrl = (images: string[]): string | null => {
    return images.length > 0 ? images[0] : null;
  };

  const formatAttributes = (attributes: Record<string, string>): string => {
    return Object.values(attributes).join(', ');
  };

  return (
    <div className="absolute z-50 mt-2 w-full border border-gray-200 rounded-lg shadow-lg bg-white max-h-96 overflow-y-auto">
      {/* Close button */}
      <div className="sticky top-0 bg-white border-b border-gray-200 px-4 py-2 flex justify-end">
        <button
          onClick={onClose}
          className="text-gray-500 hover:text-gray-700 focus:outline-none"
          aria-label="Close"
        >
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>

      {/* Results */}
      <div className="p-2">
        {products.length === 0 ? (
          <div className="text-center py-8 text-gray-500">No results found</div>
        ) : (
          <div className="space-y-2">
            {products.map((product) => (
              <div key={product.id} className="border-b border-gray-200 last:border-0 pb-2 last:pb-0">
                {product.hasVariants ? (
                  // Product with variants
                  <div>
                    <div className="flex items-center gap-2 mb-2">
                      {getImageUrl(product.images) ? (
                        <img
                          src={getImageUrl(product.images)!}
                          alt={product.name}
                          className="w-12 h-12 object-cover rounded"
                        />
                      ) : (
                        <div className="w-12 h-12 bg-gray-200 rounded flex items-center justify-center">
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
                              d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"
                            />
                          </svg>
                        </div>
                      )}
                      <div className="font-medium text-gray-900">{product.name}</div>
                    </div>
                    <div className="space-y-1 ml-14">
                      {product.variants.map((variant) => {
                        const isOutOfStock = variant.currentStock === 0;
                        return (
                          <div
                            key={variant.id}
                            className={`flex items-center gap-2 px-2 py-1.5 rounded ${
                              isOutOfStock ? 'bg-red-50' : ''
                            }`}
                          >
                            {getImageUrl(variant.images) ? (
                              <img
                                src={getImageUrl(variant.images)!}
                                alt={variant.sku}
                                className="w-10 h-10 object-cover rounded"
                              />
                            ) : (
                              <div className="w-10 h-10 bg-gray-200 rounded flex items-center justify-center">
                                <svg
                                  className="w-5 h-5 text-gray-400"
                                  fill="none"
                                  stroke="currentColor"
                                  viewBox="0 0 24 24"
                                >
                                  <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    strokeWidth={2}
                                    d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"
                                  />
                                </svg>
                              </div>
                            )}
                            <div className="flex-1 grid grid-cols-4 gap-2 items-center text-sm">
                              <div className="text-gray-600">{variant.sku}</div>
                              <div className="text-gray-600">
                                {formatAttributes(variant.attributes)}
                              </div>
                              <div className="text-gray-600">
                                Stock: {variant.currentStock}
                              </div>
                              <div className="flex justify-end">
                                <Button
                                  onClick={() => onSelectVariant(product.id, variant.id)}
                                  disabled={isOutOfStock}
                                  size="sm"
                                >
                                  Select
                                </Button>
                              </div>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>
                ) : (
                  // Product without variants (single variant)
                  <div>
                    {product.variants.map((variant) => {
                      const isOutOfStock = variant.currentStock === 0;
                      return (
                        <div
                          key={variant.id}
                          className={`flex items-center gap-2 px-2 py-1.5 rounded ${
                            isOutOfStock ? 'bg-red-50' : ''
                          }`}
                        >
                          {getImageUrl(product.images) ? (
                            <img
                              src={getImageUrl(product.images)!}
                              alt={product.name}
                              className="w-12 h-12 object-cover rounded"
                            />
                          ) : (
                            <div className="w-12 h-12 bg-gray-200 rounded flex items-center justify-center">
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
                                  d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"
                                />
                              </svg>
                            </div>
                          )}
                          <div className="flex-1 grid grid-cols-4 gap-2 items-center text-sm">
                            <div className="font-medium text-gray-900">{product.name}</div>
                            <div className="text-gray-600">{variant.sku}</div>
                            <div className="text-gray-600">
                              Stock: {variant.currentStock}
                            </div>
                            <div className="flex justify-end">
                              <Button
                                onClick={() => onSelectVariant(product.id, variant.id)}
                                disabled={isOutOfStock}
                                size="sm"
                              >
                                Select
                              </Button>
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
