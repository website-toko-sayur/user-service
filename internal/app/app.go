package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"user-service/config"
	"user-service/internal/adapter/handler"
	"user-service/internal/adapter/message"
	"user-service/internal/adapter/repository"
	"user-service/internal/adapter/storage"
	middlewareGateway "user-service/internal/middleware"

	"user-service/internal/core/service"

	"github.com/gofiber/fiber/v3"
	fiberCors "github.com/gofiber/fiber/v3/middleware/cors"
	fiberRecover "github.com/gofiber/fiber/v3/middleware/recover"

	"github.com/rs/zerolog/log"
)

func RunServer() {
	cfg := config.NewConfig()

	db, err := cfg.ConnectionPostgres()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("source", "internal.app.RunServer").
			Msg("failed connect postgres")
	}

	minio, err := cfg.NewMinio()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("source", "internal.app.RunServer").
			Msg("failed connect to minio")
	}

	redis, err := cfg.NewRedisClient()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("source", "internal.app.RunServer").
			Msg("failed connect to redis")
	}

	producer := cfg.NewKafkaProducer()

	var (
		emailVerificationProducer   *message.EmailVerificationProducer
		emailForgotPasswordProducer *message.EmailForgotPasswordProducer
		emailCreateCustomerProducer *message.EmailCreateCustomerProducer
		emailUpdateCustomerProducer *message.EmailUpdateCustomerProducer
		pushNotificationProducer    *message.PushNotificationProducer
	)

	if producer != nil {
		emailVerificationProducer = message.NewEmailVerficationProducer(producer)
		emailForgotPasswordProducer = message.NewEmailForgotPasswordProducer(producer)
		emailCreateCustomerProducer = message.NewEmailCreateCustomerProducer(producer)
		emailUpdateCustomerProducer = message.NewEmailUpdateCustomerProducer(producer)
		pushNotificationProducer = message.NewPushNotificationProducer(producer)
	}

	storageHandler := storage.NewMinioStorage(cfg, minio)

	userRepo := repository.NewUserRepository(db.DB)
	tokenRepo := repository.NewVerificationTokenRepository(db.DB)
	roleRepo := repository.NewRoleRepository(db.DB)

	jwtService := service.NewJwtService(cfg)
	userService := service.NewUserService(
		userRepo,
		cfg,
		jwtService,
		tokenRepo,
		redis,
		emailVerificationProducer,
		emailForgotPasswordProducer,
		emailCreateCustomerProducer,
		emailUpdateCustomerProducer,
		pushNotificationProducer,
	)

	roleService := service.NewRoleService(roleRepo)

	app := cfg.NewFiber()
	app.Use(fiberRecover.New())
	app.Use(fiberCors.New())
	app.Use(middlewareGateway.GatewayValidationMiddleware())

	app.Get("/api/check", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	handler.NewUserHandler(app, userService, cfg, jwtService, redis)
	handler.NewUploadImage(app, cfg, storageHandler, jwtService, redis)
	handler.NewRoleHandler(app, roleService, cfg, jwtService, redis)

	go func() {
		if cfg.App.AppPort == "" {
			cfg.App.AppPort = os.Getenv("APP_PORT")
		}

		port := ":" + cfg.App.AppPort

		log.Info().
			Str("port", port).
			Str("source", "internal.app.RunServer").
			Msg("server started")

		err = app.Listen(
			port,
			fiber.ListenConfig{
				EnablePrefork: cfg.App.WebPrefork,
			},
		)

		if err != nil {
			log.Fatal().
				Err(err).
				Str("source", "internal.app.RunServer").
				Msg("failed start server")
		}
	}()

	// =========================
	// Graceful Shutdown
	// =========================
	quit := make(chan os.Signal, 1)

	signal.Notify(
		quit,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-quit

	log.Info().
		Str("source", "internal.app.RunServer").
		Msg("shutting down server in 5 seconds")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.app.RunServer").
			Msg("failed shutdown server")
	}

	log.Info().
		Str("source", "internal.app.RunServer").
		Msg("server stopped gracefully")
}
