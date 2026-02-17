'use client';

import { create } from 'zustand';
import { initialProducts } from '@/data/products';
import { apiClient, PaginatedApiResponse } from '@/lib/api';

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

interface ProductImageApi {
  id?: number;
  imageUrl: string;
  sortOrder: number;
}

interface ProductUnitApi {
  id: number;
  name: string;
  conversionFactor: number;
  convertsToId?: number | null;
  toBaseUnit: number;
  isBase: boolean;
}

interface VariantAttributeApi {
  attributeName: string;
  attributeValue: string;
}

interface VariantImageApi {
  id?: number;
  imageUrl: string;
  sortOrder: number;
}

interface VariantPricingTierApi {
  id?: number;
  minQty: number;
  value: number;
}

interface SupplierSummaryApi {
  id: number;
  name: string;
}

interface RackSummaryApi {
  id: number;
  name: string;
}

interface ProductVariantApi {
  id: string;
  sku?: string;
  barcode?: string;
  currentStock?: number;
  attributes?: VariantAttributeApi[];
  images?: VariantImageApi[];
  pricingTiers?: VariantPricingTierApi[];
  racks?: RackSummaryApi[];
}

interface ProductApi {
  id: number;
  name: string;
  description?: string;
  categoryId: number;
  priceSetting: PriceSetting;
  markupType?: MarkupType | null;
  hasVariants: boolean;
  status: 'active' | 'inactive';
  images?: ProductImageApi[];
  suppliers?: SupplierSummaryApi[];
  units?: ProductUnitApi[];
  variants?: ProductVariantApi[];
}

interface ProductListApi extends ProductApi {
  variantCount?: number;
}

interface ProductQueryParams {
  page?: number;
  pageSize?: number;
  search?: string;
  sortBy?: string;
  sortDir?: string;
  status?: 'active' | 'inactive';
  categoryId?: number;
  supplierId?: number;
}

interface CreateProductImagePayload {
  imageUrl: string;
  sortOrder: number;
}

interface CreateProductUnitPayload {
  name: string;
  isBase: boolean;
  conversionFactor?: number;
  convertsToName?: string;
}

interface CreateVariantAttributePayload {
  attributeName: string;
  attributeValue: string;
}

interface CreateVariantImagePayload {
  imageUrl: string;
  sortOrder: number;
}

interface CreateVariantPayload {
  id?: string;
  sku: string;
  barcode: string;
  attributes: CreateVariantAttributePayload[];
  images: CreateVariantImagePayload[];
  pricingTiers: PricingTier[];
  rackIds: number[];
}

interface CreateProductPayload {
  name: string;
  description: string;
  categoryId: number;
  priceSetting: PriceSetting;
  markupType?: MarkupType;
  hasVariants: boolean;
  status: 'active' | 'inactive';
  supplierIds: number[];
  images: CreateProductImagePayload[];
  units: CreateProductUnitPayload[];
  variants: CreateVariantPayload[];
}

interface ProductState {
  products: Product[];
  addProduct: (product: Omit<Product, 'id'>) => void;
  updateProduct: (id: number, product: Partial<Product>) => void;
  deleteProduct: (id: number) => void;
  getProduct: (id: number) => Product | undefined;
  fetchAllProducts: () => Promise<Product[]>;
  fetchProductById: (id: number) => Promise<Product>;
  createProduct: (product: Omit<Product, 'id'>) => Promise<Product>;
  updateProductRemote: (id: number, product: Omit<Product, 'id'>) => Promise<Product>;
  deleteProductRemote: (id: number) => Promise<void>;
}

function buildProductQuery(params: ProductQueryParams = {}): string {
  const query = new URLSearchParams();
  if (params.page) query.set('page', String(params.page));
  if (params.pageSize) query.set('pageSize', String(params.pageSize));
  if (params.search) query.set('search', params.search);
  if (params.sortBy) query.set('sortBy', params.sortBy);
  if (params.sortDir) query.set('sortDir', params.sortDir);
  if (params.status) query.set('status', params.status);
  if (params.categoryId) query.set('categoryId', String(params.categoryId));
  if (params.supplierId) query.set('supplierId', String(params.supplierId));
  const qs = query.toString();
  return qs ? `?${qs}` : '';
}

function toAttributesRecord(attributes: VariantAttributeApi[] = []): Record<string, string> {
  return attributes.reduce<Record<string, string>>((acc, item) => {
    const key = item.attributeName?.trim();
    const value = item.attributeValue?.trim();
    if (key && value) {
      acc[key] = value;
    }
    return acc;
  }, {});
}

function buildVariantAttributes(variants: ProductVariant[]): VariantAttribute[] {
  const map = new Map<string, Set<string>>();

  variants.forEach((variant) => {
    Object.entries(variant.attributes).forEach(([key, value]) => {
      const name = key.trim();
      const val = value.trim();
      if (!name || !val) return;
      if (!map.has(name)) {
        map.set(name, new Set());
      }
      map.get(name)?.add(val);
    });
  });

  return Array.from(map.entries())
    .map(([name, values]) => ({
      name,
      values: Array.from(values.values()).sort((a, b) => a.localeCompare(b)),
    }))
    .sort((a, b) => a.name.localeCompare(b.name));
}

function normalizeProduct(api: ProductApi | ProductListApi): Product {
  const units: ProductUnit[] = (api.units ?? []).map((unit) => ({
    id: String(unit.id),
    name: unit.name,
    conversionFactor: Number(unit.conversionFactor ?? 1),
    convertsTo: unit.convertsToId == null ? null : String(unit.convertsToId),
    toBaseUnit: Number(unit.toBaseUnit ?? 1),
    isBase: Boolean(unit.isBase),
  }));

  const variants: ProductVariant[] = (api.variants ?? []).map((variant) => ({
    id: variant.id,
    sku: variant.sku ?? '',
    barcode: variant.barcode ?? '',
    attributes: toAttributesRecord(variant.attributes),
    pricingTiers: (variant.pricingTiers ?? []).map((tier) => ({
      minQty: tier.minQty,
      value: Number(tier.value),
    })),
    images: (variant.images ?? [])
      .slice()
      .sort((a, b) => a.sortOrder - b.sortOrder)
      .map((image) => image.imageUrl),
    rackIds: (variant.racks ?? []).map((rack) => rack.id),
    currentStock: Number(variant.currentStock ?? 0),
  }));

  const images = (api.images ?? [])
    .slice()
    .sort((a, b) => a.sortOrder - b.sortOrder)
    .map((image) => image.imageUrl);

  const markupType = api.markupType ?? undefined;

  return {
    id: api.id,
    name: api.name,
    description: api.description ?? '',
    categoryId: api.categoryId,
    images,
    priceSetting: api.priceSetting,
    markupType: markupType === null ? undefined : markupType,
    hasVariants: api.hasVariants,
    status: api.status,
    units,
    variants,
    variantAttributes: buildVariantAttributes(variants),
    supplierIds: (api.suppliers ?? []).map((supplier) => supplier.id),
  };
}

function isUuid(value: string): boolean {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value);
}

function toProductPayload(product: Omit<Product, 'id'>, includeVariantIds: boolean): CreateProductPayload {
  const unitNameById = new Map(product.units.map((unit) => [unit.id, unit.name.trim()]));

  const units: CreateProductUnitPayload[] = [];
  product.units.forEach((unit) => {
    const name = unit.name.trim();
    if (!name) return;

    if (unit.isBase) {
      units.push({
        name,
        isBase: true,
      });
      return;
    }

    units.push({
      name,
      isBase: false,
      conversionFactor: Number(unit.conversionFactor),
      convertsToName: unit.convertsTo ? unitNameById.get(unit.convertsTo) ?? '' : '',
    });
  });

  const variants: CreateVariantPayload[] = product.variants.map((variant) => {
    const payload: CreateVariantPayload = {
      sku: variant.sku.trim(),
      barcode: variant.barcode.trim(),
      attributes: Object.entries(variant.attributes)
        .map(([attributeName, attributeValue]) => ({
          attributeName: attributeName.trim(),
          attributeValue: attributeValue.trim(),
        }))
        .filter((item) => item.attributeName && item.attributeValue),
      images: variant.images
        .map((imageUrl, index) => ({
          imageUrl: imageUrl.trim(),
          sortOrder: index,
        }))
        .filter((item) => item.imageUrl),
      pricingTiers: variant.pricingTiers.map((tier) => ({
        minQty: Number(tier.minQty),
        value: Number(tier.value),
      })),
      rackIds: Array.from(new Set(variant.rackIds)),
    };

    if (includeVariantIds && isUuid(variant.id)) {
      payload.id = variant.id;
    }

    return payload;
  });

  const payload: CreateProductPayload = {
    name: product.name.trim(),
    description: product.description.trim(),
    categoryId: product.categoryId,
    priceSetting: product.priceSetting,
    hasVariants: product.hasVariants,
    status: product.status,
    supplierIds: Array.from(new Set(product.supplierIds)),
    images: product.images
      .map((imageUrl, index) => ({
        imageUrl: imageUrl.trim(),
        sortOrder: index,
      }))
      .filter((item) => item.imageUrl),
    units,
    variants,
  };

  if (product.priceSetting === 'markup' && product.markupType) {
    payload.markupType = product.markupType;
  }

  return payload;
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

  fetchAllProducts: async () => {
    const allSummaries: ProductListApi[] = [];
    const pageSize = 50;
    let page = 1;
    let totalPages = 1;

    while (page <= totalPages) {
      const response = await apiClient<ProductListApi[]>(
        `/api/v1/products${buildProductQuery({ page, pageSize, sortBy: 'id', sortDir: 'asc' })}`,
      );
      const paginated = response as unknown as PaginatedApiResponse<ProductListApi>;
      allSummaries.push(...paginated.data);
      totalPages = paginated.meta.totalPages || 0;
      if (totalPages === 0) break;
      page += 1;
    }

    if (allSummaries.length === 0) {
      set({ products: [] });
      return [];
    }

    const uniqueIds = Array.from(new Set(allSummaries.map((product) => product.id)));
    const detailedProducts = await Promise.all(
      uniqueIds.map(async (id) => {
        const response = await apiClient<ProductApi>(`/api/v1/products/${id}`);
        return normalizeProduct(response.data);
      }),
    );

    set({ products: detailedProducts });
    return detailedProducts;
  },

  fetchProductById: async (id) => {
    const response = await apiClient<ProductApi>(`/api/v1/products/${id}`);
    const product = normalizeProduct(response.data);

    set((state) => {
      const exists = state.products.some((item) => item.id === id);
      if (exists) {
        return {
          products: state.products.map((item) => (item.id === id ? product : item)),
        };
      }
      return { products: [...state.products, product] };
    });

    return product;
  },

  createProduct: async (product) => {
    const response = await apiClient<ProductApi>('/api/v1/products', {
      method: 'POST',
      body: JSON.stringify(toProductPayload(product, false)),
    });

    const created = normalizeProduct(response.data);
    set((state) => ({ products: [...state.products, created] }));
    return created;
  },

  updateProductRemote: async (id, product) => {
    const response = await apiClient<ProductApi>(`/api/v1/products/${id}`, {
      method: 'PUT',
      body: JSON.stringify(toProductPayload(product, true)),
    });

    const updated = normalizeProduct(response.data);
    set((state) => ({
      products: state.products.map((item) => (item.id === id ? updated : item)),
    }));
    return updated;
  },

  deleteProductRemote: async (id) => {
    await apiClient(`/api/v1/products/${id}`, {
      method: 'DELETE',
    });

    set((state) => ({
      products: state.products.filter((item) => item.id !== id),
    }));
  },
}));
