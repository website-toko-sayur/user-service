package handler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"user-service/config"
	"user-service/internal/adapter"
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

func NewUserHandler(
	app *fiber.App,
	userService service.UserServiceInterface,
	cfg *config.Config,
	jwtService service.JwtServiceInterface,
) UserHandlerInterface {
	userHandler := &userHandler{
		userService: userService,
	}

	mid := adapter.NewMiddlewareAdapter(cfg, jwtService)

	// =========================
	// Public Routes
	// =========================
	app.Post("/signin", userHandler.SignIn)
	app.Post("/signup", userHandler.CreateUserAccount)
	app.Post("/forgot-password", userHandler.ForgotPassword)
	app.Get("/verify-account", userHandler.VerifyAccount)
	app.Put("/update-password", userHandler.UpdatePassword)

	// =========================
	// Admin Routes
	// =========================
	adminGroup := app.Group("/admin", mid.CheckToken())

	adminGroup.Get("/customers", userHandler.GetCustomerAll)
	adminGroup.Post("/customers", userHandler.CreateCustomer)
	adminGroup.Put("/customers/:id", userHandler.UpdateCustomer)
	adminGroup.Get("/customers/:id", userHandler.GetCustomerByID)
	adminGroup.Delete("/customers/:id", userHandler.DeleteCustomer)

	adminGroup.Get("/check", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// =========================
	// Auth Routes
	// =========================
	authGroup := app.Group("/auth", mid.CheckToken())

	authGroup.Get("/profile", userHandler.GetProfileUser)
	authGroup.Put("/profile", userHandler.UpdateDataUser)

	return userHandler
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

func (u *userHandler) CreateCustomer(c fiber.Ctx) error {
	var req request.CustomerRequest

	ctx := c.Context()

	user, ok := c.Locals("user").(string)
	if !ok || user == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "data token not valid")
	}

	if err := c.Bind().Body(&req); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.CreateCustomer").
			Msg("failed bind/validate request")

		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if req.Password != req.PasswordConfirmation {
		log.Error().
			Str("source", "internal.adapter.userHandler.CreateCustomer").
			Msg("password confirmation mismatch")

		return fiber.NewError(fiber.StatusUnprocessableEntity, "password and confirm password does not match")
	}

	latString := strconv.FormatFloat(req.Lat, 'g', -1, 64)
	lngString := strconv.FormatFloat(req.Lng, 'g', -1, 64)

	reqEntity := entity.UserEntity{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Phone:    req.Phone,
		Address:  req.Address,
		Lat:      latString,
		Lng:      lngString,
		Photo:    req.Photo,
		RoleID:   req.RoleID,
	}

	if err := u.userService.CreateCustomer(ctx, reqEntity); err != nil {
		log.Error().
			Err(err).
			Str("email", req.Email).
			Str("source", "internal.adapter.userHandler.CreateCustomer").
			Msg("failed create customer")

		return fiber.NewError(fiber.StatusInternalServerError, "failed to create customer")
	}

	return c.Status(fiber.StatusCreated).JSON(response.DefaultResponse{
		Message: "success",
		Data:    nil,
	})
}

func (u *userHandler) GetCustomerByID(c fiber.Ctx) error {
	ctx := c.Context()

	user, ok := c.Locals("user").(string)
	if !ok || user == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "data token not valid")
	}

	idParam := c.Params("id")
	if idParam == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id invalid")
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Error().
			Err(err).
			Str("id", idParam).
			Str("source", "internal.adapter.userHandler.GetCustomerByID").
			Msg("invalid customer id")

		return fiber.NewError(fiber.StatusBadRequest, "invalid customer id")
	}

	result, err := u.userService.GetCustomerByID(ctx, id)
	if err != nil {
		log.Error().
			Err(err).
			Int64("customer_id", id).
			Str("source", "internal.adapter.userHandler.GetCustomerByID").
			Msg("failed get customer by id")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "customer not found")
		}

		return err
	}

	respUser := response.CustomerResponse{
		ID:      result.ID,
		RoleID:  result.RoleID,
		Name:    result.Name,
		Email:   result.Email,
		Phone:   result.Phone,
		Address: result.Address,
		Photo:   result.Photo,
		Lat:     result.Lat,
		Lng:     result.Lng,
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "success get customer by id",
		Data:    respUser,
	})
}

func (u *userHandler) GetCustomerAll(c fiber.Ctx) error {
	ctx := c.Context()

	user, ok := c.Locals("user").(string)
	if !ok || user == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "data token not valid")
	}

	search := c.Query("search")

	orderBy := c.Query("order_by", "created_at")

	orderType := c.Query("order_type", "desc")
	if orderType != "asc" && orderType != "desc" {
		orderType = "desc"
	}

	page, err := conv.StringToInt64(c.Query("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := conv.StringToInt64(c.Query("limit", "10"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	reqEntity := entity.QueryStringCustomer{
		Search:    search,
		Page:      page,
		Limit:     limit,
		OrderBy:   orderBy,
		OrderType: orderType,
	}

	results, countData, totalPages, err := u.userService.GetCustomerAll(ctx, reqEntity)
	if err != nil {
		log.Error().
			Err(err).
			Str("search", search).
			Int64("page", page).
			Int64("limit", limit).
			Str("source", "internal.adapter.userHandler.GetCustomerAll").
			Msg("failed get customer list")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "data not found")
		}

		return err
	}

	respUser := make([]response.CustomerListResponse, 0, len(results))

	for _, val := range results {
		respUser = append(respUser, response.CustomerListResponse{
			ID:    val.ID,
			Name:  val.Name,
			Email: val.Email,
			Photo: val.Photo,
			Phone: val.Phone,
		})
	}

	return c.Status(fiber.StatusOK).JSON(
		response.DefaultResponseWithPaginations{
			Message: "data retrieved successfully",
			Data:    respUser,
			Pagination: &response.Pagination{
				Page:       page,
				TotalCount: countData,
				PerPage:    limit,
				TotalPage:  totalPages,
			},
		},
	)
}

func (u *userHandler) UpdateDataUser(c fiber.Ctx) error {
	var (
		req         request.UpdateDataUserRequest
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
			Str("source", "internal.adapter.userHandler.UpdateDataUser").
			Msg("failed parse jwt user data")

		return fiber.NewError(fiber.StatusBadRequest, "invalid token data")
	}

	if err := c.Bind().Body(&req); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.UpdateDataUser").
			Msg("failed bind/validate request")

		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	reqEntity := entity.UserEntity{
		ID:      jwtUserData.UserID,
		Name:    req.Name,
		Email:   req.Email,
		Address: req.Address,
		Lat:     req.Lat,
		Lng:     req.Lng,
		Phone:   req.Phone,
		Photo:   req.Photo,
	}

	if err := u.userService.UpdateDataUser(ctx, reqEntity); err != nil {
		log.Error().
			Err(err).
			Int64("user_id", jwtUserData.UserID).
			Str("source", "internal.adapter.userHandler.UpdateDataUser").
			Msg("failed update user data")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}

		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "success",
		Data:    nil,
	})
}

func (u *userHandler) GetProfileUser(c fiber.Ctx) error {
	var jwtUserData entity.JwtUserData

	ctx := c.Context()

	user, ok := c.Locals("user").(string)
	if !ok || user == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "data token not found")
	}

	if err := json.Unmarshal([]byte(user), &jwtUserData); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.GetProfileUser").
			Msg("failed parse jwt user data")

		return fiber.NewError(fiber.StatusBadRequest, "invalid token data")
	}

	dataUser, err := u.userService.GetProfileUser(ctx, jwtUserData.UserID)
	if err != nil {
		log.Error().
			Err(err).
			Int64("user_id", jwtUserData.UserID).
			Str("source", "internal.adapter.userHandler.GetProfileUser").
			Msg("failed get profile user")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}

		return err
	}

	respProfile := response.ProfileResponse{
		ID:       dataUser.ID,
		Name:     dataUser.Name,
		Email:    dataUser.Email,
		Phone:    dataUser.Phone,
		Address:  dataUser.Address,
		Lat:      dataUser.Lat,
		Lng:      dataUser.Lng,
		Photo:    dataUser.Photo,
		RoleName: dataUser.RoleName,
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "success",
		Data:    respProfile,
	})
}

func (u *userHandler) UpdatePassword(c fiber.Ctx) error {
	var req request.UpdatePasswordRequest

	ctx := c.Context()

	tokenString := c.Query("token")
	if tokenString == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "missing or invalid token")
	}

	if err := c.Bind().Body(&req); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.UpdatePassword").
			Msg("failed bind/validate request")

		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if req.NewPassword != req.ConfirmPassword {
		log.Error().
			Str("source", "internal.adapter.userHandler.UpdatePassword").
			Msg("password confirmation mismatch")

		return fiber.NewError(fiber.StatusUnprocessableEntity, "new password and confirm password does not match")
	}

	reqEntity := entity.UserEntity{
		Password: req.NewPassword,
		Token:    tokenString,
	}

	if err := u.userService.UpdatePassword(ctx, reqEntity); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.UpdatePassword").
			Msg("failed update password")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}

		if err.Error() == "401" {
			return fiber.NewError(fiber.StatusUnauthorized, "token expired or invalid")
		}

		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "password updated successfully",
		Data:    nil,
	})
}

func (u *userHandler) VerifyAccount(c fiber.Ctx) error {
	ctx := c.Context()

	tokenString := c.Query("token")
	if tokenString == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "missing or invalid token")
	}

	user, err := u.userService.VerifyToken(ctx, tokenString)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.VerifyAccount").
			Msg("failed verify account")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}

		if err.Error() == "401" {
			return fiber.NewError(fiber.StatusUnauthorized, "token expired or invalid")
		}

		return err
	}

	respSignIn := response.SignInResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Role:        user.RoleName,
		Lat:         user.Lat,
		Lng:         user.Lng,
		Phone:       user.Phone,
		AccessToken: user.Token,
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "success",
		Data:    respSignIn,
	})
}

func (u *userHandler) ForgotPassword(c fiber.Ctx) error {
	var req request.ForgotPasswordRequest

	ctx := c.Context()

	if err := c.Bind().Body(&req); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.ForgotPassword").
			Msg("failed bind/validate request")

		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	reqEntity := entity.UserEntity{
		Email: req.Email,
	}

	if err := u.userService.ForgotPassword(ctx, reqEntity); err != nil {
		log.Error().
			Err(err).
			Str("email", req.Email).
			Str("source", "internal.adapter.userHandler.ForgotPassword").
			Msg("failed forgot password")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}

		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "success",
		Data:    nil,
	})
}

func (u *userHandler) CreateUserAccount(c fiber.Ctx) error {
	var req request.SignUpRequest

	ctx := c.Context()

	if err := c.Bind().Body(&req); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.CreateUserAccount").
			Msg("failed bind/validate request")

		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if req.Password != req.PasswordConfirmation {
		log.Error().
			Str("source", "internal.adapter.userHandler.CreateUserAccount").
			Msg("password confirmation mismatch")

		return fiber.NewError(fiber.StatusUnprocessableEntity, "passwords do not match")
	}

	reqEntity := entity.UserEntity{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	if err := u.userService.CreateUserAccount(ctx, reqEntity); err != nil {
		log.Error().
			Err(err).
			Str("email", req.Email).
			Str("source", "internal.adapter.userHandler.CreateUserAccount").
			Msg("failed create user account")

		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.DefaultResponse{
		Message: "success",
		Data:    nil,
	})
}

func (u *userHandler) SignIn(c fiber.Ctx) error {
	var req request.SignInRequest

	ctx := c.Context()

	if err := c.Bind().Body(&req); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.userHandler.SignIn").
			Msg("failed bind/validate request")

		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	reqEntity := entity.UserEntity{
		Email:    req.Email,
		Password: req.Password,
	}

	user, token, err := u.userService.SignIn(ctx, reqEntity)
	if err != nil {
		log.Error().
			Err(err).
			Str("email", req.Email).
			Str("source", "internal.adapter.userHandler.SignIn").
			Msg("failed sign in")

		if err.Error() == "404" {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}

		return err
	}

	respSignIn := response.SignInResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Role:        user.RoleName,
		Lat:         user.Lat,
		Lng:         user.Lng,
		Phone:       user.Phone,
		AccessToken: token,
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "success",
		Data:    respSignIn,
	})
}
