package helper

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"siplab/app/model"
)

// EncodeCursor mengubah pasangan (createdAt, id) menjadi string base64 URL-safe.
func EncodeCursor(createdAt time.Time, id int) string {
	raw := fmt.Sprintf("%s|%d", createdAt.UTC().Format(time.RFC3339Nano), id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor menguraikan string base64 cursor kembali menjadi struct model.Cursor.
func DecodeCursor(encoded string) (model.Cursor, error) {
	if strings.TrimSpace(encoded) == "" {
		return model.Cursor{}, errors.New("cursor kosong")
	}

	b, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		b, err = base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return model.Cursor{}, errors.New("format base64 cursor tidak valid")
		}
	}

	parts := strings.Split(string(b), "|")
	if len(parts) != 2 {
		return model.Cursor{}, errors.New("struktur cursor tidak valid")
	}

	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return model.Cursor{}, errors.New("timestamp cursor tidak valid")
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil || id <= 0 {
		return model.Cursor{}, errors.New("id cursor tidak valid")
	}

	return model.Cursor{
		CreatedAt: t,
		ID:        id,
	}, nil
}

// ParseCursorQuery membaca query string (cursor/after, limit, search, role, is_active) dari request.
func ParseCursorQuery(c *fiber.Ctx) (model.CursorQuery, error) {
	q := model.CursorQuery{
		Limit:  10,
		Search: strings.TrimSpace(c.Query("search")),
		Role:   strings.TrimSpace(c.Query("role")),
	}

	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit < 1 || limit > 100 {
			return model.CursorQuery{}, BadRequest("parameter limit harus berupa angka antara 1 sampai 100")
		}
		q.Limit = limit
	}

	rawCursor := strings.TrimSpace(c.Query("cursor"))
	if rawCursor == "" {
		rawCursor = strings.TrimSpace(c.Query("after"))
	}
	if rawCursor != "" {
		cur, err := DecodeCursor(rawCursor)
		if err != nil {
			return model.CursorQuery{}, BadRequest("format cursor tidak valid")
		}
		q.After = &cur
	}

	if rawActive := strings.TrimSpace(c.Query("is_active")); rawActive != "" {
		active, err := strconv.ParseBool(rawActive)
		if err != nil {
			return model.CursorQuery{}, BadRequest("parameter is_active harus berupa boolean (true/false)")
		}
		q.IsActive = &active
	}

	return q, nil
}