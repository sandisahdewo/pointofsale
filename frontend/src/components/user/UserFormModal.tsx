'use client';

import React, { useState, useEffect, useMemo } from 'react';
import Modal from '@/components/ui/Modal';
import Input from '@/components/ui/Input';
import Textarea from '@/components/ui/Textarea';
import Select from '@/components/ui/Select';
import ImageUpload from '@/components/ui/ImageUpload';
import MultiSelect from '@/components/ui/MultiSelect';
import Button from '@/components/ui/Button';
import { User, CreateUserInput, UpdateUserInput } from '@/stores/useUserStore';
import { useRoleStore } from '@/stores/useRoleStore';

interface UserFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  user: User | null;
  onSave: (data: CreateUserInput | UpdateUserInput, isEdit: boolean) => Promise<void>;
}

export default function UserFormModal({
  isOpen,
  onClose,
  user,
  onSave,
}: UserFormModalProps) {
  const { roles } = useRoleStore();

  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [address, setAddress] = useState('');
  const [profilePicture, setProfilePicture] = useState<string[]>([]);
  const [selectedRoles, setSelectedRoles] = useState<string[]>([]);
  const [status, setStatus] = useState<'active' | 'inactive'>('active');
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  const isEdit = !!user;

  useEffect(() => {
    if (isOpen) {
      if (user) {
        setName(user.name);
        setEmail(user.email);
        setPhone(user.phone);
        setAddress(user.address);
        setProfilePicture(user.profilePicture ? [user.profilePicture] : []);
        setSelectedRoles((user.roles || []).map((r) => String(r.id)));
        setStatus(user.status === 'pending' ? 'active' : (user.status as 'active' | 'inactive'));
      } else {
        setName('');
        setEmail('');
        setPhone('');
        setAddress('');
        setProfilePicture([]);
        setSelectedRoles([]);
        setStatus('active');
      }
      setErrors({});
      setSubmitting(false);
    }
  }, [isOpen, user]);

  const roleOptions = useMemo(
    () =>
      roles.map((r) => ({
        value: String(r.id),
        label: r.name,
      })),
    [roles]
  );

  const statusOptions = [
    { value: 'active', label: 'Active' },
    { value: 'inactive', label: 'Inactive' },
  ];

  const validate = (): boolean => {
    const errs: Record<string, string> = {};

    if (!name.trim()) {
      errs.name = 'Name is required';
    }

    if (!email.trim()) {
      errs.email = 'Email is required';
    } else {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(email.trim())) {
        errs.email = 'Please enter a valid email address';
      }
    }

    setErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    setSubmitting(true);
    try {
      await onSave({
        name: name.trim(),
        email: email.trim(),
        phone: phone.trim() || undefined,
        address: address.trim() || undefined,
        roleIds: selectedRoles.map(Number),
        ...(isEdit ? { status } : {}),
      }, isEdit);
    } catch (error) {
      // Error handling is done in the parent
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={isEdit ? 'Edit User' : 'Create User'}
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Name"
          placeholder="Full name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          error={errors.name}
          required
          disabled={submitting}
        />
        <Input
          label="Email"
          type="email"
          placeholder="Email address"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          error={errors.email}
          required
          disabled={submitting}
        />
        <Input
          label="Phone"
          placeholder="Phone number"
          value={phone}
          onChange={(e) => setPhone(e.target.value)}
          disabled={submitting}
        />
        <Textarea
          label="Address"
          placeholder="Address"
          value={address}
          onChange={(e) => setAddress(e.target.value)}
          rows={3}
          disabled={submitting}
        />
        <ImageUpload
          label="Profile Picture"
          images={profilePicture}
          onChange={(imgs) => setProfilePicture(imgs.slice(0, 1))}
        />
        <MultiSelect
          label="Roles"
          options={roleOptions}
          value={selectedRoles}
          onChange={setSelectedRoles}
          placeholder="Select roles..."
        />
        {isEdit && (
          <Select
            label="Status"
            options={statusOptions}
            value={status}
            onChange={(e) =>
              setStatus(e.target.value as 'active' | 'inactive')
            }
            disabled={user?.isSuperAdmin || submitting}
          />
        )}
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="outline" onClick={onClose} disabled={submitting}>
            Cancel
          </Button>
          <Button type="submit" disabled={submitting}>
            {submitting ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
