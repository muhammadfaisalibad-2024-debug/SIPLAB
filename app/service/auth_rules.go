package service

import (
	"strings"
	"time"

	"siplab/app/model"
	"siplab/helper"
)

// File ini berisi BUSINESS RULES MURNI untuk autentikasi, otorisasi, dan manajemen user.
// TIDAK BOLEH mengimpor fiber atau menyentuh database agar dapat diuji langsung (unit test milidetik).

const DefaultRegisterRole = "mahasiswa"

// BuildRegisteredUser menyiapkan entitas User baru dari RegisterRequest.
// Menegakkan aturan bahwa pendaftar mandiri SELALU mendapat role "mahasiswa" dan status aktif.
func BuildRegisteredUser(req model.RegisterRequest, passwordHash string) model.User {
	return model.User{
		Username:     strings.TrimSpace(req.Username),
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash: passwordHash,
		Role:         DefaultRegisterRole,
		IsActive:     true,
	}
}

// IsRefreshTokenUsable memeriksa apakah refresh token belum dicabut dan belum kadaluarsa pada waktu acuan.
func IsRefreshTokenUsable(token model.RefreshToken, now time.Time) bool {
	if token.RevokedAt != nil {
		return false
	}
	return now.Before(token.ExpiresAt)
}

// CanAccessUser memutuskan apakah pemanggil boleh mengakses/mengubah data user dengan targetID.
// Menggabungkan pemeriksaan kepemilikan (ownership) dan RBAC permission (:any).
func CanAccessUser(
	current model.AuthUser,
	targetID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID <= 0 || targetID <= 0 {
		return false
	}
	if current.UserID == targetID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}

// ValidateAssignRole memeriksa aturan bisnis pergantian role yang tidak bisa diwakili tag struct:
// 1. Role tujuan harus terdaftar di PermissionSet database.
// 2. Pengguna (termasuk admin) dilarang mengubah role dirinya sendiri agar sistem tidak kehilangan admin.
func ValidateAssignRole(
	current model.AuthUser,
	targetID int,
	req model.AssignRoleRequest,
	perms *helper.PermissionSet,
) map[string]string {
	errs := map[string]string{}
	role := strings.TrimSpace(req.Role)

	if role == "" {
		errs["role"] = "wajib diisi"
		return errs
	}
	if !perms.IsKnownRole(role) {
		errs["role"] = "role tidak dikenal, pilih salah satu dari: " + strings.Join(perms.KnownRoles(), ", ")
		return errs
	}
	if current.UserID == targetID {
		errs["role"] = "tidak boleh mengubah role diri sendiri"
	}
	return errs
}

// ApplyUserPatch menggabungkan field yang dikirim pada PATCH ke dalam entitas User saat ini.
func ApplyUserPatch(current model.User, req model.PatchUserRequest) model.User {
	if req.Username != nil {
		current.Username = strings.TrimSpace(*req.Username)
	}
	if req.Email != nil {
		current.Email = strings.ToLower(strings.TrimSpace(*req.Email))
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	return current
}

// IsEmptyUserPatch memeriksa apakah request PATCH tidak menyertakan satu pun perubahan field.
func IsEmptyUserPatch(req model.PatchUserRequest) bool {
	return req.Username == nil && req.Email == nil && req.IsActive == nil
}
