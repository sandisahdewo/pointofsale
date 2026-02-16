'use client';

import { create } from 'zustand';
import { apiClient, PaginatedApiResponse } from '@/lib/api';

interface BankAccountApi {
  id: string;
  accountName: string;
  accountNumber: string;
}

interface SupplierApi {
  id: number;
  name: string;
  address: string;
  phone?: string;
  email?: string;
  website?: string;
  bankAccounts?: BankAccountApi[];
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface BankAccount {
  id: string;
  accountName: string;
  accountNumber: string;
}

export interface Supplier {
  id: number;
  name: string;
  address: string;
  phone: string;
  email: string;
  website: string;
  bankAccounts: BankAccount[];
  active: boolean;
  createdAt?: string;
  updatedAt?: string;
}

interface SupplierQueryParams {
  page?: number;
  pageSize?: number;
  search?: string;
  sortBy?: string;
  sortDir?: string;
  active?: boolean;
}

interface SupplierInput {
  name: string;
  address: string;
  phone?: string;
  email?: string;
  website?: string;
  active?: boolean;
  bankAccounts?: Array<{
    accountName: string;
    accountNumber: string;
  }>;
}

interface SupplierState {
  suppliers: Supplier[];
  fetchSuppliers: (params?: SupplierQueryParams) => Promise<PaginatedApiResponse<Supplier>>;
  fetchAllSuppliers: (params?: Omit<SupplierQueryParams, 'page' | 'pageSize'>) => Promise<Supplier[]>;
  createSupplier: (input: SupplierInput) => Promise<Supplier>;
  updateSupplier: (id: number, input: SupplierInput) => Promise<Supplier>;
  deleteSupplier: (id: number) => Promise<void>;
  getActiveSuppliers: () => Supplier[];
}

function normalizeSupplier(supplier: SupplierApi): Supplier {
  return {
    ...supplier,
    phone: supplier.phone ?? '',
    email: supplier.email ?? '',
    website: supplier.website ?? '',
    bankAccounts: supplier.bankAccounts ?? [],
  };
}

function buildQuery(params: SupplierQueryParams = {}): string {
  const query = new URLSearchParams();
  if (params.page) query.set('page', String(params.page));
  if (params.pageSize) query.set('pageSize', String(params.pageSize));
  if (params.search) query.set('search', params.search);
  if (params.sortBy) query.set('sortBy', params.sortBy);
  if (params.sortDir) query.set('sortDir', params.sortDir);
  if (params.active !== undefined) query.set('active', String(params.active));
  const qs = query.toString();
  return qs ? `?${qs}` : '';
}

function normalizeBankAccounts(input: SupplierInput['bankAccounts']): SupplierInput['bankAccounts'] {
  if (!input) return undefined;
  return input
    .filter((item) => item.accountName.trim() && item.accountNumber.trim())
    .map((item) => ({
      accountName: item.accountName.trim(),
      accountNumber: item.accountNumber.trim(),
    }));
}

export const useSupplierStore = create<SupplierState>((set, get) => ({
  suppliers: [],

  fetchSuppliers: async (params = {}) => {
    const response = await apiClient<SupplierApi[]>(`/api/v1/suppliers${buildQuery(params)}`);
    const paginated = response as unknown as PaginatedApiResponse<SupplierApi>;
    const data = paginated.data.map(normalizeSupplier);
    set({ suppliers: data });

    return {
      data,
      meta: paginated.meta,
    };
  },

  fetchAllSuppliers: async (params = {}) => {
    const pageSize = 100;
    const allSuppliers: Supplier[] = [];
    let page = 1;
    let totalPages = 1;

    while (page <= totalPages) {
      const response = await get().fetchSuppliers({
        ...params,
        page,
        pageSize,
        sortBy: params.sortBy ?? 'name',
        sortDir: params.sortDir ?? 'asc',
      });
      allSuppliers.push(...response.data);
      totalPages = response.meta.totalPages || 0;
      if (totalPages === 0) break;
      page += 1;
    }

    set({ suppliers: allSuppliers });
    return allSuppliers;
  },

  createSupplier: async (input) => {
    const response = await apiClient<SupplierApi>('/api/v1/suppliers', {
      method: 'POST',
      body: JSON.stringify({
        name: input.name,
        address: input.address,
        phone: input.phone ?? '',
        email: input.email ?? '',
        website: input.website ?? '',
        bankAccounts: normalizeBankAccounts(input.bankAccounts) ?? [],
      }),
    });

    const supplier = normalizeSupplier(response.data);
    set((state) => ({ suppliers: [...state.suppliers, supplier] }));
    return supplier;
  },

  updateSupplier: async (id, input) => {
    const response = await apiClient<SupplierApi>(`/api/v1/suppliers/${id}`, {
      method: 'PUT',
      body: JSON.stringify({
        name: input.name,
        address: input.address,
        phone: input.phone ?? '',
        email: input.email ?? '',
        website: input.website ?? '',
        active: input.active,
        bankAccounts: normalizeBankAccounts(input.bankAccounts),
      }),
    });

    const supplier = normalizeSupplier(response.data);
    set((state) => ({
      suppliers: state.suppliers.map((item) => (item.id === id ? supplier : item)),
    }));

    return supplier;
  },

  deleteSupplier: async (id) => {
    await apiClient(`/api/v1/suppliers/${id}`, {
      method: 'DELETE',
    });

    set((state) => ({
      suppliers: state.suppliers.filter((item) => item.id !== id),
    }));
  },

  getActiveSuppliers: () => get().suppliers.filter((supplier) => supplier.active),
}));
