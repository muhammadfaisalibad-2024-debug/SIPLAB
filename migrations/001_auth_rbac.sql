-- 1. Tabel roles: menyimpan daftar peran yang sah di dalam sistem SIPLAB
CREATE TABLE IF NOT EXISTS roles (
    name        VARCHAR(20)  PRIMARY KEY,
    description VARCHAR(150) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Seed 3 role utama SIPLAB
INSERT INTO roles (name, description) VALUES
    ('admin',     'Akses penuh ke seluruh manajemen user, ruang, alat, dan peminjaman'),
    ('laboran',   'Mengelola inventaris ruang, alat lab, dan memverifikasi peminjaman'),
    ('mahasiswa', 'Melihat katalog ruang/alat dan mengajukan peminjaman milik sendiri')
ON CONFLICT (name) DO NOTHING;

-- 2. Tabel permissions: daftar aksi granular berpola domain:action[:scope]
CREATE TABLE IF NOT EXISTS permissions (
    name        VARCHAR(50)  PRIMARY KEY,
    description VARCHAR(150) NOT NULL
);

-- Seed permission untuk domain user, role, room (ruang), tool (alat), dan booking (peminjaman)
INSERT INTO permissions (name, description) VALUES
    -- Domain User & Role
    ('user:list',         'Melihat daftar seluruh user'),
    ('user:read:any',     'Melihat detail profil user mana pun'),
    ('user:update:any',   'Mengubah data user mana pun'),
    ('user:delete',       'Menghapus akun user lain'),
    ('role:assign',       'Mengubah role milik user lain'),
    
    -- Domain Ruang Lab (Rooms)
    ('room:create',       'Menambahkan data ruang lab baru'),
    ('room:update',       'Mengubah data ruang lab'),
    ('room:delete',       'Menghapus data ruang lab'),
    
    -- Domain Alat Lab (Tools)
    ('tool:create',       'Menambahkan data alat lab baru'),
    ('tool:update',       'Mengubah data dan stok alat lab'),
    ('tool:delete',       'Menghapus data alat lab'),
    
    -- Domain Peminjaman (Bookings)
    ('booking:create',    'Mengajukan peminjaman ruang atau alat'),
    ('booking:list:any',  'Melihat daftar seluruh peminjaman semua user'),
    ('booking:read:any',  'Melihat detail peminjaman milik siapa pun'),
    ('booking:verify',    'Menyetujui, menolak, atau menyelesaikan peminjaman')
ON CONFLICT (name) DO NOTHING;

-- 3. Tabel penghubung role_permissions dengan Composite Primary Key
CREATE TABLE IF NOT EXISTS role_permissions (
    role_name       VARCHAR(20) NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

-- Pemetaan permission untuk Admin (memiliki seluruh hak akses manajemen & operasional)
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'user:list'),
    ('admin', 'user:read:any'),
    ('admin', 'user:update:any'),
    ('admin', 'user:delete'),
    ('admin', 'role:assign'),
    ('admin', 'room:create'),
    ('admin', 'room:update'),
    ('admin', 'room:delete'),
    ('admin', 'tool:create'),
    ('admin', 'tool:update'),
    ('admin', 'tool:delete'),
    ('admin', 'booking:create'),
    ('admin', 'booking:list:any'),
    ('admin', 'booking:read:any'),
    ('admin', 'booking:verify')
ON CONFLICT DO NOTHING;

-- Pemetaan permission untuk Laboran (fokus pada operasional lab & verifikasi booking)
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('laboran', 'user:list'),
    ('laboran', 'user:read:any'),
    ('laboran', 'room:create'),
    ('laboran', 'room:update'),
    ('laboran', 'tool:create'),
    ('laboran', 'tool:update'),
    ('laboran', 'booking:list:any'),
    ('laboran', 'booking:read:any'),
    ('laboran', 'booking:verify')
ON CONFLICT DO NOTHING;

-- Pemetaan permission untuk Mahasiswa (hanya pengajuan booking; akses profil/booking sendiri diatur via ownership di service)
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('mahasiswa', 'booking:create')
ON CONFLICT DO NOTHING;

-- 4. Tabel users dengan Foreign Key langsung ke roles(name)
CREATE TABLE IF NOT EXISTS users (
    id            SERIAL       PRIMARY KEY,
    username      VARCHAR(30)  NOT NULL UNIQUE,
    email         VARCHAR(120) NOT NULL UNIQUE,
    password_hash VARCHAR(100) NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'mahasiswa' REFERENCES roles(name) ON UPDATE CASCADE,
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Index untuk pencarian role dan keyset/cursor pagination (wajib sama dengan ORDER BY created_at DESC, id DESC)
CREATE INDEX IF NOT EXISTS users_role_idx ON users(role);
CREATE INDEX IF NOT EXISTS users_created_at_id_desc_idx ON users(created_at DESC, id DESC);

-- 5. Tabel refresh_tokens untuk menyimpan sesi refresh token yang dapat dicabut (revocable)
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         SERIAL       PRIMARY KEY,
    user_id    INTEGER      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64)  NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ  NOT NULL,
    revoked_at TIMESTAMPTZ  NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx ON refresh_tokens(user_id);
