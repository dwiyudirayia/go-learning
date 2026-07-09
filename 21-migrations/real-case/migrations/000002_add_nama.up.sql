-- EXPAND: kolom baru dibuat NULLABLE agar aman untuk kode versi lama (zero-downtime).
ALTER TABLE users ADD COLUMN nama TEXT;
