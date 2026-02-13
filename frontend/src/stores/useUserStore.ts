'use client';

import { create } from 'zustand';
import { initialUsers } from '@/data/users';
import { useRoleStore } from '@/stores/useRoleStore';

export interface User {
  id: number;
  name: string;
  email: string;
  phone: string;
  address: string;
  password: string;
  profilePicture: string;
  roles: number[];
  status: 'active' | 'pending' | 'inactive';
  isSuperAdmin: boolean;
  createdAt: string;
}

interface UserState {
  users: User[];
  addUser: (user: Omit<User, 'id' | 'createdAt'>) => void;
  updateUser: (id: number, data: Partial<Omit<User, 'id' | 'createdAt'>>) => void;
  deleteUser: (id: number) => void;
  approveUser: (id: number) => void;
  removeRoleFromUsers: (roleId: number) => void;
  getUserRoleNames: (id: number) => string[];
}

export const useUserStore = create<UserState>((set, get) => ({
  users: initialUsers,
  addUser: (user) =>
    set((state) => {
      const maxId = state.users.reduce((max, u) => Math.max(max, u.id), 0);
      return {
        users: [
          ...state.users,
          { ...user, id: maxId + 1, createdAt: new Date().toISOString() },
        ],
      };
    }),
  updateUser: (id, data) =>
    set((state) => ({
      users: state.users.map((u) => {
        if (u.id !== id) return u;
        if (u.isSuperAdmin) {
          const { isSuperAdmin: _is, status: _st, ...safeData } = data;
          return { ...u, ...safeData };
        }
        return { ...u, ...data };
      }),
    })),
  deleteUser: (id) =>
    set((state) => {
      const user = state.users.find((u) => u.id === id);
      if (!user || user.isSuperAdmin) return state;
      return { users: state.users.filter((u) => u.id !== id) };
    }),
  approveUser: (id) =>
    set((state) => ({
      users: state.users.map((u) =>
        u.id === id ? { ...u, status: 'active' as const } : u
      ),
    })),
  removeRoleFromUsers: (roleId) =>
    set((state) => ({
      users: state.users.map((u) => ({
        ...u,
        roles: u.roles.filter((r) => r !== roleId),
      })),
    })),
  getUserRoleNames: (id) => {
    const user = get().users.find((u) => u.id === id);
    if (!user) return [];
    const roles = useRoleStore.getState().roles;
    return user.roles
      .map((roleId) => roles.find((r) => r.id === roleId)?.name)
      .filter((name): name is string => !!name);
  },
}));
