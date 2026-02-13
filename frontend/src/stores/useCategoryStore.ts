'use client';

import { create } from 'zustand';
import { initialCategories } from '@/data/categories';

export interface Category {
  id: number;
  name: string;
  description: string;
}

interface CategoryState {
  categories: Category[];
  addCategory: (category: Omit<Category, 'id'>) => void;
  updateCategory: (id: number, category: Omit<Category, 'id'>) => void;
  deleteCategory: (id: number) => void;
}

export const useCategoryStore = create<CategoryState>((set) => ({
  categories: initialCategories,
  addCategory: (category) =>
    set((state) => {
      const maxId = state.categories.reduce((max, c) => Math.max(max, c.id), 0);
      return { categories: [...state.categories, { ...category, id: maxId + 1 }] };
    }),
  updateCategory: (id, category) =>
    set((state) => ({
      categories: state.categories.map((c) =>
        c.id === id ? { ...c, ...category } : c
      ),
    })),
  deleteCategory: (id) =>
    set((state) => ({
      categories: state.categories.filter((c) => c.id !== id),
    })),
}));
