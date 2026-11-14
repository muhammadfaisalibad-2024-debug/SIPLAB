package service

import (
	"testing"
	"time"

	"siplab/app/model"
	"siplab/helper"
)

// Pengujian unit murni: berjalan tanpa menyalakan Fiber dan tanpa koneksi PostgreSQL.

func TestBuildRegisteredUser_ForcesMahasiswaRole(t *testing.T) {
	req := model.RegisterRequest{
		Username: "  faisal_ibad  ",
		Email:    "  Faisal@Vokasi.Unair.ac.id ",
		Password: "PasswordKuat123!",
	}

	user := BuildRegisteredUser(req, "$2a$10$hashedpassword")

	if user.Role != "mahasiswa" {
		t.Fatalf("role pendaftar baru harus 'mahasiswa', dapat %q", user.Role)
	}
	if user.Username != "faisal_ibad" {
		t.Errorf("username seharusnya di-trim menjadi 'faisal_ibad', dapat %q", user.Username)
	}
	if user.Email != "faisal@vokasi.unair.ac.id" {
		t.Errorf("email seharusnya lowercase & di-trim, dapat %q", user.Email)
	}
	if !user.IsActive {
		t.Error("user baru seharusnya aktif (IsActive = true)")
	}
}

func TestCanAccessUser_OwnershipAndPermission(t *testing.T) {
	perms := helper.NewPermissionSet(map[string][]string{
		"admin":     {"user:read:any", "user:update:any"},
		"laboran":   {"user:read:any"},
		"mahasiswa": {},
	})

	mhs := model.AuthUser{UserID: 10, Username: "budi", Role: "mahasiswa"}
	laboran := model.AuthUser{UserID: 20, Username: "sari", Role: "laboran"}
	admin := model.AuthUser{UserID: 1, Username: "root", Role: "admin"}

	// 1. Mahasiswa mengakses datanya sendiri -> BOLEH (ownership)
	if !CanAccessUser(mhs, 10, perms, "user:read:any") {
		t.Error("mahasiswa seharusnya boleh membaca datanya sendiri")
	}
	// 2. Mahasiswa mengakses data orang lain -> DITOLAK (fail closed)
	if CanAccessUser(mhs, 99, perms, "user:read:any") {
		t.Error("mahasiswa tidak boleh membaca data user lain")
	}
	// 3. Laboran membaca data orang lain -> BOLEH (punya user:read:any)
	if !CanAccessUser(laboran, 99, perms, "user:read:any") {
		t.Error("laboran seharusnya boleh membaca data user lain")
	}
	// 4. Laboran mengubah data orang lain -> DITOLAK (tidak punya user:update:any)
	if CanAccessUser(laboran, 99, perms, "user:update:any") {
		t.Error("laboran tidak boleh mengubah data user lain")
	}
	// 5. Admin mengubah data orang lain -> BOLEH
	if !CanAccessUser(admin, 99, perms, "user:update:any") {
		t.Error("admin seharusnya boleh mengubah data user lain")
	}
}

func TestValidateAssignRole_PreventsSelfDemotionAndUnknownRoles(t *testing.T) {
	perms := helper.NewPermissionSet(map[string][]string{
		"admin":     {"role:assign"},
		"laboran":   {},
		"mahasiswa": {},
	})
	admin := model.AuthUser{UserID: 1, Username: "admin_utama", Role: "admin"}

	// Kasus 1: Admin mengubah role dirinya sendiri -> harus gagal
	errsSelf := ValidateAssignRole(admin, 1, model.AssignRoleRequest{Role: "mahasiswa"}, perms)
	if _, exists := errsSelf["role"]; !exists {
		t.Fatal("mengubah role diri sendiri seharusnya menghasilkan error pada field 'role'")
	}

	// Kasus 2: Role tidak terdaftar di database -> harus gagal
	errsUnknown := ValidateAssignRole(admin, 2, model.AssignRoleRequest{Role: "rektor"}, perms)
	if _, exists := errsUnknown["role"]; !exists {
		t.Fatal("role yang tidak dikenal seharusnya menghasilkan error pada field 'role'")
	}

	// Kasus 3: Admin mengubah role user lain ke 'laboran' -> harus lolos
	errsValid := ValidateAssignRole(admin, 2, model.AssignRoleRequest{Role: "laboran"}, perms)
	if len(errsValid) != 0 {
		t.Fatalf("assign role yang sah tidak boleh menghasilkan error, dapat: %v", errsValid)
	}
}

func TestIsRefreshTokenUsable(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	validToken := model.RefreshToken{
		ID:        1,
		UserID:    5,
		ExpiresAt: now.Add(24 * time.Hour),
		RevokedAt: nil,
	}

	if !IsRefreshTokenUsable(validToken, now) {
		t.Error("token yang belum expired dan belum di-revoke seharusnya usable")
	}

	expiredToken := validToken
	expiredToken.ExpiresAt = now.Add(-1 * time.Minute)
	if IsRefreshTokenUsable(expiredToken, now) {
		t.Error("token yang sudah melewati expires_at seharusnya tidak usable")
	}

	revokedTime := now.Add(-10 * time.Minute)
	revokedToken := validToken
	revokedToken.RevokedAt = &revokedTime
	if IsRefreshTokenUsable(revokedToken, now) {
		t.Error("token yang sudah di-revoke seharusnya tidak usable")
	}
}
