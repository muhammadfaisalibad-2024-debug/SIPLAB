package model

import "time"

type Room struct {
	ID        int       `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Capacity  int       `json:"capacity"`
	Location  string    `json:"location"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateRoomRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
	Location string `json:"location"`
	Status   string `json:"status"`
}

type UpdateRoomRequest struct {
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
	Location string `json:"location"`
	Status   string `json:"status"`
}