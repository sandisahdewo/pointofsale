'use client';

import { create } from 'zustand';
import { apiClient, PaginatedApiResponse } from '@/lib/api';

export interface UserRole {
  id: number;
  name: string;
}

export interface User {
  id: number;
  name: string;
  email: string;
  phone: string;
  address: string;
  profilePicture: string | null;
  roles: UserRole[];
  status: 'active' | 'pending' | 'inactive';
  isSuperAdmin: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateUserInput {
  name: string;
  email: string;
  phone?: string;
  address?: string;
  roleIds?: number[];
  profilePicture?: string | null;
}

export interface UpdateUserInput {
  name: string;
  email: string;
  phone?: string;
  address?: string;
  roleIds?: number[];
  status?: string;
  profilePicture?: string | null;
}

interface UserState {
  // API functions (all async, throw on error)
  fetchUsers: (params: {
    page?: number;
    pageSize?: number;
    search?: string;
    sortBy?: string;
    sortDir?: string;
    status?: string;
  }) => Promise<PaginatedApiResponse<User>>;
  getUser: (id: number) => Promise<User>;
  createUser: (input: CreateUserInput) => Promise<User>;
  updateUser: (id: number, input: UpdateUserInput) => Promise<User>;
  deleteUser: (id: number) => Promise<void>;
  approveUser: (id: number) => Promise<User>;
  rejectUser: (id: number) => Promise<void>;
}

export const useUserStore = create<UserState>(() => ({
  fetchUsers: async (params) => {
    const query = new URLSearchParams();
    if (params.page) query.set('page', String(params.page));
    if (params.pageSize) query.set('pageSize', String(params.pageSize));
    if (params.search) query.set('search', params.search);
    if (params.sortBy) query.set('sortBy', params.sortBy);
    if (params.sortDir) query.set('sortDir', params.sortDir);
    if (params.status) query.set('status', params.status);
    const qs = query.toString();
    const response = await apiClient<User[]>(`/api/v1/users${qs ? `?${qs}` : ''}`);
    return response as unknown as PaginatedApiResponse<User>;
  },
  getUser: async (id) => {
    const response = await apiClient<User>(`/api/v1/users/${id}`);
    return response.data;
  },
  createUser: async (input) => {
    const response = await apiClient<User>('/api/v1/users', {
      method: 'POST',
      body: JSON.stringify({
        name: input.name,
        email: input.email,
        phone: input.phone,
        address: input.address,
        roleIds: input.roleIds,
      }),
    });
    return response.data;
  },
  updateUser: async (id, input) => {
    const response = await apiClient<User>(`/api/v1/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify({
        name: input.name,
        email: input.email,
        phone: input.phone,
        address: input.address,
        roleIds: input.roleIds,
        status: input.status,
      }),
    });
    return response.data;
  },
  deleteUser: async (id) => {
    await apiClient(`/api/v1/users/${id}`, {
      method: 'DELETE',
    });
  },
  approveUser: async (id) => {
    const response = await apiClient<User>(`/api/v1/users/${id}/approve`, {
      method: 'PATCH',
    });
    return response.data;
  },
  rejectUser: async (id) => {
    await apiClient(`/api/v1/users/${id}/reject`, {
      method: 'DELETE',
    });
  },
}));
