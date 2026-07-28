-- ============================================================
-- Template: Tambahkan Role Baru (SQL Biasa)
-- ============================================================

-- 1. Tambah Role
INSERT INTO roles (name, description) 
VALUES ('nama_role_baru', 'Deskripsi untuk role baru');

-- 2. Tambah Permission (jika belum ada)
INSERT INTO permissions (name) 
VALUES ('permission.baru');

-- 3. Hubungkan Role dengan Permission
-- Pastikan ganti (SELECT id FROM roles WHERE name = 'nama_role_baru')
-- dengan ID yang sesuai jika diperlukan.
INSERT INTO role_permissions (role_id, permission)
VALUES (
    (SELECT id FROM roles WHERE name = 'nama_role_baru'), 
    'permission.baru'
);
