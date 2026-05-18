package handler

import (
	"fmt"
	"strconv"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"
	"user-service/utils/conv"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type userHandler struct {
	userService service.UserServiceInterface
}

type UserHandlerInterface interface {
	SignIn(c fiber.Ctx) error
	CreateUserAccount(c fiber.Ctx) error
	ForgotPassword(c fiber.Ctx) error
	VerifyAccount(c fiber.Ctx) error
	UpdatePassword(c fiber.Ctx) error
	GetProfileUser(c fiber.Ctx) error
	UpdateDataUser(c fiber.Ctx) error

	// Modul Customers Admin
	GetCustomerAll(c fiber.Ctx) error
	GetCustomerByID(c fiber.Ctx) error
	CreateCustomer(c fiber.Ctx) error
	UpdateCustomer(c fiber.Ctx) error
	DeleteCustomer(c fiber.Ctx) error
}

func (u *userHandler) DeleteCustomer(c fiber.Ctx) error {
	ctx := c.Context()

	user, ok := c.Locals("user").(string)
	if !ok || user == "" {
		log.Error().
			Str("source", "internal.adapter.userHandler.DeleteCustomer").
			Msg("data token not found")

		return fiber.NewError(fiber.StatusUnauthorized, "data token not valid")
	}

	idParamStr := c.Params("id")
	if idParamStr == "" {
		log.Info().
			Str("source", "internal.adapter.userHandler.DeleteCustomer").
			Msg("missing or invalid customer ID")

		return fiber.NewError(fiber.StatusBadRequest, "missing or invalid customer ID")
	}

	id, err := conv.StringToInt64(idParamStr)
	if err != nil {
		log.Info().
			Err(err).
			Str("id", idParamStr).
			Str("source", "internal.adapter.userHandler.DeleteCustomer").
			Msg("invalid customer ID")

		return fiber.NewError(fiber.StatusBadRequest, "Invalid customer ID")
	}

	err = u.userService.DeleteCustomer(ctx, id)
	if err != nil {
		log.Error().
			Err(err).
			Int64("customer_id", id).
			Str("source", "internal.adapter.userHandler.DeleteCustomer").
			Msg("failed delete customer")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "Customer not found")
		}

		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "customer deleted successfully",
		Data:    nil,
	})
}

func (u *userHandler) UpdateCustomer(c fiber.Ctx) error {
	var req request.CustomerRequest

	ctx := c.Context()

	user, ok := c.Locals("user").(string)
	if !ok || user == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "data token not valid")
	}

	if err := c.Bind().Body(&req); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.UpdateCustomer").
			Msg("failed bind/validate request")

		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	idParamStr := c.Params("id")
	if idParamStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing or invalid customer ID")
	}

	id, err := conv.StringToInt64(idParamStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("id", idParamStr).
			Str("source", "internal.adapter.userHandler.UpdateCustomer").
			Msg("invalid customer ID")

		return fiber.NewError(fiber.StatusBadRequest, "invalid customer ID")
	}

	latString := ""
	lngString := ""

	if req.Lat != 0 {
		latString = strconv.FormatFloat(req.Lat, 'g', -1, 64)
	}

	if req.Lng != 0 {
		lngString = strconv.FormatFloat(req.Lng, 'g', -1, 64)
	}

	phoneString := fmt.Sprintf("%v", req.Phone)

	reqEntity := entity.UserEntity{
		ID:       id,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Phone:    phoneString,
		Address:  req.Address,
		Lat:      latString,
		Lng:      lngString,
		Photo:    req.Photo,
	}

	if err := u.userService.UpdateDataUser(ctx, reqEntity); err != nil {
		log.Error().
			Err(err).
			Int64("customer_id", id).
			Str("source", "internal.adapter.userHandler.UpdateCustomer").
			Msg("failed update customer")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "customer not found")
		}

		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "success",
		Data:    nil,
	})
}
