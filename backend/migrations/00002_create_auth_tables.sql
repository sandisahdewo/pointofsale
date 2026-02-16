-- +goose Up
-- Add new columns to users table
ALTER TABLE users ADD COLUMN phone VARCHAR(50);
ALTER TABLE users ADD COLUMN address TEXT;
ALTER TABLE users ADD COLUMN password_hash TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN profile_picture TEXT;
ALTER TABLE users ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';
ALTER TABLE users ADD COLUMN is_super_admin BOOLEAN NOT NULL DEFAULT false;

-- Migrate password data to password_hash
UPDATE users SET password_hash = password WHERE password_hash = '';

-- Drop old columns
ALTER TABLE users DROP COLUMN password;
ALTER TABLE users DROP COLUMN role;
ALTER TABLE users DROP COLUMN is_active;

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- Create roles table
CREATE TABLE roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create user_roles join table
CREATE TABLE user_roles (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- Create permissions table
CREATE TABLE permissions (
    id BIGSERIAL PRIMARY KEY,
    module VARCHAR(100) NOT NULL,
    feature VARCHAR(100) NOT NULL,
    actions TEXT[] NOT NULL
);

CREATE UNIQUE INDEX idx_permissions_module_feature ON permissions(module, feature);

-- Create role_permissions table
CREATE TABLE role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    actions TEXT[] NOT NULL,
    UNIQUE(role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);

-- +goose Down
-- Drop tables in reverse order
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;

-- Restore old columns
ALTER TABLE users ADD COLUMN password VARCHAR(255) NOT NULL DEFAULT '';
UPDATE users SET password = password_hash WHERE password = '';
ALTER TABLE users DROP COLUMN password_hash;
ALTER TABLE users ADD COLUMN role VARCHAR(50) NOT NULL DEFAULT 'admin';
ALTER TABLE users ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE users DROP COLUMN phone;
ALTER TABLE users DROP COLUMN address;
ALTER TABLE users DROP COLUMN profile_picture;
ALTER TABLE users DROP COLUMN status;
ALTER TABLE users DROP COLUMN is_super_admin;

-- Drop new indexes
DROP INDEX IF EXISTS idx_users_status;
