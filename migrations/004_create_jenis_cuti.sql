-- Migration 004: Tabel jenis_cuti
CREATE TABLE IF NOT EXISTS jenis_cuti (
    id         SERIAL       PRIMARY KEY,
    nama       VARCHAR(100) NOT NULL,
    deskripsi  TEXT,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Nama jenis cuti unik tanpa membedakan huruf besar/kecil
CREATE UNIQUE INDEX IF NOT EXISTS idx_jenis_cuti_nama_lower ON jenis_cuti (LOWER(nama));