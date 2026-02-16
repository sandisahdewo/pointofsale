'use client';

import React, { useState, useMemo, useEffect, useCallback, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import AdminLayout from '@/components/layout/AdminLayout';
import Button from '@/components/ui/Button';
import Checkbox from '@/components/ui/Checkbox';
import ConfirmModal from '@/components/ui/ConfirmModal';
import Link from 'next/link';
import { useRoleStore, Permission, Role } from '@/stores/useRoleStore';
import { useToastStore } from '@/stores/useToastStore';
import { ApiError } from '@/lib/api';

interface ModuleGroup {
  module: string;
  permissions: Permission[];
}

function buildModuleGroups(permissions: Permission[]): ModuleGroup[] {
  const map = new Map<string, Permission[]>();
  for (const perm of permissions) {
    const list = map.get(perm.module) || [];
    list.push(perm);
    map.set(perm.module, list);
  }
  return Array.from(map.entries()).map(([module, perms]) => ({
    module,
    permissions: perms,
  }));
}

type PermissionMap = Map<number, Set<string>>;

function clonePermissionMap(map: PermissionMap): PermissionMap {
  const clone: PermissionMap = new Map();
  for (const [key, value] of map) {
    clone.set(key, new Set(value));
  }
  return clone;
}

function permissionMapEquals(a: PermissionMap, b: PermissionMap): boolean {
  if (a.size !== b.size) return false;
  for (const [key, aSet] of a) {
    const bSet = b.get(key);
    if (!bSet) return false;
    if (aSet.size !== bSet.size) return false;
    for (const v of aSet) {
      if (!bSet.has(v)) return false;
    }
  }
  return true;
}

type CheckState = 'checked' | 'unchecked' | 'indeterminate';

function getFeatureCheckState(
  perm: Permission,
  checkedActions: Set<string> | undefined
): CheckState {
  if (!checkedActions || checkedActions.size === 0) return 'unchecked';
  const available = perm.actions;
  const checkedCount = available.filter((a) => checkedActions.has(a)).length;
  if (checkedCount === 0) return 'unchecked';
  if (checkedCount === available.length) return 'checked';
  return 'indeterminate';
}

function getModuleCheckState(
  perms: Permission[],
  permMap: PermissionMap
): CheckState {
  let totalAvailable = 0;
  let totalChecked = 0;
  for (const perm of perms) {
    const checked = permMap.get(perm.id);
    for (const action of perm.actions) {
      totalAvailable++;
      if (checked?.has(action)) totalChecked++;
    }
  }
  if (totalChecked === 0) return 'unchecked';
  if (totalChecked === totalAvailable) return 'checked';
  return 'indeterminate';
}

export default function PermissionsPage() {
  const params = useParams();
  const router = useRouter();
  const roleId = Number(params.id);
  const { getRole, fetchPermissions, fetchRolePermissions, updateRolePermissions } = useRoleStore();
  const { addToast } = useToastStore();

  const [role, setRole] = useState<Role | null>(null);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [permMap, setPermMap] = useState<PermissionMap>(new Map());
  const savedSnapshot = useRef<PermissionMap>(new Map());

  const isSuperAdmin = role?.isSystem && role?.name === 'Super Admin';

  const moduleGroups = useMemo(() => buildModuleGroups(permissions), [permissions]);

  const allActions = useMemo(() => {
    const actionSet = new Set<string>();
    for (const perm of permissions) {
      for (const action of perm.actions) {
        actionSet.add(action);
      }
    }
    return Array.from(actionSet);
  }, [permissions]);

  useEffect(() => {
    const loadData = async () => {
      setLoading(true);
      try {
        const [roleData, permsData, rolePermsData] = await Promise.all([
          getRole(roleId),
          fetchPermissions(),
          fetchRolePermissions(roleId),
        ]);
        setRole(roleData);
        setPermissions(permsData);

        const map: PermissionMap = new Map();
        for (const rp of rolePermsData.permissions) {
          if (rp.grantedActions.length > 0) {
            map.set(rp.permissionId, new Set(rp.grantedActions));
          }
        }
        setPermMap(map);
        savedSnapshot.current = clonePermissionMap(map);
      } catch (error) {
        if (error instanceof ApiError && error.status === 404) {
          setRole(null);
        } else {
          addToast('Failed to load permissions', 'error');
        }
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, [roleId, getRole, fetchPermissions, fetchRolePermissions, addToast]);

  const [collapsedModules, setCollapsedModules] = useState<Set<string>>(
    new Set()
  );

  const [confirmModal, setConfirmModal] = useState<{
    isOpen: boolean;
    title: string;
    message: string;
    cancelLabel: string;
    confirmLabel: string;
    variant: 'primary' | 'danger';
    onConfirm: () => void;
  }>({
    isOpen: false,
    title: '',
    message: '',
    cancelLabel: 'Stay',
    confirmLabel: 'Leave',
    variant: 'danger',
    onConfirm: () => {},
  });

  const isDirty = useMemo(
    () => !permissionMapEquals(permMap, savedSnapshot.current),
    [permMap]
  );

  // Browser beforeunload warning
  useEffect(() => {
    const handler = (e: BeforeUnloadEvent) => {
      if (isDirty) {
        e.preventDefault();
      }
    };
    window.addEventListener('beforeunload', handler);
    return () => window.removeEventListener('beforeunload', handler);
  }, [isDirty]);

  const toggleModule = (module: string) => {
    setCollapsedModules((prev) => {
      const next = new Set(prev);
      if (next.has(module)) {
        next.delete(module);
      } else {
        next.add(module);
      }
      return next;
    });
  };

  const toggleAction = useCallback(
    (permId: number, action: string, checked: boolean) => {
      setPermMap((prev) => {
        const next = clonePermissionMap(prev);
        const current = next.get(permId) || new Set<string>();
        if (checked) {
          current.add(action);
        } else {
          current.delete(action);
        }
        if (current.size === 0) {
          next.delete(permId);
        } else {
          next.set(permId, current);
        }
        return next;
      });
    },
    []
  );

  const toggleFeature = useCallback(
    (perm: Permission) => {
      setPermMap((prev) => {
        const next = clonePermissionMap(prev);
        const current = next.get(perm.id);
        const state = getFeatureCheckState(perm, current);
        if (state === 'checked') {
          // Uncheck all
          next.delete(perm.id);
        } else {
          // Check all (from unchecked or indeterminate)
          next.set(perm.id, new Set(perm.actions));
        }
        return next;
      });
    },
    []
  );

  const toggleModuleCheckbox = useCallback(
    (group: ModuleGroup) => {
      setPermMap((prev) => {
        const next = clonePermissionMap(prev);
        const state = getModuleCheckState(group.permissions, prev);
        if (state === 'checked') {
          // Uncheck all features in this module
          for (const perm of group.permissions) {
            next.delete(perm.id);
          }
        } else {
          // Check all (from unchecked or indeterminate)
          for (const perm of group.permissions) {
            next.set(perm.id, new Set(perm.actions));
          }
        }
        return next;
      });
    },
    []
  );

  const handleSave = async () => {
    setSaving(true);
    try {
      const permissionsInput = [];
      for (const perm of permissions) {
        const actions = permMap.get(perm.id);
        if (actions && actions.size > 0) {
          permissionsInput.push({
            permissionId: perm.id,
            actions: Array.from(actions),
          });
        }
      }
      await updateRolePermissions(roleId, permissionsInput);
      savedSnapshot.current = clonePermissionMap(permMap);
      addToast(`Permissions updated for ${role?.name}.`, 'success');
      router.push('/settings/roles');
    } catch (error) {
      if (error instanceof ApiError) {
        addToast(error.message, 'error');
      } else {
        addToast('An error occurred', 'error');
      }
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    if (isDirty) {
      setConfirmModal({
        isOpen: true,
        title: 'Unsaved Changes',
        message:
          'You have unsaved changes. Are you sure you want to leave?',
        cancelLabel: 'Stay',
        confirmLabel: 'Leave',
        variant: 'danger',
        onConfirm: () => {
          setConfirmModal((prev) => ({ ...prev, isOpen: false }));
          router.push('/settings/roles');
        },
      });
      return;
    }
    router.push('/settings/roles');
  };

  if (loading) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center py-20">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      </AdminLayout>
    );
  }

  if (!role) {
    return (
      <AdminLayout>
        <div className="flex flex-col items-center justify-center py-20">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">
            Role not found
          </h1>
          <p className="text-gray-500 mb-6">
            The role you are looking for does not exist or has been deleted.
          </p>
          <Link
            href="/settings/roles"
            className="text-blue-600 hover:text-blue-800 text-sm font-medium"
          >
            Back to Roles
          </Link>
        </div>
      </AdminLayout>
    );
  }

  return (
    <AdminLayout>
      <div className="space-y-6">
        {/* Sticky header */}
        <div className="sticky top-16 z-10 bg-gray-50 -mx-6 -mt-6 px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <Link
              href="/settings/roles"
              className="inline-flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
              onClick={(e) => {
                if (isDirty) {
                  e.preventDefault();
                  handleCancel();
                }
              }}
            >
              <svg
                className="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M15 19l-7-7 7-7"
                />
              </svg>
              Back to Roles
            </Link>
            {!isSuperAdmin && (
              <div className="flex gap-2">
                <Button variant="outline" onClick={handleCancel} disabled={saving}>
                  Cancel
                </Button>
                <Button onClick={handleSave} disabled={saving}>
                  {saving ? 'Saving...' : 'Save'}
                </Button>
              </div>
            )}
          </div>
        </div>

        {/* Page title */}
        <h1 className="text-2xl font-bold text-gray-900">
          Permissions — {role.name}
        </h1>

        {/* Super Admin notice */}
        {isSuperAdmin && (
          <div className="rounded-md bg-blue-50 border border-blue-200 p-4">
            <p className="text-sm text-blue-800">
              Super Admin has full access to all features. Permissions cannot be
              modified.
            </p>
          </div>
        )}

        {/* Permissions tree */}
        <div className="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-gray-50 text-gray-600 uppercase text-xs">
                <tr>
                  <th className="px-4 py-3 font-medium">Module / Feature</th>
                  {allActions.map((action) => (
                    <th
                      key={action}
                      className="px-4 py-3 font-medium text-center w-24"
                    >
                      {action.charAt(0).toUpperCase() + action.slice(1)}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {moduleGroups.map((group) => {
                  const moduleState = isSuperAdmin
                    ? 'checked'
                    : getModuleCheckState(group.permissions, permMap);
                  const isCollapsed = collapsedModules.has(group.module);

                  return (
                    <React.Fragment key={group.module}>
                      {/* Module row */}
                      <tr className="bg-gray-50">
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-2">
                            <Checkbox
                              checked={moduleState === 'checked'}
                              indeterminate={moduleState === 'indeterminate'}
                              onChange={() => toggleModuleCheckbox(group)}
                              disabled={isSuperAdmin}
                            />
                            <button
                              type="button"
                              onClick={() => toggleModule(group.module)}
                              className="flex items-center gap-1 font-medium text-gray-900 cursor-pointer"
                            >
                              <svg
                                className={`w-4 h-4 transition-transform ${
                                  isCollapsed ? '-rotate-90' : ''
                                }`}
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                              >
                                <path
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                  strokeWidth={2}
                                  d="M19 9l-7 7-7-7"
                                />
                              </svg>
                              {group.module}
                            </button>
                          </div>
                        </td>
                        {allActions.map((action) => (
                          <td key={action} className="px-4 py-3" />
                        ))}
                      </tr>

                      {/* Feature rows */}
                      {!isCollapsed &&
                        group.permissions.map((perm) => {
                          const checkedActions = permMap.get(perm.id);
                          const featureState = isSuperAdmin
                            ? 'checked'
                            : getFeatureCheckState(perm, checkedActions);

                          return (
                            <tr key={perm.id} className="hover:bg-gray-50">
                              <td className="px-4 py-3 pl-12">
                                <Checkbox
                                  label={perm.feature}
                                  checked={featureState === 'checked'}
                                  indeterminate={
                                    featureState === 'indeterminate'
                                  }
                                  onChange={() => toggleFeature(perm)}
                                  disabled={isSuperAdmin}
                                />
                              </td>
                              {allActions.map((action) => {
                                const isAvailable =
                                  perm.actions.includes(action);
                                if (!isAvailable) {
                                  return (
                                    <td
                                      key={action}
                                      className="px-4 py-3 text-center text-gray-400"
                                    >
                                      —
                                    </td>
                                  );
                                }
                                const isChecked = isSuperAdmin
                                  ? true
                                  : checkedActions?.has(action) ?? false;
                                return (
                                  <td
                                    key={action}
                                    className="px-4 py-3 text-center"
                                  >
                                    <Checkbox
                                      checked={isChecked}
                                      onChange={(checked) =>
                                        toggleAction(perm.id, action, checked)
                                      }
                                      disabled={isSuperAdmin}
                                    />
                                  </td>
                                );
                              })}
                            </tr>
                          );
                        })}
                    </React.Fragment>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <ConfirmModal
        isOpen={confirmModal.isOpen}
        onClose={() =>
          setConfirmModal((prev) => ({ ...prev, isOpen: false }))
        }
        onConfirm={confirmModal.onConfirm}
        title={confirmModal.title}
        message={confirmModal.message}
        cancelLabel={confirmModal.cancelLabel}
        confirmLabel={confirmModal.confirmLabel}
        variant={confirmModal.variant}
      />
    </AdminLayout>
  );
}
