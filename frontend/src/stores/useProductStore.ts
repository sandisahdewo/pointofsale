'use client';

import { create } from 'zustand';
import { initialProducts } from '@/data/products';

export interface ProductUnit {
  id: string;
  name: string;
  conversionFactor: number;
  convertsTo: string | null;
  toBaseUnit: number;
  isBase: boolean;
}

export interface VariantAttribute {
  name: string;
  values: string[];
}

export interface PricingTier {
  minQty: number;
  value: number;
}

export interface ProductVariant {
  id: string;
  sku: string;
  barcode: string;
  attributes: Record<string, string>;
  pricingTiers: PricingTier[];
  images: string[];
  rackIds: number[];
  currentStock: number;
}

export type PriceSetting = 'fixed' | 'markup';
export type MarkupType = 'percentage' | 'fixed_amount';

export interface Product {
  id: number;
  name: string;
  description: string;
  categoryId: number;
  images: string[];
  priceSetting: PriceSetting;
  markupType?: MarkupType;
  hasVariants: boolean;
  status: 'active' | 'inactive';
  units: ProductUnit[];
  variantAttributes: VariantAttribute[];
  variants: ProductVariant[];
  supplierIds: number[];
}

interface ProductState {
  products: Product[];
  addProduct: (product: Omit<Product, 'id'>) => void;
  updateProduct: (id: number, product: Partial<Product>) => void;
  deleteProduct: (id: number) => void;
  getProduct: (id: number) => Product | undefined;
}

export const useProductStore = create<ProductState>((set, get) => ({
  products: initialProducts,
  addProduct: (product) =>
    set((state) => {
      const maxId = state.products.reduce((max, p) => Math.max(max, p.id), 0);
      return { products: [...state.products, { ...product, id: maxId + 1 }] };
    }),
  updateProduct: (id, product) =>
    set((state) => ({
      products: state.products.map((p) =>
        p.id === id ? { ...p, ...product } : p
      ),
    })),
  deleteProduct: (id) =>
    set((state) => ({
      products: state.products.filter((p) => p.id !== id),
    })),
  getProduct: (id) => get().products.find((p) => p.id === id),
}));
