-- Migration 003: Tabel RBAC (roles, permissions, role_permissions)

CREATE TABLE IF NOT EXISTS roles (
    name        VARCHAR(20)  PRIMARY KEY,
    description VARCHAR(150) NOT NULL,
    created_at  TIMESTAMPTZ  DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS permissions (
    name        VARCHAR(50)  PRIMARY KEY,
    description VARCHAR(150) NOT NULL
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_name       VARCHAR(20) REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50) REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

-- Tambahkan FK dari users.role ke roles.name
ALTER TABLE users
    ADD CONSTRAINT fk_users_role FOREIGN KEY (role) REFERENCES roles(name);

-- Seed: tiga role utama
INSERT INTO roles (name, description) VALUES
    ('admin',    'Administrator sistem'),
    ('hr',       'Human Resources'),
    ('employee', 'Karyawan')
ON CONFLICT DO NOTHING;

-- Seed: seluruh permission
INSERT INTO permissions (name, description) VALUES
    ('user:list',         'Melihat daftar seluruh user'),
    ('user:read:any',     'Melihat data user mana pun'),
    ('user:create',       'Membuat user baru'),
    ('user:update:any',   'Mengubah data user mana pun'),
    ('user:delete',       'Menghapus user'),
    ('role:assign',       'Mengubah role user lain'),
    ('jenis_cuti:list',   'Melihat daftar jenis cuti'),
    ('jenis_cuti:read',   'Melihat detail satu jenis cuti'),
    ('jenis_cuti:create', 'Menambah jenis cuti baru'),
    ('jenis_cuti:update', 'Mengubah jenis cuti'),
    ('cuti:create',       'Mengajukan cuti'),
    ('cuti:list:any',     'Melihat seluruh pengajuan cuti semua user'),
    ('cuti:read:any',     'Melihat detail pengajuan cuti milik user lain'),
    ('cuti:approve',      'Menyetujui atau menolak pengajuan cuti')
ON CONFLICT DO NOTHING;

-- Pemetaan: admin
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'user:list'),
    ('admin', 'user:read:any'),
    ('admin', 'user:create'),
    ('admin', 'user:update:any'),
    ('admin', 'user:delete'),
    ('admin', 'role:assign'),
    ('admin', 'jenis_cuti:list'),
    ('admin', 'jenis_cuti:read'),
    ('admin', 'jenis_cuti:create'),
    ('admin', 'jenis_cuti:update'),
    ('admin', 'cuti:list:any'),
    ('admin', 'cuti:read:any')
ON CONFLICT DO NOTHING;

-- Pemetaan: hr
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('hr', 'jenis_cuti:list'),
    ('hr', 'jenis_cuti:read'),
    ('hr', 'cuti:list:any'),
    ('hr', 'cuti:read:any'),
    ('hr', 'cuti:approve')
ON CONFLICT DO NOTHING;

-- Pemetaan: employee
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('employee', 'jenis_cuti:list'),
    ('employee', 'jenis_cuti:read'),
    ('employee', 'cuti:create')
ON CONFLICT DO NOTHING;