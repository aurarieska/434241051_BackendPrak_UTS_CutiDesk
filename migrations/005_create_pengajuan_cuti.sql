-- Migration 005: Tabel pengajuan_cuti
CREATE TABLE IF NOT EXISTS pengajuan_cuti (
    id              SERIAL       PRIMARY KEY,
    user_id         INTEGER      NOT NULL REFERENCES users(id),
    jenis_cuti_id   INTEGER      NOT NULL REFERENCES jenis_cuti(id),
    tanggal_mulai   DATE         NOT NULL,
    tanggal_selesai DATE         NOT NULL,
    alasan          TEXT         NOT NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pengajuan_cuti_user_id ON pengajuan_cuti (user_id);
CREATE INDEX IF NOT EXISTS idx_pengajuan_cuti_status ON pengajuan_cuti (status);
CREATE INDEX IF NOT EXISTS idx_pengajuan_cuti_jenis ON pengajuan_cuti (jenis_cuti_id);