import { Permission } from '@/stores/useRoleStore';

export const initialPermissions: Permission[] = [
  {
    id: 1,
    module: 'Master Data',
    feature: 'Product',
    actions: ['read', 'create', 'update', 'delete', 'export'],
  },
  {
    id: 2,
    module: 'Master Data',
    feature: 'Category',
    actions: ['read', 'create', 'update', 'delete'],
  },
  {
    id: 3,
    module: 'Master Data',
    feature: 'Supplier',
    actions: ['read', 'create', 'update', 'delete', 'export'],
  },
  {
    id: 4,
    module: 'Transaction',
    feature: 'Sales',
    actions: ['read', 'create', 'update', 'delete', 'export'],
  },
  {
    id: 5,
    module: 'Transaction',
    feature: 'Purchase',
    actions: ['read', 'create', 'update', 'delete', 'export'],
  },
  {
    id: 6,
    module: 'Report',
    feature: 'Sales Report',
    actions: ['read', 'export'],
  },
  {
    id: 7,
    module: 'Report',
    feature: 'Purchase Report',
    actions: ['read', 'export'],
  },
  {
    id: 8,
    module: 'Settings',
    feature: 'Users',
    actions: ['read', 'create', 'update', 'delete'],
  },
  {
    id: 9,
    module: 'Settings',
    feature: 'Roles & Permissions',
    actions: ['read', 'create', 'update', 'delete'],
  },
];
