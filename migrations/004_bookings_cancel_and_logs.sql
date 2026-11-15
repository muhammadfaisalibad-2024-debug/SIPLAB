-- Migration 004: dukungan status 'cancelled' + tabel audit booking_logs
-- Status peminjaman SIPLAB: pending, approved, rejected, cancelled, completed.

-- 1. Perbarui constraint status agar mahasiswa dapat membatalkan peminjamannya sendiri.
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_status_check;
ALTER TABLE bookings ADD CONSTRAINT bookings_status_check
    CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled', 'completed'));

-- 2. Tabel booking_logs: mencatat setiap perubahan status beserta aktornya.
CREATE TABLE IF NOT EXISTS booking_logs (
    id          SERIAL       PRIMARY KEY,
    booking_id  INTEGER      NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    actor_id    INTEGER      NULL REFERENCES users(id) ON DELETE SET NULL,
    from_status VARCHAR(20)  NOT NULL,
    to_status   VARCHAR(20)  NOT NULL,
    note        VARCHAR(300) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS booking_logs_booking_id_idx ON booking_logs(booking_id, created_at DESC);
