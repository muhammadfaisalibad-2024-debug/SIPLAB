-- Skema Tabel Ruangan (rooms)
CREATE TABLE IF NOT EXISTS rooms (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    capacity INT NOT NULL DEFAULT 0,
    location VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'AVAILABLE', -- AVAILABLE, MAINTENANCE, UNAVAILABLE
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Skema Tabel Peralatan/Alat Lab (tools)
CREATE TABLE IF NOT EXISTS tools (
    id SERIAL PRIMARY KEY,
    room_id INT REFERENCES rooms(id) ON DELETE CASCADE,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    total_quantity INT NOT NULL DEFAULT 0,
    available_quantity INT NOT NULL DEFAULT 0,
    condition VARCHAR(20) NOT NULL DEFAULT 'GOOD', -- GOOD, REPAIR, BROKEN
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seeder Data Awal Ruangan
INSERT INTO rooms (code, name, capacity, location, status) VALUES
('LAB-AJ-01', 'Laboratorium Arsitektur Jaringan', 30, 'Gedung A Lt. 2', 'AVAILABLE'),
('LAB-RPL-01', 'Laboratorium Rekayasa Perangkat Lunak', 40, 'Gedung A Lt. 3', 'AVAILABLE')
ON CONFLICT (code) DO NOTHING;

-- Seeder Data Awal Alat
INSERT INTO tools (room_id, code, name, total_quantity, available_quantity, condition) VALUES
(1, 'TL-RTR-01', 'Router Cisco 2911', 10, 10, 'GOOD'),
(1, 'TL-SWT-01', 'Switch Cisco 2960', 15, 15, 'GOOD'),
(2, 'TL-MNTR-01', 'Monitor LG 24 inch', 20, 20, 'GOOD')
ON CONFLICT (code) DO NOTHING;