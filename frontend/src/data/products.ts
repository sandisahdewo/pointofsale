import { Product } from '@/stores/useProductStore';

export const initialProducts: Product[] = [
  // 1. T-Shirt — Linear units + variants + tiered pricing
  {
    id: 1,
    name: 'T-Shirt',
    description: 'Premium cotton t-shirt available in multiple colors and sizes',
    categoryId: 2, // Clothing
    images: [],
    priceSetting: 'fixed',
    hasVariants: true,
    status: 'active',
    supplierIds: [1, 2],
    units: [
      { id: 'pcs', name: 'Pcs', conversionFactor: 1, convertsTo: null, toBaseUnit: 1, isBase: true },
      { id: 'dozen', name: 'Dozen', conversionFactor: 12, convertsTo: 'pcs', toBaseUnit: 12, isBase: false },
      { id: 'box', name: 'Box', conversionFactor: 12, convertsTo: 'dozen', toBaseUnit: 144, isBase: false },
    ],
    variantAttributes: [
      { name: 'Color', values: ['Red', 'Blue', 'Black'] },
      { name: 'Size', values: ['S', 'M', 'L', 'XL'] },
    ],
    variants: [
      {
        id: 'ts-red-s',
        sku: 'TS-RED-S',
        barcode: '8901234000101',
        attributes: { Color: 'Red', Size: 'S' },
        pricingTiers: [
          { minQty: 1, value: 75000 },
          { minQty: 12, value: 70000 },
          { minQty: 144, value: 65000 },
        ],
        images: [],
        rackIds: [1, 2],
        currentStock: 50,
      },
      {
        id: 'ts-red-m',
        sku: 'TS-RED-M',
        barcode: '8901234000102',
        attributes: { Color: 'Red', Size: 'M' },
        pricingTiers: [
          { minQty: 1, value: 75000 },
          { minQty: 12, value: 70000 },
          { minQty: 144, value: 65000 },
        ],
        images: [],
        rackIds: [1, 2],
        currentStock: 100,
      },
      {
        id: 'ts-blue-l',
        sku: 'TS-BLUE-L',
        barcode: '8901234000201',
        attributes: { Color: 'Blue', Size: 'L' },
        pricingTiers: [
          { minQty: 1, value: 80000 },
          { minQty: 12, value: 74000 },
          { minQty: 144, value: 68000 },
        ],
        images: [],
        rackIds: [1, 2],
        currentStock: 0,
      },
      {
        id: 'ts-black-xl',
        sku: 'TS-BLACK-XL',
        barcode: '8901234000301',
        attributes: { Color: 'Black', Size: 'XL' },
        pricingTiers: [
          { minQty: 1, value: 85000 },
          { minQty: 12, value: 78000 },
          { minQty: 144, value: 72000 },
        ],
        images: [],
        rackIds: [1, 2],
        currentStock: 5,
      },
    ],
  },

  // 2. Rice 5kg — Branching units + tiered pricing
  {
    id: 2,
    name: 'Rice 5kg',
    description: 'Premium quality white rice, 5kg pack',
    categoryId: 3, // Food & Beverages
    images: [],
    priceSetting: 'fixed',
    hasVariants: false,
    status: 'active',
    supplierIds: [3],
    units: [
      { id: 'kg', name: 'Kg', conversionFactor: 1, convertsTo: null, toBaseUnit: 1, isBase: true },
      { id: 'karung', name: 'Karung', conversionFactor: 50, convertsTo: 'kg', toBaseUnit: 50, isBase: false },
      { id: 'bag', name: 'Bag', conversionFactor: 25, convertsTo: 'kg', toBaseUnit: 25, isBase: false },
    ],
    variantAttributes: [],
    variants: [
      {
        id: 'rice-5kg-default',
        sku: 'RICE-5KG',
        barcode: '8901234001001',
        attributes: {},
        pricingTiers: [
          { minQty: 1, value: 65000 },
        ],
        images: [],
        rackIds: [4],
        currentStock: 200,
      },
    ],
  },

  // 3. Notebook A5 — No variants + markup pricing
  {
    id: 3,
    name: 'Notebook A5',
    description: 'Lined notebook, A5 size, 100 pages',
    categoryId: 5, // Stationery
    images: [],
    priceSetting: 'markup',
    markupType: 'percentage',
    hasVariants: false,
    status: 'active',
    supplierIds: [2],
    units: [
      { id: 'pcs', name: 'Pcs', conversionFactor: 1, convertsTo: null, toBaseUnit: 1, isBase: true },
      { id: 'carton', name: 'Carton', conversionFactor: 48, convertsTo: 'pcs', toBaseUnit: 48, isBase: false },
    ],
    variantAttributes: [],
    variants: [
      {
        id: 'nb-a5-default',
        sku: 'NB-A5',
        barcode: '8901234002001',
        attributes: {},
        pricingTiers: [
          { minQty: 1, value: 25 },
        ],
        images: [],
        rackIds: [1],
        currentStock: 150,
      },
    ],
  },

  // 4. Cooking Oil — Simple product with Liter base unit
  {
    id: 4,
    name: 'Cooking Oil',
    description: 'Premium vegetable cooking oil',
    categoryId: 3, // Food & Beverages
    images: [],
    priceSetting: 'fixed',
    hasVariants: false,
    status: 'active',
    supplierIds: [],
    units: [
      { id: 'liter', name: 'Liter', conversionFactor: 1, convertsTo: null, toBaseUnit: 1, isBase: true },
    ],
    variantAttributes: [],
    variants: [
      {
        id: 'co-default',
        sku: 'CO-1L',
        barcode: '8901234003001',
        attributes: {},
        pricingTiers: [
          { minQty: 1, value: 18000 },
        ],
        images: [],
        rackIds: [3],
        currentStock: 0,
      },
    ],
  },
];
