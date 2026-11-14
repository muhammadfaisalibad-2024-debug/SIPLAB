package model

import "time"

// Package model adalah Layer 1 (Entities): TIDAK BOLEH mengimpor package apa pun dari proyek sendiri.

// User merepresentasikan entitas pengguna di dalam sistem SIPLAB.
// PasswordHash diberi tag json:"-" agar tidak pernah bocor saat serialisasi JSON ke client.
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateUserRequest untuk pembuatan user oleh admin (jika diperlukan).
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Email    string `json:"email"    validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,min=8,max=72,nospace,strongpassword"`
}

// ReplaceUserRequest untuk pembaruan penuh via PUT /users/:id.
type ReplaceUserRequest struct {
	Username string `json:"username"  validate:"required,min=3,max=30,username"`
	Email    string `json:"email"     validate:"required,email,max=120"`
	IsActive bool   `json:"is_active"`
}

// PatchUserRequest untuk pembaruan sebagian via PATCH /users/:id.
// Menggunakan pointer dan tag omitnil agar field yang tidak dikirim (nil) dilewati,
// namun field yang dikirim string kosong ("") tetap ditolak oleh min=3.
type PatchUserRequest struct {
	Username *string `json:"username,omitempty"  validate:"omitnil,min=3,max=30,username"`
	Email    *string `json:"email,omitempty"     validate:"omitnil,email,max=120"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// AssignRoleRequest dipakai pada endpoint PATCH /users/:id/role.
type AssignRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin laboran mahasiswa"`
}

// Cursor merepresentasikan penanda posisi pada keyset pagination (created_at + id sebagai tie-breaker unik).
type Cursor struct {
	CreatedAt time.Time
	ID        int
}

// CursorQuery menyimpan parameter filter dan cursor untuk daftar user.
type CursorQuery struct {
	After    *Cursor
	Limit    int
	Search   string
	Role     string
	IsActive *bool
}

// CursorMeta menyimpan metadata pagination berbasis cursor (tanpa Total/TotalPages agar bebas COUNT(*)).
type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// WebResponse adalah amplop standar untuk response berhasil.
type WebResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    any         `json:"data,omitempty"`
	Meta    *CursorMeta `json:"meta,omitempty"`
}

// ErrorResponse adalah amplop tunggal terpusat untuk seluruh response kegagalan.
type ErrorResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}
