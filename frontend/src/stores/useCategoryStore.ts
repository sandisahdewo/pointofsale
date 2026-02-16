'use client';

import { create } from 'zustand';
import { apiClient, PaginatedApiResponse } from '@/lib/api';

interface CategoryApi {
  id: number;
  name: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Category {
  id: number;
  name: string;
  description: string;
  createdAt?: string;
  updatedAt?: string;
}

interface CategoryQueryParams {
  page?: number;
  pageSize?: number;
  search?: string;
  sortBy?: string;
  sortDir?: string;
}

interface CategoryInput {
  name: string;
  description?: string;
}

interface CategoryState {
  categories: Category[];
  fetchCategories: (params?: CategoryQueryParams) => Promise<PaginatedApiResponse<Category>>;
  fetchAllCategories: () => Promise<Category[]>;
  createCategory: (input: CategoryInput) => Promise<Category>;
  updateCategory: (id: number, input: CategoryInput) => Promise<Category>;
  deleteCategory: (id: number) => Promise<void>;
}

function normalizeCategory(category: CategoryApi): Category {
  return {
    ...category,
    description: category.description ?? '',
  };
}

function buildQuery(params: CategoryQueryParams = {}): string {
  const query = new URLSearchParams();
  if (params.page) query.set('page', String(params.page));
  if (params.pageSize) query.set('pageSize', String(params.pageSize));
  if (params.search) query.set('search', params.search);
  if (params.sortBy) query.set('sortBy', params.sortBy);
  if (params.sortDir) query.set('sortDir', params.sortDir);
  const qs = query.toString();
  return qs ? `?${qs}` : '';
}

export const useCategoryStore = create<CategoryState>((set, get) => ({
  categories: [],

  fetchCategories: async (params = {}) => {
    const response = await apiClient<CategoryApi[]>(`/api/v1/categories${buildQuery(params)}`);
    const paginated = response as unknown as PaginatedApiResponse<CategoryApi>;
    const data = paginated.data.map(normalizeCategory);
    set({ categories: data });

    return {
      data,
      meta: paginated.meta,
    };
  },

  fetchAllCategories: async () => {
    const pageSize = 100;
    const allCategories: Category[] = [];
    let page = 1;
    let totalPages = 1;

    while (page <= totalPages) {
      const response = await get().fetchCategories({
        page,
        pageSize,
        sortBy: 'name',
        sortDir: 'asc',
      });
      allCategories.push(...response.data);
      totalPages = response.meta.totalPages || 0;
      if (totalPages === 0) break;
      page += 1;
    }

    set({ categories: allCategories });
    return allCategories;
  },

  createCategory: async (input) => {
    const response = await apiClient<CategoryApi>('/api/v1/categories', {
      method: 'POST',
      body: JSON.stringify({
        name: input.name,
        description: input.description ?? '',
      }),
    });

    const category = normalizeCategory(response.data);
    set((state) => ({ categories: [...state.categories, category] }));
    return category;
  },

  updateCategory: async (id, input) => {
    const response = await apiClient<CategoryApi>(`/api/v1/categories/${id}`, {
      method: 'PUT',
      body: JSON.stringify({
        name: input.name,
        description: input.description ?? '',
      }),
    });

    const category = normalizeCategory(response.data);
    set((state) => ({
      categories: state.categories.map((item) => (item.id === id ? category : item)),
    }));

    return category;
  },

  deleteCategory: async (id) => {
    await apiClient(`/api/v1/categories/${id}`, {
      method: 'DELETE',
    });

    set((state) => ({
      categories: state.categories.filter((item) => item.id !== id),
    }));
  },
}));
