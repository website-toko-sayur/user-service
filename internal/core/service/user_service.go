package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"user-service/config"
	"user-service/internal/adapter/message"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/model"
	"user-service/utils"
	"user-service/utils/conv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type userService struct {
	repo                        repository.UserRepositoryInterface
	cfg                         *config.Config
	jwtService                  JwtServiceInterface
	repoToken                   repository.VerificationTokenRepositoryInterface
	emailVerificationProducer   *message.EmailVerificationProducer
	emailForgotPasswordProducer *message.EmailForgotPasswordProducer
	emailCreateCustomerProducer *message.EmailCreateCustomerProducer
	emailUpdateCustomerProducer *message.EmailUpdateCustomerProducer
	pushNotificationProducer    *message.PushNotificationProducer
	redis                       *redis.Client
}

type UserServiceInterface interface {
	SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error)
	CreateUserAccount(ctx context.Context, req entity.UserEntity) error
	ForgotPassword(ctx context.Context, req entity.UserEntity) error
	VerifyToken(ctx context.Context, token string) (*entity.UserEntity, error)
	UpdatePassword(ctx context.Context, req entity.UserEntity) error
	GetProfileUser(ctx context.Context, userID int64) (*entity.UserEntity, error)
	UpdateDataUser(ctx context.Context, req entity.UserEntity) error

	// Modul Customers Admin
	GetCustomerAll(ctx context.Context, query entity.QueryStringCustomer) ([]entity.UserEntity, int64, int64, error)
	GetCustomerByID(ctx context.Context, customerID int64) (*entity.UserEntity, error)
	CreateCustomer(ctx context.Context, req entity.UserEntity) error
	UpdateCustomer(ctx context.Context, req entity.UserEntity) error
	DeleteCustomer(ctx context.Context, customerID int64) error
}

func NewUserService(repo repository.UserRepositoryInterface, cfg *config.Config,
	jwtService JwtServiceInterface, repoToken repository.VerificationTokenRepositoryInterface,
	redis *redis.Client,
	emailVerificationProducer *message.EmailVerificationProducer,
	emailForgotPasswordProducer *message.EmailForgotPasswordProducer,
	emailCreateCustomerProducer *message.EmailCreateCustomerProducer,
	emailUpdateCustomerProducer *message.EmailUpdateCustomerProducer,
	pushNotificationProducer *message.PushNotificationProducer) UserServiceInterface {
	return &userService{
		repo:                        repo,
		cfg:                         cfg,
		jwtService:                  jwtService,
		repoToken:                   repoToken,
		redis:                       redis,
		emailVerificationProducer:   emailVerificationProducer,
		emailForgotPasswordProducer: emailForgotPasswordProducer,
		emailCreateCustomerProducer: emailCreateCustomerProducer,
		emailUpdateCustomerProducer: emailUpdateCustomerProducer,
		pushNotificationProducer:    pushNotificationProducer,
	}
}

func (u *userService) DeleteCustomer(ctx context.Context, customerID int64) error {
	return u.repo.DeleteCustomer(ctx, customerID)
}

func (u *userService) UpdateCustomer(ctx context.Context, req entity.UserEntity) error {
	passwordNoencrypt := ""
	if req.Password != "" {
		passwordNoencrypt = req.Password
		password, err := conv.HashPassword(req.Password)
		if err != nil {
			log.Error().
				Err(err).
				Str("source", "internal.core.userService.UpdateCustomer")
			return err
		}

		req.Password = password
	}

	err := u.repo.UpdateCustomer(ctx, req)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.UpdateCustomer")
		return err
	}

	if passwordNoencrypt != "" {
		messageparam := fmt.Sprintf("You're account has been updated. Please login use: \n Email: %s\nPassword: %s", req.Email, passwordNoencrypt)
		if u.emailUpdateCustomerProducer != nil {
			event := &model.UserNotificationEvent{
				UserID:    req.ID,
				Recipient: req.Email,
				Subject:   "Updated Data Customer",
				Message:   messageparam,
			}

			log.Info().
				Str("source", "internal.core.userService.UpdateCustomer").
				Msg("Publishing update customer event")

			if err = u.emailUpdateCustomerProducer.Send(event); err != nil {
				log.Warn().
					Err(err).
					Str("source", "internal.core.userService.UpdateCustomer").
					Msg("Failed publish update customer event")
				return err
			}
		} else {
			log.Info().
				Str("source", "internal.core.userService.UpdateCustomer").
				Msg("Kafka producer is disabled, skipping update customer event")
		}

	}

	return nil
}

func (u *userService) CreateCustomer(ctx context.Context, req entity.UserEntity) error {
	passwordNoEncrypt := req.Password
	password, err := conv.HashPassword(passwordNoEncrypt)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.CreateCustomer")
		return err
	}

	req.Password = password
	userID, err := u.repo.CreateCustomer(ctx, req)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.CreateCustomer")
		return err
	}

	messageparam := fmt.Sprintf("You have been registered in Sayur Project. Please login use: \n Email: %s\nPassword: %s", req.Email, passwordNoEncrypt)
	if u.emailCreateCustomerProducer != nil {
		event := &model.UserNotificationEvent{
			UserID:    userID,
			Recipient: req.Email,
			Subject:   "Account Exists",
			Message:   messageparam,
		}

		log.Info().
			Str("source", "internal.core.userService.CreateCustomer").
			Msg("Publishing create customer event")

		if err = u.emailCreateCustomerProducer.Send(event); err != nil {
			log.Warn().
				Err(err).
				Str("source", "internal.core.userService.CreateCustomer").
				Msg("Failed publish create customer event")
			return err
		}
	} else {
		log.Info().
			Str("source", "internal.core.userService.CreateCustomer").
			Msg("Kafka producer is disabled, skipping create customer event")
	}

	return nil
}

func (u *userService) GetCustomerByID(ctx context.Context, customerID int64) (*entity.UserEntity, error) {
	return u.repo.GetCustomerByID(ctx, customerID)
}

func (u *userService) GetCustomerAll(ctx context.Context, query entity.QueryStringCustomer) ([]entity.UserEntity, int64, int64, error) {
	return u.repo.GetCustomerAll(ctx, query)
}

func (u *userService) UpdateDataUser(ctx context.Context, req entity.UserEntity) error {
	return u.repo.UpdateDataUser(ctx, req)
}

func (u *userService) GetProfileUser(ctx context.Context, userID int64) (*entity.UserEntity, error) {
	return u.repo.GetUserByID(ctx, userID)
}

func (u *userService) UpdatePassword(ctx context.Context, req entity.UserEntity) error {
	token, err := u.repoToken.GetDataByToken(ctx, req.Token)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.UpdatePassword")
		return err
	}

	if token.TokenType != "reset_password" {
		err = errors.New("401")
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.UpdatePassword")
		return err
	}

	// validasi expired
	if time.Now().After(token.ExpiresAt) {
		err = errors.New("token expired")
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.UpdatePassword")
		return err
	}

	password, err := conv.HashPassword(req.Password)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.UpdatePassword")
		return err
	}
	req.Password = password
	req.ID = token.UserID

	err = u.repo.UpdatePasswordByID(ctx, req)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.UpdatePassword")
		return err
	}

	return nil
}

func (u *userService) VerifyToken(ctx context.Context, token string) (*entity.UserEntity, error) {
	verifyToken, err := u.repoToken.GetDataByToken(ctx, token)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.VerifyToken")
		return nil, err
	}

	user, err := u.repo.UpdateUserVerified(ctx, verifyToken.UserID)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.VerifyToken")
		return nil, err
	}

	accessToken, err := u.jwtService.GenerateToken(user.ID)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.VerifyToken")
		return nil, err
	}

	sessionData := map[string]any{
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"logged_in":  true,
		"created_at": time.Now().String(),
		"token":      token,
		"role_name":  user.RoleName,
	}

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.VerifyToken").
			Msg("Error encoding JSON")
		return nil, err
	}

	err = u.redis.Set(ctx, token, jsonData, time.Hour*23).Err()
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.VerifyToken").
			Msg("Error set session data to redis")
		return nil, err
	}

	user.Token = accessToken

	return user, nil
}

func (u *userService) ForgotPassword(ctx context.Context, req entity.UserEntity) error {
	user, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.ForgotPassword")
		return err
	}

	token := uuid.New().String()
	reqEntity := entity.VerificationTokenEntity{
		UserID:    user.ID,
		Token:     token,
		TokenType: utils.NOTIF_EMAIL_FORGOT_PASSWORD,
	}

	err = u.repoToken.CreateVerificationToken(ctx, reqEntity)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.ForgotPassword")
		return err
	}

	urlForgot := fmt.Sprintf("%s/auth/update-password?token=%s", u.cfg.App.UrlFrontFE, token)
	messageparam := fmt.Sprintf("Please click link below for reset password: %v", urlForgot)
	if u.emailForgotPasswordProducer != nil {
		event := &model.UserNotificationEvent{
			UserID:    user.ID,
			Recipient: req.Email,
			Subject:   "Reset password",
			Message:   messageparam,
		}

		log.Info().
			Str("source", "internal.core.userService.ForgotPassword").
			Msg("Publishing forgot password event")

		if err = u.emailForgotPasswordProducer.Send(event); err != nil {
			log.Warn().
				Err(err).
				Str("source", "internal.core.userService.ForgotPassword").
				Msg("Failed publish forgot password event")
			return err
		}
	} else {
		log.Info().
			Str("source", "internal.core.userService.CreateCustomer").
			Msg("Kafka producer is disabled, skipping forgot password event")
	}

	return nil
}

func (u *userService) CreateUserAccount(ctx context.Context, req entity.UserEntity) error {
	password, err := conv.HashPassword(req.Password)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.CreateUserAccount")
		return err
	}

	req.Password = password
	req.Token = uuid.New().String()

	userID, err := u.repo.CreateUserAccount(ctx, req)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.CreateUserAccount")
		return err
	}

	verifyURL := fmt.Sprintf("%s/auth/verify-account?token=%s", u.cfg.App.UrlFrontFE, req.Token)
	verifyMsg := fmt.Sprintf("Please verify your account by clicking the link: %s", verifyURL)

	if u.emailVerificationProducer != nil {
		event := &model.UserNotificationEvent{
			UserID:    userID,
			Recipient: req.Email,
			Subject:   "Verify your account",
			Message:   verifyMsg,
		}

		log.Info().
			Str("source", "internal.core.userService.CreateUserAccount").
			Msg("Publishing email verification create user account event")

		if err = u.emailVerificationProducer.Send(event); err != nil {
			log.Warn().
				Err(err).
				Str("source", "internal.core.userService.CreateUserAccount").
				Msg("Failed publish email verification create user account event")
			return err
		}
	} else {
		log.Info().
			Str("source", "internal.core.userService.CreateUserAccount").
			Msg("Kafka producer is disabled, skipping email verification create user account event")
	}

	return nil
}

func (u *userService) SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error) {
	user, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.SignIn")
		return nil, "", err
	}

	if checkPass := conv.CheckPasswordHash(req.Password, user.Password); !checkPass {
		err = errors.New("password is incorrect")
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.SignIn")
		return nil, "", err
	}

	token, err := u.jwtService.GenerateToken(user.ID)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.SignIn")
		return nil, "", err
	}

	sessionData := map[string]interface{}{
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"logged_in":  true,
		"created_at": time.Now().String(),
		"token":      token,
		"role_name":  user.RoleName,
	}

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.SignIn").
			Msg("Error encoding JSON")
		return nil, "", err
	}

	err = u.redis.Set(ctx, token, jsonData, time.Hour*23).Err()
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.core.userService.SignIn")
		return nil, "", err
	}

	return user, token, nil
}
