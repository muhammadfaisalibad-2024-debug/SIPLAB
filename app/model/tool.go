package model

import "time"

type Tool struct {
	ID                int       `json:"id"`
	RoomID            int       `json:"room_id"`
	Code              string    `json:"code"`
	Name              string    `json:"name"`
	TotalQuantity     int       `json:"total_quantity"`
	AvailableQuantity int       `json:"available_quantity"`
	Condition         string    `json:"condition"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// RoomAvailability menggabungkan data master ruang dengan hasil pengecekan
// jadwal peminjaman yang sudah disetujui (approved) pada tanggal tertentu.
type RoomAvailability struct {
	Room
	BookedCount int  `json:"booked_count"`
	IsAvailable bool `json:"is_available"`
}

type CreateToolRequest struct {
	RoomID        int    `json:"room_id"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	TotalQuantity int    `json:"total_quantity"`
	Condition     string `json:"condition"`
}

type UpdateToolRequest struct {
	Name              string `json:"name"`
	TotalQuantity     int    `json:"total_quantity"`
	AvailableQuantity int    `json:"available_quantity"`
	Condition         string `json:"condition"`
}