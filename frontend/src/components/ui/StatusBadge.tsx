'use client';

import React from 'react';

type StatusColor = 'green' | 'blue' | 'yellow' | 'amber' | 'gray' | 'red';
type StatusSize = 'sm' | 'md';

interface StatusBadgeProps {
  status: string;
  colorMap?: Record<string, StatusColor>;
  size?: StatusSize;
  className?: string;
}

const colorClasses: Record<StatusColor, string> = {
  green: 'bg-green-100 text-green-800',
  blue: 'bg-blue-100 text-blue-800',
  yellow: 'bg-yellow-100 text-yellow-800',
  amber: 'bg-amber-100 text-amber-800',
  gray: 'bg-gray-100 text-gray-800',
  red: 'bg-red-100 text-red-800',
};

const sizeClasses: Record<StatusSize, string> = {
  sm: 'px-2 py-0.5 text-xs',
  md: 'px-2.5 py-1 text-sm',
};

export default function StatusBadge({
  status,
  colorMap = {},
  size = 'sm',
  className = '',
}: StatusBadgeProps) {
  const color = colorMap[status] || 'gray';

  return (
    <span
      className={`inline-flex items-center rounded-full font-medium ${colorClasses[color]} ${sizeClasses[size]} ${className}`}
    >
      {status}
    </span>
  );
}
