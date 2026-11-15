package model

import "time"

// Status peminjaman laboratorium (bookings).
const (
	BookingStatusPending   = "pending"
	BookingStatusApproved  = "approved"
	BookingStatusRejected  = "rejected"
	BookingStatusCancelled = "cancelled"
	BookingStatusCompleted = "completed"
)

// BookingLog mencatat riwayat perubahan status sebuah peminjaman (tabel booking_logs).
type BookingLog struct {
	ID         int       `json:"id"`
	BookingID  int       `json:"booking_id"`
	ActorID    *int      `json:"actor_id,omitempty"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	Note       string    `json:"note,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Booking merepresentasikan entitas peminjaman ruangan atau alat laboratorium di SIPLAB.
type Booking struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	RoomID     *int      `json:"room_id,omitempty"`
	ToolID     *int      `json:"tool_id,omitempty"`
	Quantity   int       `json:"quantity"`
	Purpose    string    `json:"purpose"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Status     string    `json:"status"`
	VerifiedBy *int      `json:"verified_by,omitempty"`
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateBookingRequest dipakai pada POST /api/v1/bookings.
// Mahasiswa mengajukan peminjaman ruangan (room_id) dan/atau alat (tool_id).
type CreateBookingRequest struct {
	RoomID    *int   `json:"room_id,omitempty" validate:"omitnil,min=1"`
	ToolID    *int   `json:"tool_id,omitempty" validate:"omitnil,min=1"`
	Quantity  int    `json:"quantity"          validate:"required,min=1,max=100"`
	Purpose   string `json:"purpose"           validate:"required,min=5,max=255"`
	StartTime string `json:"start_time"        validate:"required"`
	EndTime   string `json:"end_time"          validate:"required"`
}

// VerifyBookingRequest dipakai oleh Laboran/Admin pada PATCH /api/v1/bookings/:id/verify.
type VerifyBookingRequest struct {
	Status string `json:"status" validate:"required,oneof=approved rejected completed"`
	Notes  string `json:"notes"  validate:"max=300"`
}

// BookingFilter menyimpan parameter cursor pagination & filter untuk daftar peminjaman.
type BookingFilter struct {
	After  *Cursor
	Limit  int
	Status string
	UserID *int
}