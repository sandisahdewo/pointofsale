'use client';

import { create } from 'zustand';
import { initialRoles } from '@/data/roles';
import { initialPermissions } from '@/data/permissions';
import { initialRolePermissions } from '@/data/rolePermissions';

export interface Role {
  id: number;
  name: string;
  description: string;
  isSystem: boolean;
  createdAt: string;
}

export interface Permission {
  id: number;
  module: string;
  feature: string;
  actions: string[];
}

export interface RolePermission {
  roleId: number;
  permissionId: number;
  actions: string[];
}

interface RoleState {
  roles: Role[];
  permissions: Permission[];
  rolePermissions: RolePermission[];
  addRole: (role: Omit<Role, 'id' | 'createdAt'>) => void;
  updateRole: (id: number, data: Partial<Omit<Role, 'id' | 'createdAt' | 'isSystem'>>) => void;
  deleteRole: (id: number) => void;
  setRolePermissions: (roleId: number, permissionId: number, actions: string[]) => void;
  getRolePermissions: (roleId: number) => RolePermission[];
}

export const useRoleStore = create<RoleState>((set, get) => ({
  roles: initialRoles,
  permissions: initialPermissions,
  rolePermissions: initialRolePermissions,
  addRole: (role) =>
    set((state) => {
      const maxId = state.roles.reduce((max, r) => Math.max(max, r.id), 0);
      return {
        roles: [
          ...state.roles,
          { ...role, id: maxId + 1, createdAt: new Date().toISOString() },
        ],
      };
    }),
  updateRole: (id, data) =>
    set((state) => ({
      roles: state.roles.map((r) =>
        r.id === id && !r.isSystem ? { ...r, ...data } : r
      ),
    })),
  deleteRole: (id) =>
    set((state) => {
      const role = state.roles.find((r) => r.id === id);
      if (!role || role.isSystem) return state;
      return {
        roles: state.roles.filter((r) => r.id !== id),
        rolePermissions: state.rolePermissions.filter((rp) => rp.roleId !== id),
      };
    }),
  setRolePermissions: (roleId, permissionId, actions) =>
    set((state) => {
      const existing = state.rolePermissions.find(
        (rp) => rp.roleId === roleId && rp.permissionId === permissionId
      );
      if (actions.length === 0) {
        return {
          rolePermissions: state.rolePermissions.filter(
            (rp) => !(rp.roleId === roleId && rp.permissionId === permissionId)
          ),
        };
      }
      if (existing) {
        return {
          rolePermissions: state.rolePermissions.map((rp) =>
            rp.roleId === roleId && rp.permissionId === permissionId
              ? { ...rp, actions }
              : rp
          ),
        };
      }
      return {
        rolePermissions: [...state.rolePermissions, { roleId, permissionId, actions }],
      };
    }),
  getRolePermissions: (roleId) =>
    get().rolePermissions.filter((rp) => rp.roleId === roleId),
}));
