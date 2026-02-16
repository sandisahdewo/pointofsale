'use client';

import React from 'react';
import { useRouter } from 'next/navigation';
import { useSidebarStore } from '@/stores/useSidebarStore';
import { useAuthStore } from '@/stores/useAuthStore';
import Dropdown from '@/components/ui/Dropdown';
import { useToastStore } from '@/stores/useToastStore';

export default function Header() {
  const { toggle } = useSidebarStore();
  const { addToast } = useToastStore();
  const { user, logout } = useAuthStore();
  const router = useRouter();

  const handleLogout = async () => {
    try {
      await logout();
      addToast('Logged out successfully', 'success');
      router.push('/login');
    } catch (error) {
      addToast('Logout failed', 'error');
    }
  };

  const userMenuItems = [
    {
      label: 'Edit Profile',
      onClick: () => addToast('Edit Profile page coming soon', 'info'),
    },
    {
      label: 'Change Password',
      onClick: () => addToast('Change Password page coming soon', 'info'),
    },
    {
      label: 'Logout',
      onClick: handleLogout,
    },
  ];

  return (
    <header className="h-16 bg-white border-b border-gray-200 flex items-center justify-between px-4 fixed top-0 left-0 right-0 z-40">
      <div className="flex items-center gap-3">
        <button
          onClick={toggle}
          className="p-2 text-gray-600 hover:bg-gray-100 rounded-md cursor-pointer"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
            <span className="text-white font-bold text-sm">P</span>
          </div>
          <span className="text-lg font-semibold text-gray-900">Point of Sale</span>
        </div>
      </div>

      <Dropdown
        trigger={
          <div className="flex items-center gap-2 px-3 py-2 rounded-md hover:bg-gray-100">
            <div className="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center">
              <span className="text-gray-600 text-sm font-medium">
                {user?.name?.charAt(0).toUpperCase() || 'U'}
              </span>
            </div>
            <span className="text-sm text-gray-700">{user?.name || 'User'}</span>
            <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </div>
        }
        items={userMenuItems}
      />
    </header>
  );
}
