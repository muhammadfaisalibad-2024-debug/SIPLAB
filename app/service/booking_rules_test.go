package service

import (
	"testing"

	"siplab/app/model"
	"siplab/helper"
)

func TestValidateBookingRequest(t *testing.T) {
	roomID := 1
	cases := []struct {
		name    string
		req     model.CreateBookingRequest
		wantErr bool
	}{
		{
			name: "valid peminjaman ruangan",
			req: model.CreateBookingRequest{
				RoomID:    &roomID,
				Quantity:  1,
				Purpose:   "Praktikum Jaringan Komputer",
				StartTime: "2026-11-20T08:00:00Z",
				EndTime:   "2026-11-20T10:00:00Z",
			},
			wantErr: false,
		},
		{
			name: "tanpa room_id maupun tool_id ditolak",
			req: model.CreateBookingRequest{
				Quantity:  1,
				Purpose:   "Praktikum Jaringan Komputer",
				StartTime: "2026-11-20T08:00:00Z",
				EndTime:   "2026-11-20T10:00:00Z",
			},
			wantErr: true,
		},
		{
			name: "end_time sebelum start_time ditolak",
			req: model.CreateBookingRequest{
				RoomID:    &roomID,
				Quantity:  1,
				Purpose:   "Praktikum Jaringan Komputer",
				StartTime: "2026-11-20T10:00:00Z",
				EndTime:   "2026-11-20T08:00:00Z",
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, errs := ValidateBookingRequest(tc.req)
			if (errs != nil) != tc.wantErr {
				t.Errorf("ValidateBookingRequest() errs = %v, wantErr = %v", errs, tc.wantErr)
			}
		})
	}
}

func TestCanAccessBooking_OwnershipAndPermission(t *testing.T) {
	perms := helper.NewPermissionSet(map[string][]string{
		"admin":     {"booking:read:any"},
		"laboran":   {"booking:read:any"},
		"mahasiswa": {"booking:create"},
	})

	mhs1 := model.AuthUser{UserID: 10, Username: "mhs1", Role: "mahasiswa"}
	laboran := model.AuthUser{UserID: 2, Username: "lab1", Role: "laboran"}

	if !CanAccessBooking(mhs1, 10, perms) {
		t.Error("mahasiswa seharusnya boleh melihat booking miliknya sendiri")
	}
	if CanAccessBooking(mhs1, 99, perms) {
		t.Error("mahasiswa tidak boleh melihat booking milik orang lain")
	}
	if !CanAccessBooking(laboran, 99, perms) {
		t.Error("laboran dengan booking:read:any seharusnya boleh melihat booking siapa pun")
	}
}

func TestValidateStatusTransition(t *testing.T) {
	if !ValidateStatusTransition(model.BookingStatusPending, model.BookingStatusApproved) {
		t.Error("pending -> approved seharusnya diizinkan")
	}
	if !ValidateStatusTransition(model.BookingStatusPending, model.BookingStatusRejected) {
		t.Error("pending -> rejected seharusnya diizinkan")
	}
	if !ValidateStatusTransition(model.BookingStatusApproved, model.BookingStatusCompleted) {
		t.Error("approved -> completed seharusnya diizinkan")
	}
	if ValidateStatusTransition(model.BookingStatusRejected, model.BookingStatusApproved) {
		t.Error("rejected -> approved seharusnya ditolak")
	}
	if ValidateStatusTransition(model.BookingStatusCompleted, model.BookingStatusRejected) {
		t.Error("completed -> rejected seharusnya ditolak")
	}
}
func TestValidateStatusTransition_Cancel(t *testing.T) {
	if !ValidateStatusTransition(model.BookingStatusPending, model.BookingStatusCancelled) {
		t.Error("pending -> cancelled seharusnya diizinkan")
	}
	if !ValidateStatusTransition(model.BookingStatusApproved, model.BookingStatusCancelled) {
		t.Error("approved -> cancelled seharusnya diizinkan")
	}
	if ValidateStatusTransition(model.BookingStatusCancelled, model.BookingStatusApproved) {
		t.Error("cancelled -> approved seharusnya ditolak")
	}
}
