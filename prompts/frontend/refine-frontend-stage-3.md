# Frontend App - Admin Panel (Phase 3)

## Overview

Phase 3 adds a **Settings** section to the sidebar with two modules: **User Management** and **Roles & Permissions**. Users can be created by a super admin or self-registered (pending approval). Roles group permissions that control access to application features. Permission enforcement is **not** implemented in this phase — all menus and features remain accessible to all users. Enforcement will be added in a future phase with the backend API.

> **Backend note**: All data is client-side (Zustand) for now. Design stores and data models so they can be replaced with backend API calls in the future without changing the UI.

---

## 1. Sidebar Update

Add a new top-level **Settings** section to the existing sidebar menu, placed **after** Report:

```
Master Data
  Product
  Category
  Supplier
Transaction
  Sales
  Purchase
Report
  Sales Report
  Purchase Report
Settings              ← NEW
  Users               ← NEW
  Roles & Permissions ← NEW
```

---

## 2. User Management (`/settings/users`)

Full CRUD page for managing application users.

### 2.1 User Data Model

```typescript
interface User {
  id: number;
  name: string;
  email: string;
  phone: string;
  address: string;
  password: string;        // stored in state but NEVER displayed in the UI
  profilePicture: string;  // base64 or URL string, optional
  roles: number[];         // array of role IDs, can be empty
  status: 'active' | 'pending' | 'inactive';
  isSuperAdmin: boolean;   // true only for the built-in super admin user
  createdAt: string;       // ISO date string
}
```

**Status meanings:**
- **Active** — User can log in and use the system.
- **Pending** — User self-registered and is awaiting super admin approval.
- **Inactive** — User has been deactivated by super admin.

### 2.2 User List View

Table-based list following the existing table design, layout, and style (same as Master Category / Product):

**Table columns:**

| Column | Sortable | Notes |
|--------|----------|-------|
| ID | Yes | Auto-increment |
| Profile Picture | No | Avatar thumbnail (show initials placeholder if no image) |
| Name | Yes | Full name |
| Email | Yes | Email address |
| Phone | No | Phone number |
| Roles | No | Comma-separated role names, or "—" if no roles assigned |
| Status | Yes | Badge: green for Active, yellow for Pending, gray for Inactive |
| Actions | No | Action buttons (see below) |

**Features:**
- **Search**: text search across name and email.
- **Sorting**: sortable on ID, Name, Email, Status columns.
- **Pagination**: with items-per-page selector (reuse existing Table component).
- **Create User** button at the top → opens modal form (section 2.3).

**Actions per row:**
- **Edit** (pencil icon) → opens modal form pre-filled with user data.
- **Delete** (trash icon) → opens confirmation modal.
- **Approve** (check icon) → only shown when status is `pending`. Sets status to `active` and shows success toast: "User {name} has been approved."
- **Reject** (x icon) → only shown when status is `pending`. Opens confirmation modal: title "Reject User", message "Are you sure you want to reject **{name}**? This will remove their registration.", Cancel / Reject (danger). On confirm, delete the user from state and show toast: "User {name} has been rejected."

**Super admin row restrictions:**
- The super admin user row **cannot** be deleted — the delete button is hidden or disabled with tooltip: "Super admin cannot be deleted."
- The super admin user **can** be edited (name, email, phone, address, profile picture) but the `isSuperAdmin` flag and `status` cannot be changed.

### 2.3 Create / Edit User Modal

Open a modal (not a full page — the form is simple enough for a modal) for creating and editing users.

**Form fields:**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Name | Text input | Yes | Full name |
| Email | Text input | Yes | Must be valid email format. Must be unique across all users. |
| Phone | Text input | No | Phone number |
| Address | Textarea | No | Address |
| Profile Picture | Image upload | No | Single image. Reuse existing `ImageUpload` component in single-image mode. |
| Roles | Multi-select | No | Select from existing roles. User can have zero or more roles. Show as checkboxes or multi-select dropdown. |
| Status | Select/Toggle | Yes | Active / Inactive. Default: Active. Only shown in edit mode. Not shown when creating (new users default to Active). |

**Create mode specifics:**
- Title: "Create User"
- No password field is shown. The system generates a default password internally (mock: store `"password123"` or similar in state).
- On successful save, show a toast: "User created successfully. Credentials have been sent to {email}." (simulate the email — no actual email sent).
- Validate email uniqueness against existing users. Show inline error: "Email is already registered." if duplicate.

**Edit mode specifics:**
- Title: "Edit User"
- Pre-fill all fields with existing user data.
- Password is **never** shown or editable in the edit form. (Password reset is a separate future feature.)
- On successful save, show toast: "User updated successfully."

**Buttons:**
- **Cancel** (secondary) — closes modal without saving.
- **Save** (primary) — validates and saves.

### 2.4 Delete User Confirmation

Use the existing `ConfirmModal` component:

| Property | Value |
|----------|-------|
| Title | Delete User |
| Message | Are you sure you want to delete **{name}**? This action cannot be undone. |
| Cancel button | Cancel |
| Confirm button | Delete (danger variant) |

On confirm, remove user from state and show toast: "User {name} has been deleted."

### 2.5 State Management

Create a new Zustand store `useUserStore` with:

- `users: User[]` — list of all users.
- `addUser(user)` — add a new user (auto-generates ID and createdAt).
- `updateUser(id, data)` — update user fields.
- `deleteUser(id)` — remove a user (block deletion of super admin).
- `approveUser(id)` — set status to `active`.
- `getUserRoleNames(id)` — helper to resolve role IDs to role names (reads from role store).

---

## 3. Roles & Permissions (`/settings/roles`)

Manage roles and assign permissions to each role.

### 3.1 Role Data Model

```typescript
interface Role {
  id: number;
  name: string;
  description: string;
  isSystem: boolean;   // true for built-in roles (Super Admin) that cannot be deleted
  createdAt: string;
}
```

### 3.2 Permission Data Model

Permissions are **seeder data** — predefined by the developer and not manageable via CRUD by the user. When new features are added to the app, developers add corresponding permission entries to the seed data.

```typescript
interface Permission {
  id: number;
  module: string;       // top-level group, e.g. "Master Data"
  feature: string;      // specific feature, e.g. "Product"
  actions: string[];    // available actions for this feature, e.g. ["read", "create", "update", "delete", "export"]
}

interface RolePermission {
  roleId: number;
  permissionId: number;
  actions: string[];    // subset of the permission's actions that are granted
}
```

### 3.3 Permission Seed Data

Define permissions that mirror the application's current module/feature structure. Not every feature has all five actions — some features may have fewer.

| Module | Feature | Available Actions |
|--------|---------|-------------------|
| Master Data | Product | Read, Create, Update, Delete, Export |
| Master Data | Category | Read, Create, Update, Delete |
| Master Data | Supplier | Read, Create, Update, Delete, Export |
| Transaction | Sales | Read, Create, Update, Delete, Export |
| Transaction | Purchase | Read, Create, Update, Delete, Export |
| Report | Sales Report | Read, Export |
| Report | Purchase Report | Read, Export |
| Settings | Users | Read, Create, Update, Delete |
| Settings | Roles & Permissions | Read, Create, Update, Delete |

> Reports only have Read and Export because they display data — there is nothing to Create, Update, or Delete.

### 3.4 Role List View

Table-based list following the existing table design:

**Table columns:**

| Column | Sortable | Notes |
|--------|----------|-------|
| ID | Yes | Auto-increment |
| Name | Yes | Role name |
| Description | Yes | Short description of the role's purpose |
| Users | No | Count of users assigned to this role (e.g., "3 users") |
| Actions | No | Action buttons (see below) |

**Features:**
- **Search**: text search across name and description.
- **Sorting**: sortable on ID, Name, Description columns.
- **Pagination**: with items-per-page selector.
- **Create Role** button at the top → opens modal form (section 3.5).

**Actions per row:**
- **Permissions** (shield/key icon with label "Permissions") → navigates to permission assignment page `/settings/roles/[id]/permissions` (section 3.6).
- **Edit** (pencil icon) → opens modal form pre-filled with role data.
- **Delete** (trash icon) → opens confirmation modal.

**System role restrictions (Super Admin):**
- The "Super Admin" role **cannot** be deleted — delete button hidden or disabled with tooltip: "System role cannot be deleted."
- The "Super Admin" role **cannot** be edited — edit button hidden or disabled with tooltip: "System role cannot be edited."
- The "Super Admin" role's Permissions button navigates to the permission page with **all actions checked and disabled** (read-only). A notice at the top: "Super Admin has full access to all features. Permissions cannot be modified." The Save button is hidden.

### 3.5 Create / Edit Role Modal

Modal form for creating and editing roles.

**Form fields:**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Name | Text input | Yes | Must be unique across all roles. |
| Description | Textarea | No | Brief description of the role's purpose. |

**Create mode:**
- Title: "Create Role"
- On save, show toast: "Role {name} created successfully."

**Edit mode:**
- Title: "Edit Role"
- Pre-fill fields.
- On save, show toast: "Role {name} updated successfully."

**Validation:**
- Name is required.
- Name must be unique (case-insensitive). Show inline error: "Role name already exists."

### 3.6 Permissions Page (`/settings/roles/[id]/permissions`)

When the user clicks the **Permissions** button on a role row, navigate to a **full page** for permission assignment. This is a dedicated page (not a modal) because the permission tree will grow as new features are added to the application.

**Page layout:**

```
┌─────────────────────────────────────────────────────────────────────────┐
│  ← Back to Roles                                    [Save] [Cancel]    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Permissions — Manager                                                  │
│                                                                         │
│  Module / Feature              | Read | Create | Update | Delete | Export│
│  ────────────────────────────────────────────────────────────────────── │
│  ☑ ▼ Master Data                                                        │
│      ☑ Product                 |  ☑   |   ☑    |   ☑   |   ☐   |  ☑   │
│      ☑ Category                |  ☑   |   ☑    |   ☑   |   ☐   |  —   │
│      ☐ Supplier                |  ☐   |   ☐    |   ☐   |   ☐   |  ☐   │
│                                                                         │
│  ☐ ▼ Transaction                                                        │
│      ☑ Sales                   |  ☑   |   ☑    |   ☐   |   ☐   |  ☑   │
│      ☑ Purchase                |  ☑   |   ☑    |   ☐   |   ☐   |  ☑   │
│                                                                         │
│  ☑ ▼ Report                                                             │
│      ☑ Sales Report            |  ☑   |   —    |   —   |   —   |  ☑   │
│      ☐ Purchase Report         |  ☐   |   —    |   —   |   —   |  ☐   │
│                                                                         │
│  ☐ ▼ Settings                                                           │
│      ☐ Users                   |  ☐   |   ☐    |   ☐   |   ☐   |  —   │
│      ☐ Roles & Permissions     |  ☐   |   ☐    |   ☐   |   ☐   |  —   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

**Behavior:**

- **Module group rows** (Master Data, Transaction, etc.) are collapsible — click the arrow ▼/▶ to expand/collapse. Default: all expanded.
- **Module-level checkbox** (beside the module name, e.g., "☑ Master Data"):
  - **Check it** → enables all available actions for **all features** within that module.
  - **Uncheck it** → disables all actions for all features within that module.
  - **Indeterminate state** — shown when some (but not all) child feature actions are checked.
  - Clicking an indeterminate module checkbox **checks all** (select all behavior).
- **Feature checkbox** (left of feature name, e.g., "☑ Product") is a "select all" toggle for that feature row:
  - Check it → enables all available actions for that feature.
  - Uncheck it → disables all actions for that feature.
  - Indeterminate state when some actions are checked.
  - Clicking an indeterminate feature checkbox **checks all** available actions.
  - Toggling a feature checkbox also recalculates the parent module checkbox state.
- **Action checkboxes** — individual toggles per action per feature. Toggling an action recalculates both the feature checkbox and the module checkbox states.
- **"—" (dash)** is shown instead of a checkbox when that action is not available for the feature (e.g., Report features have no Create/Update/Delete). These cells are non-interactive. Dashes do not count toward the select-all calculation — only available actions are considered.

**Page header:**
- **"← Back to Roles"** link at the top-left → navigates back to `/settings/roles`.
- **Save** and **Cancel** buttons at the top-right (sticky, visible without scrolling — same pattern as the Product form).

**Buttons:**
- **Cancel** — navigates back to role list. If there are unsaved changes, show confirmation modal: title "Unsaved Changes", message "You have unsaved changes. Are you sure you want to leave?", Stay / Leave.
- **Save** — persists the permission assignments to the role, shows toast: "Permissions updated for {role name}.", and navigates back to the role list.

### 3.7 Delete Role Confirmation

| Property | Value |
|----------|-------|
| Title | Delete Role |
| Message | Are you sure you want to delete the role **{name}**? Users assigned to this role will lose these permissions. |
| Cancel button | Cancel |
| Confirm button | Delete (danger variant) |

On confirm:
- Remove the role from state.
- Remove the role ID from all users' `roles[]` arrays.
- Remove all `RolePermission` entries for this role.
- Show toast: "Role {name} has been deleted."

### 3.8 State Management

Create a new Zustand store `useRoleStore` with:

- `roles: Role[]` — list of all roles.
- `permissions: Permission[]` — seeded permission definitions (read-only for users).
- `rolePermissions: RolePermission[]` — which actions each role has for each permission.
- `addRole(role)` — add a new role.
- `updateRole(id, data)` — update role fields.
- `deleteRole(id)` — remove a role (block system roles). Also cleans up `rolePermissions`.
- `setRolePermissions(roleId, permissionId, actions)` — set/update the granted actions for a role on a permission.
- `getRolePermissions(roleId)` — get all permission assignments for a role.

---

## 4. Mock Data

### 4.1 Mock Users

Provide **6 predefined users** to populate the user list:

| # | Name | Email | Phone | Roles | Status | isSuperAdmin | Notes |
|---|------|-------|-------|-------|--------|--------------|-------|
| 1 | Super Admin | admin@pointofsale.com | +62-812-0000-0001 | Super Admin | Active | true | Cannot be deleted |
| 2 | Budi Santoso | budi@pointofsale.com | +62-812-0000-0002 | Manager | Active | false | |
| 3 | Siti Rahayu | siti@pointofsale.com | +62-812-0000-0003 | Cashier | Active | false | |
| 4 | Ahmad Wijaya | ahmad@pointofsale.com | +62-812-0000-0004 | Warehouse, Accountant | Active | false | Multiple roles |
| 5 | Dewi Lestari | dewi@pointofsale.com | +62-812-0000-0005 | Cashier | Inactive | false | Deactivated user |
| 6 | Rizky Pratama | rizky@pointofsale.com | +62-812-0000-0006 | — (none) | Pending | false | Self-registered, awaiting approval |

All mock users have default password `"password123"` stored in state (never displayed in UI).

### 4.2 Mock Roles

| # | Name | Description | isSystem |
|---|------|-------------|----------|
| 1 | Super Admin | Full system access. Cannot be modified or deleted. | true |
| 2 | Manager | Manage products, transactions, and view reports. | false |
| 3 | Cashier | Process sales transactions. | false |
| 4 | Accountant | View transactions and generate reports. | false |
| 5 | Warehouse | Manage product stock and purchase orders. | false |

### 4.3 Mock Role Permissions

Pre-assign sensible default permissions:

**Super Admin** — all actions on all features (enforced by system, not stored as individual entries).

**Manager:**
- Master Data: Product (all), Category (all), Supplier (all)
- Transaction: Sales (Read, Create, Update, Export), Purchase (Read, Create, Update, Export)
- Report: Sales Report (Read, Export), Purchase Report (Read, Export)
- Settings: none

**Cashier:**
- Transaction: Sales (Read, Create)
- Report: Sales Report (Read)

**Accountant:**
- Transaction: Sales (Read, Export), Purchase (Read, Export)
- Report: Sales Report (Read, Export), Purchase Report (Read, Export)

**Warehouse:**
- Master Data: Product (Read, Update), Supplier (Read)
- Transaction: Purchase (Read, Create, Update)

---

## 5. Validation Summary

### User-level
- Name is required.
- Email is required and must be a valid email format.
- Email must be unique across all users (case-insensitive).
- Super admin user cannot be deleted.
- Super admin's `isSuperAdmin` flag and `status` cannot be changed.

### Role-level
- Name is required.
- Name must be unique across all roles (case-insensitive).
- System roles (Super Admin) cannot be deleted or edited.
- Deleting a role cascades: removes role from users' `roles[]` and deletes associated `rolePermissions`.

### Permission-level
- Permissions are read-only seed data — no user-facing CRUD.
- Only the action assignments (via checkboxes) are editable per role.
- Super Admin role's permissions are always full and non-editable.

---

## 6. New Components

Build these new components as needed (or extend existing ones):

- **Checkbox** — standard checkbox input with label, supports checked / unchecked / indeterminate states. Used in the permission tree.
- **MultiSelect** — multi-value select input (dropdown with checkboxes or tag-style). Used for assigning roles to users. Can reuse/extend the existing `Select` component or build as a new component.
- **Badge** — small colored label. Used for user status (Active/Pending/Inactive) and role tags. May already exist in current styling — formalize as a reusable component if not.
- **Avatar** — circular image with initials fallback. Used for user profile picture in the table.

---

## 7. Confirmation Dialogs

All confirmation dialogs use the existing `ConfirmModal` component (not `window.confirm()`).

| Trigger | Title | Message | Cancel | Confirm |
|---------|-------|---------|--------|---------|
| Delete user | Delete User | Are you sure you want to delete **{name}**? This action cannot be undone. | Cancel | Delete (danger) |
| Reject pending user | Reject User | Are you sure you want to reject **{name}**? This will remove their registration. | Cancel | Reject (danger) |
| Delete role | Delete Role | Are you sure you want to delete the role **{name}**? Users assigned to this role will lose these permissions. | Cancel | Delete (danger) |
| Leave permissions page with unsaved changes | Unsaved Changes | You have unsaved changes. Are you sure you want to leave? | Stay | Leave |

---

## 8. General Notes

- **No permission enforcement in this phase.** All menus and routes remain accessible regardless of the logged-in user's role. This phase only builds the UI for managing roles and permissions. Enforcement (route guards, menu filtering, feature toggling) will be implemented in a future phase with the backend API.
- **Toast notifications** on all create/update/delete/approve/reject actions.
- Follow existing table design, layout, and styling patterns established in Master Category and Master Product pages.
- User and role forms use **modals** (not full pages) since the forms are simple.
- The permission tree uses a **full page** (`/settings/roles/[id]/permissions`) — not a modal — because the permission list will grow as features are added. The page follows the same layout pattern as the Product form (sticky header with Save/Cancel, back link, content below).
- Profile picture upload reuses the existing `ImageUpload` component in single-image mode.
- All data persisted in Zustand stores with mock initial data loaded on app start.
