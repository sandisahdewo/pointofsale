'use client';

import { create } from 'zustand';
import { apiClient, PaginatedApiResponse } from '@/lib/api';

interface RackApi {
  id: number;
  name: string;
  code: string;
  location: string;
  capacity: number;
  description?: string;
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Rack {
  id: number;
  name: string;
  code: string;
  location: string;
  capacity: number;
  description: string;
  active: boolean;
  createdAt?: string;
  updatedAt?: string;
}

interface RackQueryParams {
  page?: number;
  pageSize?: number;
  search?: string;
  sortBy?: string;
  sortDir?: string;
  active?: boolean;
}

interface RackInput {
  name: string;
  code: string;
  location: string;
  capacity: number;
  description?: string;
  active?: boolean;
}

interface RackState {
  racks: Rack[];
  fetchRacks: (params?: RackQueryParams) => Promise<PaginatedApiResponse<Rack>>;
  fetchAllRacks: (params?: Omit<RackQueryParams, 'page' | 'pageSize'>) => Promise<Rack[]>;
  createRack: (input: RackInput) => Promise<Rack>;
  updateRack: (id: number, input: RackInput) => Promise<Rack>;
  deleteRack: (id: number) => Promise<void>;
  getActiveRacks: () => Rack[];
  isCodeUnique: (code: string, excludeId?: number) => boolean;
}

function normalizeRack(rack: RackApi): Rack {
  return {
    ...rack,
    description: rack.description ?? '',
  };
}

function buildQuery(params: RackQueryParams = {}): string {
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

export const useRackStore = create<RackState>((set, get) => ({
  racks: [],

  fetchRacks: async (params = {}) => {
    const response = await apiClient<RackApi[]>(`/api/v1/racks${buildQuery(params)}`);
    const paginated = response as unknown as PaginatedApiResponse<RackApi>;
    const data = paginated.data.map(normalizeRack);
    set({ racks: data });

    return {
      data,
      meta: paginated.meta,
    };
  },

  fetchAllRacks: async (params = {}) => {
    const pageSize = 100;
    const allRacks: Rack[] = [];
    let page = 1;
    let totalPages = 1;

    while (page <= totalPages) {
      const response = await get().fetchRacks({
        ...params,
        page,
        pageSize,
        sortBy: params.sortBy ?? 'name',
        sortDir: params.sortDir ?? 'asc',
      });
      allRacks.push(...response.data);
      totalPages = response.meta.totalPages || 0;
      if (totalPages === 0) break;
      page += 1;
    }

    set({ racks: allRacks });
    return allRacks;
  },

  createRack: async (input) => {
    const response = await apiClient<RackApi>('/api/v1/racks', {
      method: 'POST',
      body: JSON.stringify({
        name: input.name,
        code: input.code,
        location: input.location,
        capacity: input.capacity,
        description: input.description ?? '',
      }),
    });

    const rack = normalizeRack(response.data);
    set((state) => ({ racks: [...state.racks, rack] }));
    return rack;
  },

  updateRack: async (id, input) => {
    const response = await apiClient<RackApi>(`/api/v1/racks/${id}`, {
      method: 'PUT',
      body: JSON.stringify({
        name: input.name,
        code: input.code,
        location: input.location,
        capacity: input.capacity,
        description: input.description ?? '',
        active: input.active,
      }),
    });

    const rack = normalizeRack(response.data);
    set((state) => ({
      racks: state.racks.map((item) => (item.id === id ? rack : item)),
    }));

    return rack;
  },

  deleteRack: async (id) => {
    await apiClient(`/api/v1/racks/${id}`, {
      method: 'DELETE',
    });

    set((state) => ({
      racks: state.racks.filter((item) => item.id !== id),
    }));
  },

  getActiveRacks: () => get().racks.filter((rack) => rack.active),

  isCodeUnique: (code, excludeId) => {
    const normalizedCode = code.trim().toLowerCase();
    return !get().racks.some(
      (rack) => rack.code.toLowerCase() === normalizedCode && rack.id !== excludeId,
    );
  },
}));
