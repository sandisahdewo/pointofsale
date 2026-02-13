'use client';

import React, { useState, useEffect } from 'react';
import Modal from '@/components/ui/Modal';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Textarea from '@/components/ui/Textarea';
import { useRoleStore, Role } from '@/stores/useRoleStore';
import { useToastStore } from '@/stores/useToastStore';

interface RoleFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  editingRole?: Role | null;
}

export default function RoleFormModal({
  isOpen,
  onClose,
  editingRole,
}: RoleFormModalProps) {
  const { roles, addRole, updateRole } = useRoleStore();
  const { addToast } = useToastStore();

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (isOpen) {
      if (editingRole) {
        setName(editingRole.name);
        setDescription(editingRole.description);
      } else {
        setName('');
        setDescription('');
      }
      setErrors({});
    }
  }, [isOpen, editingRole]);

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};
    if (!name.trim()) {
      newErrors.name = 'Name is required';
    } else {
      const duplicate = roles.find(
        (r) =>
          r.name.toLowerCase() === name.trim().toLowerCase() &&
          r.id !== editingRole?.id
      );
      if (duplicate) {
        newErrors.name = 'Role name already exists.';
      }
    }
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    if (editingRole) {
      updateRole(editingRole.id, {
        name: name.trim(),
        description: description.trim(),
      });
      addToast(`Role ${name.trim()} updated successfully.`, 'success');
    } else {
      addRole({
        name: name.trim(),
        description: description.trim(),
        isSystem: false,
      });
      addToast(`Role ${name.trim()} created successfully.`, 'success');
    }
    onClose();
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={editingRole ? 'Edit Role' : 'Create Role'}
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Name"
          placeholder="Role name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          error={errors.name}
        />
        <Textarea
          label="Description"
          placeholder="Brief description of the role's purpose"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={3}
        />
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit">
            {editingRole ? 'Update' : 'Create'}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
