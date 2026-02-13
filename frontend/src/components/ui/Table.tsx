'use client';

import React from 'react';

export interface Column<T> {
  key: string;
  label: string;
  sortable?: boolean;
  render?: (item: T) => React.ReactNode;
}

export type SortDirection = 'asc' | 'desc' | null;

interface TableProps<T> {
  columns: Column<T>[];
  data: T[];
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
  sortKey?: string | null;
  sortDirection?: SortDirection;
  onSort?: (key: string, direction: SortDirection) => void;
  pageSize?: number;
  onPageSizeChange?: (size: number) => void;
  totalItems?: number;
}

const PAGE_SIZE_OPTIONS = [5, 10, 25, 50, 100];

export default function Table<T extends { id: number }>({
  columns,
  data,
  currentPage,
  totalPages,
  onPageChange,
  sortKey = null,
  sortDirection = null,
  onSort,
  pageSize = 10,
  onPageSizeChange,
  totalItems,
}: TableProps<T>) {
  const handleSortClick = (col: Column<T>) => {
    if (!col.sortable || !onSort) return;
    let nextDirection: SortDirection;
    if (sortKey !== col.key || sortDirection === null) {
      nextDirection = 'asc';
    } else if (sortDirection === 'asc') {
      nextDirection = 'desc';
    } else {
      nextDirection = null;
    }
    onSort(col.key, nextDirection);
  };

  const renderSortIcon = (col: Column<T>) => {
    if (!col.sortable) return null;
    if (sortKey === col.key && sortDirection === 'asc') {
      return <span className="ml-1 text-blue-600">&#9650;</span>;
    }
    if (sortKey === col.key && sortDirection === 'desc') {
      return <span className="ml-1 text-blue-600">&#9660;</span>;
    }
    return <span className="ml-1 text-gray-400">&#8693;</span>;
  };

  const showingStart = totalItems && totalItems > 0 ? (currentPage - 1) * pageSize + 1 : 0;
  const showingEnd = totalItems ? Math.min(currentPage * pageSize, totalItems) : 0;

  return (
    <div>
      <div className="overflow-x-auto">
        <table className="w-full text-sm text-left">
          <thead className="bg-gray-50 text-gray-600 uppercase text-xs">
            <tr>
              {columns.map((col) => (
                <th
                  key={col.key}
                  className={`px-4 py-3 font-medium${col.sortable ? ' cursor-pointer select-none hover:text-gray-900' : ''}`}
                  onClick={() => handleSortClick(col)}
                >
                  <span className="inline-flex items-center">
                    {col.label}
                    {renderSortIcon(col)}
                  </span>
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {data.length === 0 ? (
              <tr>
                <td
                  colSpan={columns.length}
                  className="px-4 py-8 text-center text-gray-500"
                >
                  No data available
                </td>
              </tr>
            ) : (
              data.map((item) => (
                <tr key={item.id} className="hover:bg-gray-50">
                  {columns.map((col) => (
                    <td key={col.key} className="px-4 py-3 text-gray-700">
                      {col.render
                        ? col.render(item)
                        : (item as Record<string, unknown>)[col.key] as React.ReactNode}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <div className="flex items-center justify-between px-4 py-3 border-t border-gray-200">
        <div className="flex items-center gap-4">
          {totalItems !== undefined && (
            <p className="text-sm text-gray-600">
              Showing {showingStart}-{showingEnd} of {totalItems} items
            </p>
          )}
          {!totalItems && totalItems !== 0 && (
            <p className="text-sm text-gray-600">
              Page {currentPage} of {totalPages}
            </p>
          )}
          {onPageSizeChange && (
            <div className="flex items-center gap-2">
              <label htmlFor="page-size" className="text-sm text-gray-600">
                Items per page:
              </label>
              <select
                id="page-size"
                value={pageSize}
                onChange={(e) => onPageSizeChange(Number(e.target.value))}
                className="text-sm border border-gray-300 rounded-md px-2 py-1 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 cursor-pointer"
              >
                {PAGE_SIZE_OPTIONS.map((size) => (
                  <option key={size} value={size}>
                    {size}
                  </option>
                ))}
              </select>
            </div>
          )}
        </div>
        {totalPages > 1 && (
          <div className="flex gap-1">
            <button
              onClick={() => onPageChange(currentPage - 1)}
              disabled={currentPage === 1}
              className="px-3 py-1 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
            >
              Previous
            </button>
            {Array.from({ length: totalPages }, (_, i) => i + 1).map((page) => (
              <button
                key={page}
                onClick={() => onPageChange(page)}
                className={`px-3 py-1 text-sm rounded-md cursor-pointer ${
                  page === currentPage
                    ? 'bg-blue-600 text-white'
                    : 'border border-gray-300 hover:bg-gray-50'
                }`}
              >
                {page}
              </button>
            ))}
            <button
              onClick={() => onPageChange(currentPage + 1)}
              disabled={currentPage === totalPages}
              className="px-3 py-1 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
            >
              Next
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
