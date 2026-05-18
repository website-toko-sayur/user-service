package response

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type DefaultResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type DefaultResponseWithPaginations struct {
	Message    string      `json:"message"`
	Data       any         `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page       int64 `json:"page"`
	TotalCount int64 `json:"total_count"`
	PerPage    int64 `json:"per_page"`
	TotalPage  int64 `json:"total_page"`
}

func RespondWithError(c fiber.Ctx, code int, context string, err error) error {
	log.Error().
		Err(err).
		Str("context", context).
		// Str("method", c.Method()).
		// Str("path", c.Path()).
		// Str("ip", c.IP()).
		Str("source", "internal.adapter.handler.response.RespondWithError").
		Msg("request failed")

	resp := DefaultResponse{
		Message: err.Error(),
		Data:    nil,
	}

	return c.Status(code).JSON(resp)
}
