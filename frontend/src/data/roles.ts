import { Role } from '@/stores/useRoleStore';

export const initialRoles: Role[] = [
  {
    id: 1,
    name: 'Super Admin',
    description: 'Full system access. Cannot be modified or deleted.',
    isSystem: true,
    createdAt: '2025-01-01T00:00:00.000Z',
  },
  {
    id: 2,
    name: 'Manager',
    description: 'Manage products, transactions, and view reports.',
    isSystem: false,
    createdAt: '2025-01-01T00:00:00.000Z',
  },
  {
    id: 3,
    name: 'Cashier',
    description: 'Process sales transactions.',
    isSystem: false,
    createdAt: '2025-01-01T00:00:00.000Z',
  },
  {
    id: 4,
    name: 'Accountant',
    description: 'View transactions and generate reports.',
    isSystem: false,
    createdAt: '2025-01-01T00:00:00.000Z',
  },
  {
    id: 5,
    name: 'Warehouse',
    description: 'Manage product stock and purchase orders.',
    isSystem: false,
    createdAt: '2025-01-01T00:00:00.000Z',
  },
];
