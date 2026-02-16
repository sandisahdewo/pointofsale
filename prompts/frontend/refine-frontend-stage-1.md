# Frontend App - Admin Panel (Phase 1)

## Tech Stack

- **Framework**: Next.js (latest version)
- **Rendering**: Pure CSR (SPA-like) - all pages use `'use client'`, no SSR/SSG
- **Styling**: Tailwind CSS (latest version)
- **State Management**: Zustand
- **UI Components**: Custom-built (no shadcn-ui or third-party UI libraries)
- **Color Palette**: Tailwind default palette for now (will switch to custom palette later)

## Project Structure

- Create a `frontend/` folder at the project root and place all frontend code inside it

## Layout

### Responsiveness

- Desktop-first design with basic mobile usability (no broken layouts on smaller screens)

### Header

- Logo on the left
- User name on the right
- Clicking user name opens a dropdown menu with:
  - Edit Profile
  - Change Password
  - Logout

### Sidebar

- Tree-structured navigation menu
- Structure:
  ```
  - Master Data
    - Product
    - Category
    - Supplier
  - Transaction
    - Sales
    - Purchase
  - Report
    - Sales Report
    - Purchase Report
  ```
- Parent items are collapsible (click to expand/collapse children)
- Sidebar can be toggled (show/hide) via a hamburger/toggle button in the header
- When hidden, the main content area expands to fill the full width

### Footer

- Simple footer with copyright text (e.g., "© 2026 Point of Sale. All rights reserved.")

## Reusable Components

Build these custom components from scratch:

- **Button** - with variants (primary, secondary, danger, outline) and sizes (sm, md, lg)
- **Input** - text input with label, placeholder, and error state
- **Card** - container with optional title and padding
- **Modal** - overlay dialog for forms (used in CRUD operations)
- **Table** - data table with pagination support
- **Toast** - pop-up notification that appears in a corner and auto-dismisses
- **Alert** - inline alert banner (success, error, warning, info variants)
- **Dropdown** - dropdown menu component (used in header user menu)
- **Sidebar Menu** - collapsible tree menu component

## Notification System

- **Toast notifications**: for global action feedback (top-right corner, auto-dismiss after a few seconds)
- **Inline alerts**: for form validation errors (displayed above/within forms)
- Use both together where appropriate

## Authentication

- No route protection or auth guard for phase 1 - all pages are freely accessible
- Login/register/reset password pages are UI-only with simulated behavior

## Phase 1 Pages

### 1. Login (`/login`)

- Fields: email, password
- "Remember me" checkbox
- Login button
- Links to: Register, Reset Password
- On login button click: show a **success toast** message (simulated)

### 2. Register (`/register`)

- Fields: name, email, password, confirm password
- Register button
- Link to: Login
- On register button click: show an **error inline alert** "Email is already registered" (simulated)

### 3. Reset Password (`/reset-password`)

- Fields: email
- Reset Password button
- Link to: Login
- On button click: show a **success toast** "Reset link sent to your email" (simulated)

### 4. Dashboard (`/dashboard`)

- Blank page with layout (header, sidebar, footer) applied
- Placeholder content only (e.g., "Welcome to Dashboard" text)

### 5. Master Category (`/master/category`)

- Full CRUD with search functionality
- **Data source**: mock data from a JSON file or simple variable
- **Table**: displays category list with columns (ID, Name, Description, Actions)
- **Pagination**: paginated table
- **Search**: text search/filter above the table
- **Add**: button opens a **modal dialog** with a form (name, description fields)
- **Edit**: action button on each row opens the same **modal dialog** pre-filled with data
- **Delete**: action button on each row shows a **confirmation modal**, then removes the item
- Show appropriate toast messages on add/edit/delete actions

## General Requirements

- All buttons must be interactive and clickable
- Forms should have basic client-side validation (required fields, email format, password match)
- Use consistent spacing, typography, and color usage across all pages
