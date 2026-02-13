'use client';

import { useParams } from 'next/navigation';
import ProductForm from '@/components/product/ProductForm';
import AdminLayout from '@/components/layout/AdminLayout';
import { useProductStore } from '@/stores/useProductStore';
import Link from 'next/link';

export default function EditProductPage() {
  const params = useParams();
  const id = Number(params.id);
  const product = useProductStore((s) => s.getProduct(id));

  if (!product) {
    return (
      <AdminLayout>
        <div className="flex flex-col items-center justify-center py-20">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Product not found</h1>
          <p className="text-gray-500 mb-6">
            The product you are looking for does not exist or has been deleted.
          </p>
          <Link
            href="/master/product"
            className="text-blue-600 hover:text-blue-800 text-sm font-medium"
          >
            Back to Product List
          </Link>
        </div>
      </AdminLayout>
    );
  }

  return <ProductForm mode="edit" initialProduct={product} />;
}
