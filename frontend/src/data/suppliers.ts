import { Supplier } from '@/stores/useSupplierStore';

export const initialSuppliers: Supplier[] = [
  {
    id: 1,
    name: 'PT Sumber Makmur',
    address: 'Jl. Industri No. 45, Jakarta',
    phone: '+62-21-5550001',
    email: 'order@sumbermakmur.co.id',
    website: 'sumbermakmur.co.id',
    bankAccounts: [
      { id: 'ba-1-1', accountName: 'BCA - Main Account', accountNumber: '1234567890' },
      { id: 'ba-1-2', accountName: 'Mandiri - Secondary Account', accountNumber: '0987654321' },
    ],
    active: true,
    createdAt: '2026-01-15T10:00:00.000Z',
  },
  {
    id: 2,
    name: 'CV Jaya Abadi',
    address: 'Jl. Perdagangan No. 12, Surabaya',
    phone: '+62-31-5550002',
    email: 'sales@jayaabadi.com',
    website: '',
    bankAccounts: [
      { id: 'ba-2-1', accountName: 'BCA - Business Account', accountNumber: '1122334455' },
    ],
    active: true,
    createdAt: '2026-01-20T09:30:00.000Z',
  },
  {
    id: 3,
    name: 'UD Berkah Sentosa',
    address: 'Jl. Pasar Baru No. 8, Bandung',
    phone: '',
    email: '',
    website: '',
    bankAccounts: [],
    active: true,
    createdAt: '2026-02-01T14:15:00.000Z',
  },
  {
    id: 4,
    name: 'PT Global Supplies',
    address: 'Jl. Raya Serpong No. 100, Tangerang',
    phone: '+62-21-5550004',
    email: 'info@globalsupplies.co.id',
    website: 'globalsupplies.co.id',
    bankAccounts: [
      { id: 'ba-4-1', accountName: 'BNI - Corporate Account', accountNumber: '5566778899' },
      { id: 'ba-4-2', accountName: 'BRI - Operations Account', accountNumber: '9988776655' },
    ],
    active: false,
    createdAt: '2026-01-10T08:00:00.000Z',
  },
];
