-- Migration 004: Tabel jenis_cuti
CREATE TABLE IF NOT EXISTS jenis_cuti (
    id          BIGSERIAL    PRIMARY KEY,
    nama        VARCHAR(100) NOT NULL UNIQUE,
    deskripsi   TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Nama jenis cuti unik tanpa membedakan huruf besar/kecil
CREATE UNIQUE INDEX IF NOT EXISTS idx_jenis_cuti_nama_lower ON jenis_cuti (LOWER(nama));