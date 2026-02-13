'use client';

import React from 'react';

type DatePickerType = 'date' | 'datetime';

interface DatePickerProps {
  label?: string;
  value: string;
  onChange: (value: string) => void;
  error?: string;
  type?: DatePickerType;
  required?: boolean;
  className?: string;
  disabled?: boolean;
  id?: string;
}

export default function DatePicker({
  label,
  value,
  onChange,
  error,
  type = 'date',
  required = false,
  className = '',
  disabled = false,
  id,
}: DatePickerProps) {
  const inputId = id || label?.toLowerCase().replace(/\s+/g, '-');
  const inputType = type === 'datetime' ? 'datetime-local' : 'date';

  return (
    <div className="w-full">
      {label && (
        <label
          htmlFor={inputId}
          className="block text-sm font-medium text-gray-700 mb-1"
        >
          {label}
          {required && <span className="text-red-500 ml-1">*</span>}
        </label>
      )}
      <input
        id={inputId}
        type={inputType}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
        required={required}
        className={`w-full rounded-md border px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:opacity-50 disabled:cursor-not-allowed ${
          error
            ? 'border-red-500 focus:ring-red-500 focus:border-red-500'
            : 'border-gray-300'
        } ${className}`}
      />
      {error && <p className="mt-1 text-sm text-red-600">{error}</p>}
    </div>
  );
}
