'use client';

import { create } from 'zustand';
import { apiClient, ApiError } from '@/lib/api';

export interface Role {
  id: number;
  name: string;
  description?: string;
  isSystem?: boolean;
}

export interface Permission {
  module: string;
  feature: string;
  actions: string[];
}

export interface AuthUser {
  id: number;
  name: string;
  email: string;
  phone?: string;
  address?: string;
  profilePicture?: string | null;
  status: string;
  isSuperAdmin: boolean;
  createdAt: string;
  updatedAt: string;
  roles: Role[];
}

interface LoginResponseData {
  user: AuthUser;
  accessToken: string;
  refreshToken: string;
  expiresAt: string;
}

interface CurrentUserData extends AuthUser {
  permissions: Permission[];
}

interface AuthState {
  user: AuthUser | null;
  permissions: Permission[];
  isAuthenticated: boolean;
  isLoading: boolean;
  isInitialized: boolean;

  initialize: () => void;
  login: (email: string, password: string) => Promise<void>;
  register: (name: string, email: string, password: string, confirmPassword: string) => Promise<string>;
  logout: () => Promise<void>;
  forgotPassword: (email: string) => Promise<string>;
  resetPassword: (token: string, password: string, confirmPassword: string) => Promise<string>;
  fetchCurrentUser: () => Promise<void>;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  permissions: [],
  isAuthenticated: false,
  isLoading: false,
  isInitialized: false,

  initialize: () => {
    if (get().isInitialized) return;

    const accessToken = localStorage.getItem('accessToken');
    if (accessToken) {
      set({ isAuthenticated: true, isInitialized: true });
      // Fetch user data in background
      get().fetchCurrentUser().catch(() => {
        get().clearAuth();
      });
    } else {
      set({ isInitialized: true });
    }
  },

  login: async (email: string, password: string) => {
    set({ isLoading: true });
    try {
      const response = await apiClient<LoginResponseData>('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      });

      const { user, accessToken, refreshToken, expiresAt } = response.data;
      localStorage.setItem('accessToken', accessToken);
      localStorage.setItem('refreshToken', refreshToken);
      localStorage.setItem('expiresAt', expiresAt);

      set({
        user,
        isAuthenticated: true,
        isLoading: false,
        isInitialized: true,
      });

      // Fetch permissions in background
      get().fetchCurrentUser().catch(() => {});
    } catch (error) {
      set({ isLoading: false });
      throw error;
    }
  },

  register: async (name: string, email: string, password: string, confirmPassword: string) => {
    set({ isLoading: true });
    try {
      const response = await apiClient<AuthUser>('/api/v1/auth/register', {
        method: 'POST',
        body: JSON.stringify({ name, email, password, confirmPassword }),
      });
      set({ isLoading: false });
      return response.message || 'Registration successful. Please wait for admin approval.';
    } catch (error) {
      set({ isLoading: false });
      throw error;
    }
  },

  logout: async () => {
    const refreshToken = localStorage.getItem('refreshToken');
    try {
      await apiClient('/api/v1/auth/logout', {
        method: 'POST',
        body: JSON.stringify({ refreshToken }),
      });
    } catch {
      // Always clear local state regardless of API response
    }
    get().clearAuth();
  },

  forgotPassword: async (email: string) => {
    set({ isLoading: true });
    try {
      const response = await apiClient<null>('/api/v1/auth/forgot-password', {
        method: 'POST',
        body: JSON.stringify({ email }),
      });
      set({ isLoading: false });
      return response.message || 'If the email exists, a reset link has been sent.';
    } catch (error) {
      set({ isLoading: false });
      throw error;
    }
  },

  resetPassword: async (token: string, password: string, confirmPassword: string) => {
    set({ isLoading: true });
    try {
      const response = await apiClient<null>('/api/v1/auth/reset-password', {
        method: 'POST',
        body: JSON.stringify({ token, password, confirmPassword }),
      });
      set({ isLoading: false });
      return response.message || 'Password reset successfully. Please login with your new password.';
    } catch (error) {
      set({ isLoading: false });
      throw error;
    }
  },

  fetchCurrentUser: async () => {
    try {
      const response = await apiClient<CurrentUserData>('/api/v1/auth/me');
      const { permissions, ...userData } = response.data;
      set({ user: userData, permissions });
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        get().clearAuth();
      }
      throw error;
    }
  },

  clearAuth: () => {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('expiresAt');
    set({
      user: null,
      permissions: [],
      isAuthenticated: false,
    });
  },
}));
