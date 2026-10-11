-- =====================================================
-- Migration 003: RBAC - Roles & Permissions
-- =====================================================

-- Tabel roles
CREATE TABLE IF NOT EXISTS roles (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(50)  NOT NULL UNIQUE,
    description VARCHAR(255),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Tabel permissions
CREATE TABLE IF NOT EXISTS permissions (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(255),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Tabel pivot role_permissions
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- =====================================================
-- Seed 3 roles
-- =====================================================
INSERT INTO roles (name, description) VALUES
    ('admin',    'Administrator sistem'),
    ('hr',       'Human Resources'),
    ('employee', 'Karyawan biasa')
ON CONFLICT (name) DO NOTHING;

-- =====================================================
-- Seed 14 permissions (HARUS SAMA dengan config/app.go)
-- =====================================================
INSERT INTO permissions (name, description) VALUES
    ('user:create',       'Membuat user baru'),
    ('user:read',         'Melihat daftar dan detail user'),
    ('user:update',       'Mengubah data user'),
    ('user:delete',       'Menghapus/menonaktifkan user'),
    ('jenis_cuti:create', 'Membuat jenis cuti baru'),
    ('jenis_cuti:read',   'Melihat daftar jenis cuti'),
    ('jenis_cuti:update', 'Mengubah jenis cuti'),
    ('jenis_cuti:delete', 'Menghapus jenis cuti'),
    ('cuti:create',       'Mengajukan cuti'),
    ('cuti:read',         'Melihat daftar pengajuan cuti'),
    ('cuti:approve',      'Menyetujui atau menolak cuti'),
    ('riwayat:read',      'Melihat riwayat cuti'),
    ('laporan:read',      'Melihat laporan'),
    ('profil:update',     'Mengubah profil sendiri')
ON CONFLICT (name) DO NOTHING;

-- =====================================================
-- Assign permissions ke roles
-- =====================================================

-- Admin: semua 14 permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

-- HR: 12 permissions (semua kecuali user:delete dan cuti:create)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'hr'
  AND p.name IN (
    'user:create', 'user:read', 'user:update',
    'jenis_cuti:create', 'jenis_cuti:read', 'jenis_cuti:update', 'jenis_cuti:delete',
    'cuti:read', 'cuti:approve',
    'riwayat:read', 'laporan:read', 'profil:update'
  )
ON CONFLICT DO NOTHING;

-- Employee: 4 permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'employee'
  AND p.name IN (
    'cuti:create', 'cuti:read',
    'riwayat:read', 'profil:update'
  )
ON CONFLICT DO NOTHING;