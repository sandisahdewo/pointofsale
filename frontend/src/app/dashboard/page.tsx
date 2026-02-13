'use client';

import React from 'react';
import AdminLayout from '@/components/layout/AdminLayout';

export default function DashboardPage() {
  return (
    <AdminLayout>
      <div className="flex items-center justify-center h-[calc(100vh-12rem)]">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Welcome to Dashboard</h1>
          <p className="text-gray-500">Your admin panel overview will appear here.</p>
        </div>
      </div>
    </AdminLayout>
  );
}
