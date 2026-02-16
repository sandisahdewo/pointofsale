// Legacy data - no longer used with API integration
interface RolePermission {
  roleId: number;
  permissionId: number;
  actions: string[];
}

export const initialRolePermissions: RolePermission[] = [
  // Manager (roleId: 2)
  { roleId: 2, permissionId: 1, actions: ['read', 'create', 'update', 'delete', 'export'] }, // Product
  { roleId: 2, permissionId: 2, actions: ['read', 'create', 'update', 'delete'] },           // Category
  { roleId: 2, permissionId: 3, actions: ['read', 'create', 'update', 'delete', 'export'] }, // Supplier
  { roleId: 2, permissionId: 4, actions: ['read', 'create', 'update', 'export'] },           // Sales
  { roleId: 2, permissionId: 5, actions: ['read', 'create', 'update', 'export'] },           // Purchase
  { roleId: 2, permissionId: 6, actions: ['read', 'export'] },                               // Sales Report
  { roleId: 2, permissionId: 7, actions: ['read', 'export'] },                               // Purchase Report

  // Cashier (roleId: 3)
  { roleId: 3, permissionId: 4, actions: ['read', 'create'] },  // Sales
  { roleId: 3, permissionId: 6, actions: ['read'] },            // Sales Report

  // Accountant (roleId: 4)
  { roleId: 4, permissionId: 4, actions: ['read', 'export'] },  // Sales
  { roleId: 4, permissionId: 5, actions: ['read', 'export'] },  // Purchase
  { roleId: 4, permissionId: 6, actions: ['read', 'export'] },  // Sales Report
  { roleId: 4, permissionId: 7, actions: ['read', 'export'] },  // Purchase Report

  // Warehouse (roleId: 5)
  { roleId: 5, permissionId: 1, actions: ['read', 'update'] },          // Product
  { roleId: 5, permissionId: 3, actions: ['read'] },                    // Supplier
  { roleId: 5, permissionId: 5, actions: ['read', 'create', 'update'] }, // Purchase
];
