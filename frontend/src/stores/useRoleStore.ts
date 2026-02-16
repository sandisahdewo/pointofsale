'use client';

import { create } from 'zustand';
import { apiClient, PaginationMeta } from '@/lib/api';

export interface Role {
  id: number;
  name: string;
  description: string;
  isSystem: boolean;
  userCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface Permission {
  id: number;
  module: string;
  feature: string;
  actions: string[];
}

export interface RolePermissionDetail {
  permissionId: number;
  module: string;
  feature: string;
  availableActions: string[];
  grantedActions: string[];
}

export interface RolePermissionsData {
  roleId: number;
  roleName: string;
  isSystem: boolean;
  permissions: RolePermissionDetail[];
}

interface RoleState {
  roles: Role[];
  fetchRoles: (params?: {
    page?: number;
    pageSize?: number;
    search?: string;
    sortBy?: string;
    sortDir?: string;
  }) => Promise<{ data: Role[]; meta: PaginationMeta }>;
  fetchAllRoles: () => Promise<Role[]>;
  getRole: (id: number) => Promise<Role>;
  createRole: (input: { name: string; description: string }) => Promise<Role>;
  updateRole: (id: number, input: { name: string; description: string }) => Promise<Role>;
  deleteRole: (id: number) => Promise<void>;
  fetchPermissions: () => Promise<Permission[]>;
  fetchRolePermissions: (roleId: number) => Promise<RolePermissionsData>;
  updateRolePermissions: (
    roleId: number,
    permissions: { permissionId: number; actions: string[] }[]
  ) => Promise<RolePermissionsData>;
}

export const useRoleStore = create<RoleState>((set) => ({
  roles: [],

  fetchRoles: async (params = {}) => {
    const query = new URLSearchParams();
    if (params.page) query.set('page', String(params.page));
    if (params.pageSize) query.set('pageSize', String(params.pageSize));
    if (params.search) query.set('search', params.search);
    if (params.sortBy) query.set('sortBy', params.sortBy);
    if (params.sortDir) query.set('sortDir', params.sortDir);
    const qs = query.toString();
    const response = await apiClient<Role[]>(`/api/v1/roles${qs ? `?${qs}` : ''}`);
    return response as unknown as { data: Role[]; meta: PaginationMeta };
  },

  fetchAllRoles: async () => {
    const response = await apiClient<Role[]>('/api/v1/roles?pageSize=100');
    const roles = (response as unknown as { data: Role[] }).data;
    set({ roles });
    return roles;
  },

  getRole: async (id) => {
    const response = await apiClient<Role>(`/api/v1/roles/${id}`);
    return response.data;
  },

  createRole: async (input) => {
    const response = await apiClient<Role>('/api/v1/roles', {
      method: 'POST',
      body: JSON.stringify(input),
    });
    return response.data;
  },

  updateRole: async (id, input) => {
    const response = await apiClient<Role>(`/api/v1/roles/${id}`, {
      method: 'PUT',
      body: JSON.stringify(input),
    });
    return response.data;
  },

  deleteRole: async (id) => {
    await apiClient(`/api/v1/roles/${id}`, {
      method: 'DELETE',
    });
  },

  fetchPermissions: async () => {
    const response = await apiClient<Permission[]>('/api/v1/permissions');
    return response.data;
  },

  fetchRolePermissions: async (roleId) => {
    const response = await apiClient<RolePermissionsData>(
      `/api/v1/roles/${roleId}/permissions`
    );
    return response.data;
  },

  updateRolePermissions: async (roleId, permissions) => {
    const response = await apiClient<RolePermissionsData>(
      `/api/v1/roles/${roleId}/permissions`,
      {
        method: 'PUT',
        body: JSON.stringify({ permissions }),
      }
    );
    return response.data;
  },
}));
