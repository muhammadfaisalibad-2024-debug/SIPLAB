-- Migration 003: Tabel Peminjaman (Bookings) untuk SIPLAB (Bagian Anggota 3)
CREATE TABLE IF NOT EXISTS bookings (
    id          SERIAL       PRIMARY KEY,
    user_id     INTEGER      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    room_id     INTEGER      NULL,
    tool_id     INTEGER      NULL,
    quantity    INTEGER      NOT NULL DEFAULT 1 CHECK (quantity >= 1),
    purpose     VARCHAR(255) NOT NULL,
    start_time  TIMESTAMPTZ  NOT NULL,
    end_time    TIMESTAMPTZ  NOT NULL,
    status      VARCHAR(20)  NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'completed')),
    verified_by INTEGER      NULL REFERENCES users(id) ON DELETE SET NULL,
    notes       VARCHAR(300) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT bookings_target_check CHECK (room_id IS NOT NULL OR tool_id IS NOT NULL),
    CONSTRAINT bookings_time_check   CHECK (end_time > start_time)
);

CREATE INDEX IF NOT EXISTS bookings_user_id_idx ON bookings(user_id);
CREATE INDEX IF NOT EXISTS bookings_status_idx ON bookings(status);
CREATE INDEX IF NOT EXISTS bookings_created_at_id_desc_idx ON bookings(created_at DESC, id DESC);