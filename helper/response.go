package helper

import (
	"github.com/gofiber/fiber/v2"
	"siplab/app/model"
)

// Success mengirimkan response JSON standar dengan status HTTP kustom.
func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessCursor mengirimkan response daftar berbasis cursor pagination beserta metadata-nya.
func SuccessCursor(c *fiber.Ctx, message string, data any, meta *model.CursorMeta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created mengirimkan response 201 Created beserta header Location (bila diisi).
func Created(c *fiber.Ctx, message string, data any, location string) error {
	if location != "" {
		c.Location(location)
	}
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// NoContent mengirimkan response 204 No Content tanpa body.
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}