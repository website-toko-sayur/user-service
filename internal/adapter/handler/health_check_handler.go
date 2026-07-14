package handler

import (
	"net/http"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/service"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type healthCheckHandler struct {
	healthCheckService service.HealthCheckInterface
}

type HealthCheckHandlerInterface interface {
	HealthCheck(c fiber.Ctx) error
}

func NewHealthCheckHandler(
	app *fiber.App,
	healthCheckService service.HealthCheckInterface,
) HealthCheckHandlerInterface {
	healthCheckHandler := &healthCheckHandler{
		healthCheckService: healthCheckService,
	}

	app.Get("/health", healthCheckHandler.HealthCheck)

	return healthCheckHandler
}

func (u *healthCheckHandler) HealthCheck(c fiber.Ctx) error {
	ctx := c.Context()

	result, err := u.healthCheckService.HealthCheck(ctx)
	if err != nil || result.Status == "DOWN" {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.handler.healthCheckHandler.HealthCheck").
			Str("service", result.Service).
			Str("status", result.Status).
			Interface("dependencies", result.Dependencies).
			Msg("health check failed")

		// return fiber.NewError(fiber.StatusServiceUnavailable, "failed to health check")
	}

	log.Info().
		Str("source", "internal.adapter.handler.healthCheckHandler.HealthCheck").
		Str("service", result.Service).
		Str("status", result.Status).
		Interface("dependencies", result.Dependencies).
		Msg("health check completed successfully")

	return c.Status(http.StatusOK).JSON(response.DefaultResponse{
		Message: "success",
		Data:    result,
	})
}
