'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

export interface MenuItem {
  label: string;
  href?: string;
  children?: MenuItem[];
}

interface SidebarMenuProps {
  items: MenuItem[];
}

function MenuItemComponent({ item, depth = 0 }: { item: MenuItem; depth?: number }) {
  const pathname = usePathname();
  const [isExpanded, setIsExpanded] = useState(true);
  const hasChildren = item.children && item.children.length > 0;
  const isActive = item.href === pathname;

  return (
    <div>
      {hasChildren ? (
        <button
          onClick={() => setIsExpanded(!isExpanded)}
          className={`w-full flex items-center justify-between px-3 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-md cursor-pointer`}
          style={{ paddingLeft: `${12 + depth * 16}px` }}
        >
          <span className="font-medium">{item.label}</span>
          <svg
            className={`w-4 h-4 transition-transform ${isExpanded ? 'rotate-90' : ''}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
          </svg>
        </button>
      ) : (
        <Link
          href={item.href || '#'}
          className={`block px-3 py-2 text-sm rounded-md ${
            isActive
              ? 'bg-blue-50 text-blue-700 font-medium'
              : 'text-gray-600 hover:bg-gray-100'
          }`}
          style={{ paddingLeft: `${12 + depth * 16}px` }}
        >
          {item.label}
        </Link>
      )}
      {hasChildren && isExpanded && (
        <div>
          {item.children!.map((child, index) => (
            <MenuItemComponent key={index} item={child} depth={depth + 1} />
          ))}
        </div>
      )}
    </div>
  );
}

export default function SidebarMenu({ items }: SidebarMenuProps) {
  return (
    <nav className="space-y-1">
      {items.map((item, index) => (
        <MenuItemComponent key={index} item={item} />
      ))}
    </nav>
  );
}
