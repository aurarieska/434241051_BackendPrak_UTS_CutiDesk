-- Migration 007: Index komposit untuk cursor/keyset pagination
-- Urutan dan arah HARUS sama dengan ORDER BY di query: created_at DESC, id DESC

CREATE INDEX IF NOT EXISTS idx_users_cursor
    ON users (created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_jenis_cuti_cursor
    ON jenis_cuti (created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_pengajuan_cuti_cursor
    ON pengajuan_cuti (created_at DESC, id DESC);