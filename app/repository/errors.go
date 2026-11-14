package repository

import "errors"

// Sentinel errors milik layer repository.
// Layer repository TIDAK BOLEH mengetahui status HTTP atau mengimpor fiber.
var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data duplikat")
)
