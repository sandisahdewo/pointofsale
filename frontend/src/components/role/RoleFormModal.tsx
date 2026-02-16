'use client';

import React, { useState, useEffect } from 'react';
import Modal from '@/components/ui/Modal';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Textarea from '@/components/ui/Textarea';
import { Role } from '@/stores/useRoleStore';

interface RoleFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  editingRole?: Role | null;
  onSave: (input: { name: string; description: string }) => Promise<void>;
}

export default function RoleFormModal({
  isOpen,
  onClose,
  editingRole,
  onSave,
}: RoleFormModalProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setErrors({ name: 'Name is required' });
      return;
    }
    setSubmitting(true);
    try {
      await onSave({
        name: name.trim(),
        description: description.trim(),
      });
    } catch {
      // Parent handles error display
    } finally {
      setSubmitting(false);
    }
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
          disabled={submitting}
        />
        <Textarea
          label="Description"
          placeholder="Brief description of the role's purpose"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={3}
          disabled={submitting}
        />
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="outline" onClick={onClose} disabled={submitting}>
            Cancel
          </Button>
          <Button type="submit" disabled={submitting}>
            {submitting ? 'Saving...' : editingRole ? 'Update' : 'Create'}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
