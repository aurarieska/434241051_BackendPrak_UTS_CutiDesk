-- Migration 006: Tabel riwayat_status
CREATE TABLE IF NOT EXISTS riwayat_status (
    id                SERIAL       PRIMARY KEY,
    pengajuan_cuti_id INTEGER      NOT NULL REFERENCES pengajuan_cuti(id) ON DELETE CASCADE,
    status_sebelum    VARCHAR(20),
    status_sesudah    VARCHAR(20)  NOT NULL,
    catatan           TEXT,
    diubah_oleh       INTEGER      NOT NULL REFERENCES users(id),
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_riwayat_status_pengajuan ON riwayat_status (pengajuan_cuti_id);