'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import ProductForm from '@/components/product/ProductForm';
import AdminLayout from '@/components/layout/AdminLayout';
import { Product, useProductStore } from '@/stores/useProductStore';
import { useToastStore } from '@/stores/useToastStore';
import { ApiError } from '@/lib/api';
import Link from 'next/link';

export default function EditProductPage() {
  const params = useParams();
  const rawId = Array.isArray(params.id) ? params.id[0] : params.id;
  const id = Number(rawId);
  const { fetchProductById } = useProductStore();
  const { addToast } = useToastStore();
  const [product, setProduct] = useState<Product | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    if (!Number.isFinite(id) || id <= 0) {
      setNotFound(true);
      setIsLoading(false);
      return;
    }

    const loadProduct = async () => {
      try {
        setIsLoading(true);
        const loaded = await fetchProductById(id);
        setProduct(loaded);
      } catch (error) {
        if (error instanceof ApiError && error.status === 404) {
          setNotFound(true);
          return;
        }
        if (error instanceof ApiError) {
          addToast(error.message, 'error');
        } else {
          addToast('Failed to load product', 'error');
        }
      } finally {
        setIsLoading(false);
      }
    };

    void loadProduct();
  }, [id, fetchProductById, addToast]);

  if (isLoading) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center py-20 text-sm text-gray-500">
          Loading product...
        </div>
      </AdminLayout>
    );
  }

  if (notFound || !product) {
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
