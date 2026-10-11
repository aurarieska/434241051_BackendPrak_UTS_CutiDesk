-- Migration 001: Tabel users
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL    PRIMARY KEY,
    username      VARCHAR(50)  NOT NULL UNIQUE,
    email         VARCHAR(100) NOT NULL UNIQUE,
    password      VARCHAR(255) NOT NULL,
    nama_lengkap  VARCHAR(100) NOT NULL,
    role          VARCHAR(50)  NOT NULL DEFAULT 'employee',
    role_id       BIGINT       NOT NULL DEFAULT 3,
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Username unik tanpa membedakan huruf besar/kecil
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_lower ON users (LOWER(username));

-- Index untuk pencarian email
CREATE INDEX IF NOT EXISTS idx_users_email_lower ON users (LOWER(email));

-- Index untuk filter berdasarkan role
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);