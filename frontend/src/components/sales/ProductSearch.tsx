'use client';

import React, { useState } from 'react';
import { useProductStore, Product } from '@/stores/useProductStore';
import { useSalesStore } from '@/stores/useSalesStore';
import Button from '@/components/ui/Button';
import SearchResultsDropdown from './SearchResultsDropdown';

interface ProductSearchProps {
  sessionId: number;
}

export default function ProductSearch({ sessionId }: ProductSearchProps) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<Product[]>([]);
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const [showHint, setShowHint] = useState(false);

  const { products } = useProductStore();
  const { addToCart } = useSalesStore();

  const handleSearch = () => {
    if (query.trim().length < 3) {
      setShowHint(true);
      setIsDropdownOpen(false);
      return;
    }

    setShowHint(false);

    const searchQuery = query.trim().toLowerCase();

    // Filter active products and search by name, sku, or barcode
    const matchingProducts = products
      .filter((product) => product.status === 'active')
      .filter((product) => {
        // Check product name
        if (product.name.toLowerCase().includes(searchQuery)) {
          return true;
        }
        // Check variant SKU or barcode
        return product.variants.some(
          (variant) =>
            variant.sku.toLowerCase().includes(searchQuery) ||
            variant.barcode.toLowerCase().includes(searchQuery)
        );
      })
      .slice(0, 10); // Limit to 10 results

    setResults(matchingProducts);
    setIsDropdownOpen(true);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setQuery(e.target.value);
    if (showHint && e.target.value.length >= 3) {
      setShowHint(false);
    }
  };

  const handleSelectVariant = (productId: number, variantId: string) => {
    addToCart(sessionId, productId, variantId);
  };

  const handleCloseDropdown = () => {
    setIsDropdownOpen(false);
  };

  return (
    <div className="relative">
      <div className="flex gap-2">
        <div className="flex-1">
          <input
            type="text"
            value={query}
            onChange={handleInputChange}
            onKeyDown={handleKeyDown}
            placeholder="Search by product name, SKU, or barcode..."
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
          {showHint && (
            <p className="mt-1 text-sm text-gray-500">
              Type at least 3 characters
            </p>
          )}
        </div>
        <Button onClick={handleSearch} size="md">
          Search
        </Button>
      </div>

      {isDropdownOpen && (
        <SearchResultsDropdown
          products={results}
          onSelectVariant={handleSelectVariant}
          onClose={handleCloseDropdown}
        />
      )}
    </div>
  );
}
