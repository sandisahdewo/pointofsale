'use client';

import React from 'react';
import { useSidebarStore } from '@/stores/useSidebarStore';
import SidebarMenu, { MenuItem } from '@/components/ui/SidebarMenu';

const menuItems: MenuItem[] = [
  {
    label: 'Master Data',
    children: [
      { label: 'Product', href: '/master/product' },
      { label: 'Category', href: '/master/category' },
      { label: 'Supplier', href: '/master/supplier' },
      { label: 'Rack', href: '/master/rack' },
    ],
  },
  {
    label: 'Transaction',
    children: [
      { label: 'Sales', href: '/transaction/sales' },
      { label: 'Purchase', href: '/transaction/purchase' },
    ],
  },
  {
    label: 'Report',
    children: [
      { label: 'Sales Report', href: '/report/sales' },
      { label: 'Purchase Report', href: '/report/purchase' },
    ],
  },
  {
    label: 'Settings',
    children: [
      { label: 'Users', href: '/settings/users' },
      { label: 'Roles & Permissions', href: '/settings/roles' },
    ],
  },
];

export default function Sidebar() {
  const { isOpen } = useSidebarStore();

  return (
    <aside
      className={`fixed top-16 left-0 bottom-0 bg-white border-r border-gray-200 overflow-y-auto transition-all duration-300 z-30 ${
        isOpen ? 'w-60' : 'w-0 overflow-hidden'
      }`}
    >
      <div className="p-4 w-60">
        <SidebarMenu items={menuItems} />
      </div>
    </aside>
  );
}
