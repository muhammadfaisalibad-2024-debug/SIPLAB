package service

import (
	"time"

	"siplab/app/model"
	"siplab/helper"
)

// ValidateBookingRequest memeriksa aturan bisnis murni pengajuan peminjaman:
// 1. Minimal salah satu dari room_id atau tool_id wajib diisi.
// 2. Format waktu wajib RFC3339 dan end_time harus setelah start_time.
func ValidateBookingRequest(req model.CreateBookingRequest) (time.Time, time.Time, map[string]string) {
	errs := make(map[string]string)

	if req.RoomID == nil && req.ToolID == nil {
		errs["target"] = "minimal salah satu dari room_id atau tool_id wajib diisi"
	}

	startTime, errStart := time.Parse(time.RFC3339, req.StartTime)
	if errStart != nil {
		errs["start_time"] = "format waktu harus RFC3339 (contoh: 2026-11-20T08:00:00Z)"
	}

	endTime, errEnd := time.Parse(time.RFC3339, req.EndTime)
	if errEnd != nil {
		errs["end_time"] = "format waktu harus RFC3339 (contoh: 2026-11-20T10:00:00Z)"
	}

	if errStart == nil && errEnd == nil && !endTime.After(startTime) {
		errs["end_time"] = "waktu selesai (end_time) harus lebih besar dari waktu mulai (start_time)"
	}

	if len(errs) > 0 {
		return time.Time{}, time.Time{}, errs
	}
	return startTime, endTime, nil
}

// CanAccessBooking menentukan apakah pengguna berhak melihat detail peminjaman tertentu.
// Mahasiswa hanya boleh melihat miliknya sendiri (ownership), sedangkan Laboran/Admin
// memiliki permission "booking:read:any".
func CanAccessBooking(caller model.AuthUser, ownerUserID int, perms *helper.PermissionSet) bool {
	if caller.UserID == ownerUserID && caller.UserID > 0 {
		return true
	}
	return perms.Can(caller.Role, "booking:read:any")
}

// ValidateStatusTransition memastikan perubahan status verifikasi peminjaman mengikuti alur yang sah:
// - pending -> approved atau rejected
// - approved -> completed
// - rejected / completed adalah status final dan tidak dapat diubah lagi.
func ValidateStatusTransition(currentStatus, nextStatus string) bool {
	switch currentStatus {
	case model.BookingStatusPending:
		return nextStatus == model.BookingStatusApproved || nextStatus == model.BookingStatusRejected
	case model.BookingStatusApproved:
		return nextStatus == model.BookingStatusCompleted
	default:
		return false
	}
}