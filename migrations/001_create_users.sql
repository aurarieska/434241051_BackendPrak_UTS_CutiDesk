-- Migration 001: Tabel users
CREATE TABLE IF NOT EXISTS users (
    id            SERIAL       PRIMARY KEY,
    username      VARCHAR(50)  NOT NULL,
    email         VARCHAR(255) NOT NULL,
    password      VARCHAR(255) NOT NULL,
    nama_lengkap  VARCHAR(150) NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'employee',
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Username unik tanpa membedakan huruf besar/kecil
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_lower ON users (LOWER(username));

-- Index untuk pencarian email
CREATE INDEX IF NOT EXISTS idx_users_email_lower ON users (LOWER(email));

-- Index untuk filter berdasarkan role
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);