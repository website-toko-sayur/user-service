package handler

import (
	"encoding/json"
	"strconv"
	"user-service/config"
	"user-service/internal/adapter"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type roleHandler struct {
	roleService service.RoleServiceInterface
}

type RoleHandlerInterface interface {
	GetAll(c fiber.Ctx) error
	GetByID(c fiber.Ctx) error
	Create(c fiber.Ctx) error
	Delete(c fiber.Ctx) error
	Update(c fiber.Ctx) error
}

func NewRoleHandler(
	app *fiber.App,
	roleService service.RoleServiceInterface,
	cfg *config.Config,
	jwtService service.JwtServiceInterface,
	redis *redis.Client,
) RoleHandlerInterface {
	role := &roleHandler{
		roleService: roleService,
	}

	mid := adapter.NewMiddlewareAdapter(cfg, jwtService, redis)

	adminGroup := app.Group("/admin", mid.CheckToken())

	adminGroup.Get("/roles", role.GetAll)
	adminGroup.Post("/roles", role.Create)
	adminGroup.Put("/roles/:id", role.Update)
	adminGroup.Delete("/roles/:id", role.Delete)
	adminGroup.Get("/roles/:id", role.GetByID)

	return role
}

func (r *roleHandler) Create(c fiber.Ctx) error {
	var (
		req         request.RoleRequest
		jwtUserData entity.JwtUserData
	)

	ctx := c.Context()

	user, ok := c.Locals("user").(string)
	if !ok || user == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "data token not found")
	}

	if err := json.Unmarshal([]byte(user), &jwtUserData); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.roleHandler.Create").
			Msg("failed parse jwt user data")

		return fiber.NewError(fiber.StatusBadRequest, "invalid token data")
	}

	if jwtUserData.RoleName != "Super Admin" {
		log.Error().
			Str("role", jwtUserData.RoleName).
			Str("source", "internal.adapter.roleHandler.Create").
			Msg("only super admin can access role api")

		return fiber.NewError(fiber.StatusForbidden, "only Super Admin can access API role")
	}

	if err := c.Bind().Body(&req); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.roleHandler.Create").
			Msg("failed bind/validate request")

		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	roleEntity := entity.RoleEntity{
		Name: req.Name,
	}

	if err := r.roleService.Create(ctx, roleEntity); err != nil {
		log.Error().
			Err(err).
			Str("role_name", req.Name).
			Str("source", "internal.adapter.roleHandler.Create").
			Msg("failed create role")

		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.DefaultResponse{
		Message: "success",
		Data:    nil,
	})
}

func (r *roleHandler) Delete(c fiber.Ctx) error {
	var jwtUserData entity.JwtUserData

	ctx := c.Context()

	user, ok := c.Locals("user").(string)
	if !ok || user == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "data token not found")
	}

	if err := json.Unmarshal([]byte(user), &jwtUserData); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.roleHandler.Delete").
			Msg("failed parse jwt user data")

		return fiber.NewError(fiber.StatusBadRequest, "invalid token data")
	}

	if jwtUserData.RoleName != "Super Admin" {
		log.Error().
			Str("role", jwtUserData.RoleName).
			Str("source", "internal.adapter.roleHandler.Delete").
			Msg("only super admin can access role api")

		return fiber.NewError(fiber.StatusForbidden, "only Super Admin can access API role")
	}

	roleIDString := c.Params("id")
	if roleIDString == "" {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"missing or invalid role ID",
		)
	}

	roleID, err := strconv.Atoi(roleIDString)
	if err != nil {
		log.Error().
			Err(err).
			Str("role_id", roleIDString).
			Str("source", "internal.adapter.roleHandler.Delete").
			Msg("invalid role id")

		return fiber.NewError(fiber.StatusBadRequest, "invalid role ID")
	}

	if err := r.roleService.Delete(ctx, int64(roleID)); err != nil {
		log.Error().
			Err(err).
			Int("role_id", roleID).
			Str("source", "internal.adapter.roleHandler.Delete").
			Msg("failed delete role")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "role not found")
		}

		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "role deleted successfully",
		Data:    nil,
	})
}

func (r *roleHandler) GetAll(c fiber.Ctx) error {
	ctx := c.Context()

	user, ok := c.Locals("user").(string)
	if !ok || user == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "data token not found")
	}

	search := c.Query("search")

	roles, err := r.roleService.GetAll(ctx, search)
	if err != nil {
		log.Error().
			Err(err).
			Str("search", search).
			Str("source", "internal.adapter.roleHandler.GetAll").
			Msg("failed get roles")

		return err
	}

	respRole := make([]response.RoleResponse, 0, len(roles))

	for _, role := range roles {
		respRole = append(respRole, response.RoleResponse{
			ID:   role.ID,
			Name: role.Name,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "success",
		Data:    respRole,
	})
}

func (r *roleHandler) GetByID(c fiber.Ctx) error {
	var jwtUserData entity.JwtUserData

	ctx := c.Context()

	user, ok := c.Locals("user").(string)
	if !ok || user == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "data token not found")
	}

	if err := json.Unmarshal([]byte(user), &jwtUserData); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.roleHandler.GetByID").
			Msg("failed parse jwt user data")

		return fiber.NewError(fiber.StatusBadRequest, "invalid token data")
	}

	if jwtUserData.RoleName != "Super Admin" {
		log.Error().
			Str("role", jwtUserData.RoleName).
			Str("source", "internal.adapter.roleHandler.GetByID").
			Msg("only super admin can access role api")

		return fiber.NewError(fiber.StatusForbidden, "only Super Admin can access API role")
	}

	roleIDString := c.Params("id")
	if roleIDString == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing or invalid role ID")
	}

	roleID, err := strconv.Atoi(roleIDString)
	if err != nil {
		log.Error().
			Err(err).
			Str("role_id", roleIDString).
			Str("source", "internal.adapter.roleHandler.GetByID").
			Msg("invalid role id")

		return fiber.NewError(fiber.StatusBadRequest, "invalid role ID")
	}

	role, err := r.roleService.GetByID(ctx, int64(roleID))
	if err != nil {
		log.Error().
			Err(err).
			Int("role_id", roleID).
			Str("source", "internal.adapter.roleHandler.GetByID").
			Msg("failed get role by id")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "role not found")
		}

		return err
	}

	respRole := response.RoleResponse{
		ID:   role.ID,
		Name: role.Name,
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "success",
		Data:    respRole,
	})
}

func (r *roleHandler) Update(c fiber.Ctx) error {
	var (
		req         request.RoleRequest
		jwtUserData entity.JwtUserData
	)

	ctx := c.Context()

	user, ok := c.Locals("user").(string)
	if !ok || user == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "data token not found")
	}

	if err := json.Unmarshal([]byte(user), &jwtUserData); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.roleHandler.Update").
			Msg("failed parse jwt user data")

		return fiber.NewError(fiber.StatusBadRequest, "invalid token data")
	}

	if jwtUserData.RoleName != "Super Admin" {
		log.Error().
			Str("role", jwtUserData.RoleName).
			Str("source", "internal.adapter.roleHandler.Update").
			Msg("only super admin can access role api")

		return fiber.NewError(fiber.StatusForbidden, "only Super Admin can access API role")
	}

	roleIDString := c.Params("id")
	if roleIDString == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing or invalid role ID")
	}

	roleID, err := strconv.Atoi(roleIDString)
	if err != nil {
		log.Error().
			Err(err).
			Str("role_id", roleIDString).
			Str("source", "internal.adapter.roleHandler.Update").
			Msg("invalid role id")

		return fiber.NewError(fiber.StatusBadRequest, "invalid role ID")
	}

	if err := c.Bind().Body(&req); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.roleHandler.Update").
			Msg("failed bind/validate request")

		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	reqEntity := entity.RoleEntity{
		ID:   int64(roleID),
		Name: req.Name,
	}

	if err := r.roleService.Update(ctx, reqEntity); err != nil {
		log.Error().
			Err(err).
			Int("role_id", roleID).
			Str("source", "internal.adapter.roleHandler.Update").
			Msg("failed update role")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "role not found")
		}

		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "role updated successfully",
		Data:    nil,
	})
}
