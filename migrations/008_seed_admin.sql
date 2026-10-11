INSERT INTO users (username, email, password, nama_lengkap, role, role_id, is_active)
VALUES (
    'admin',
    'admin@cutidesk.local',
    '$2a$12$SqM7JPwf2.Lxl4QkrRt.vu83EVKGdapbIgZuEqQ/cW/TxX2MrSpEy',
    'Administrator',
    'admin',
    (SELECT id FROM roles WHERE name = 'admin'),
    TRUE
)
ON CONFLICT (username) DO NOTHING;