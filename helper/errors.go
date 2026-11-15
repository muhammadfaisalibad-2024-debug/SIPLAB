package helper

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"siplab/app/model"
)

// AppError adalah struktur error terpusat untuk seluruh layer service dan middleware.
// Field cause bersifat unexported agar pesan error internal database/server tidak pernah bocor ke client,
// melainkan hanya dicatat di log server.
type AppError struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string
	cause   error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.cause != nil {
		return e.Message + ": " + e.cause.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func BadRequest(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusBadRequest,
		Code:    "BAD_REQUEST",
		Message: message,
	}
}

func Unauthorized(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusUnauthorized,
		Code:    "UNAUTHORIZED",
		Message: message,
	}
}

func Forbidden(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusForbidden,
		Code:    "FORBIDDEN",
		Message: message,
	}
}

func NotFound(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusNotFound,
		Code:    "NOT_FOUND",
		Message: message,
	}
}

func Conflict(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusConflict,
		Code:    "CONFLICT",
		Message: message,
	}
}

func UnsupportedMediaType(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusUnsupportedMediaType,
		Code:    "UNSUPPORTED_MEDIA_TYPE",
		Message: message,
	}
}

func TooManyRequests(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusTooManyRequests,
		Code:    "TOO_MANY_REQUESTS",
		Message: message,
	}
}

func Validation(fields map[string]string) *AppError {
	return &AppError{
		Status:  fiber.StatusUnprocessableEntity,
		Code:    "VALIDATION_ERROR",
		Message: "validasi input gagal",
		Fields:  fields,
	}
}

func ServiceUnavailable(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusServiceUnavailable,
		Code:    "SERVICE_UNAVAILABLE",
		Message: message,
	}
}

func Internal(cause error) *AppError {
	return &AppError{
		Status:  fiber.StatusInternalServerError,
		Code:    "INTERNAL_ERROR",
		Message: "terjadi kesalahan pada server",
		cause:   cause,
	}
}

// ErrorHandler menyediakan global error handler Fiber yang menerjemahkan AppError dan fiber.Error
// menjadi model.ErrorResponse standar.
func ErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		reqID, _ := c.Locals("requestid").(string)

		var appErr *AppError
		if errors.As(err, &appErr) {
			if appErr.Status >= 500 && logger != nil {
				logger.Error("internal_error",
					slog.String("path", c.Path()),
					slog.String("method", c.Method()),
					slog.String("request_id", reqID),
					slog.Any("cause", appErr.Unwrap()),
				)
			}
			return c.Status(appErr.Status).JSON(model.ErrorResponse{
				Success:   false,
				Code:      appErr.Code,
				Message:   appErr.Message,
				Fields:    appErr.Fields,
				RequestID: reqID,
			})
		}

		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			return c.Status(fiberErr.Code).JSON(model.ErrorResponse{
				Success:   false,
				Code:      "HTTP_ERROR",
				Message:   fiberErr.Message,
				RequestID: reqID,
			})
		}

		if logger != nil {
			logger.Error("unhandled_error",
				slog.String("path", c.Path()),
				slog.String("method", c.Method()),
				slog.String("request_id", reqID),
				slog.String("error", err.Error()),
			)
		}

		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success:   false,
			Code:      "INTERNAL_ERROR",
			Message:   "terjadi kesalahan pada server",
			RequestID: reqID,
		})
	}
}