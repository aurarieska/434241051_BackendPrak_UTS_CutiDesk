INSERT INTO users (username, email, password, nama_lengkap, role)
VALUES (
    'admin',
    'admin@cutidesk.local',
    '$2a$12$cxfgcN9VialLOd3lIu9Qpetrah7crPo4vTZyX08fEIMoSJtIwcJBC',
    'Administrator',
    'admin'
)
ON CONFLICT DO NOTHING;