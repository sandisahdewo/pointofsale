'use client';

import React, { useState } from 'react';
import { useSalesStore } from '@/stores/useSalesStore';
import { useToastStore } from '@/stores/useToastStore';
import { ApiError } from '@/lib/api';
import Button from '@/components/ui/Button';
import SearchResultsDropdown from './SearchResultsDropdown';

interface ProductSearchProps {
  sessionId: number;
}

export default function ProductSearch({ sessionId }: ProductSearchProps) {
  const [query, setQuery] = useState('');
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const [showHint, setShowHint] = useState(false);

  const { searchResults, isSearching, searchProducts, addToCart, clearSearch } = useSalesStore();
  const { addToast } = useToastStore();

  const handleSearch = async () => {
    if (query.trim().length < 3) {
      setShowHint(true);
      setIsDropdownOpen(false);
      return;
    }

    setShowHint(false);

    try {
      await searchProducts(query.trim());
      setIsDropdownOpen(true);
    } catch (err) {
      if (err instanceof ApiError) {
        addToast(err.message, 'error');
      } else {
        addToast('Search failed. Please try again.', 'error');
      }
      setIsDropdownOpen(false);
    }
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
    clearSearch();
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
        <Button onClick={handleSearch} size="md" disabled={isSearching}>
          {isSearching ? 'Searching...' : 'Search'}
        </Button>
      </div>

      {isDropdownOpen && (
        <SearchResultsDropdown
          products={searchResults}
          onSelectVariant={handleSelectVariant}
          onClose={handleCloseDropdown}
        />
      )}
    </div>
  );
}
