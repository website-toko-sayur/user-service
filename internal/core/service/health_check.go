package service

import (
	"context"
	"time"
	"user-service/internal/core/domain/entity"

	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type healthCheckService struct {
	redis *redis.Client
	db    *gorm.DB
	kafka sarama.Client
}

type HealthCheckInterface interface {
	HealthCheck(ctx context.Context) (*entity.HealthCheck, error)
}

func NewHealthCheckService(
	redis *redis.Client,
	db *gorm.DB,
	kafka sarama.Client) HealthCheckInterface {
	return &healthCheckService{
		redis: redis,
		db:    db,
		kafka: kafka,
	}
}

var startedAt = time.Now()

func (u *healthCheckService) HealthCheck(ctx context.Context) (*entity.HealthCheck, error) {
	dbStatus := "UP"
	redisStatus := "UP"
	kafkaStatus := "UP"

	// db
	sqlDB, err := u.db.DB()
	if err != nil {
		dbStatus = "DOWN"
	} else if err := sqlDB.PingContext(ctx); err != nil {
		dbStatus = "DOWN"
	}

	// redis
	if err := u.redis.Ping(ctx).Err(); err != nil {
		redisStatus = "DOWN"
	}

	// kafka
	if u.kafka == nil {
		kafkaStatus = "DOWN"
	} else if err := u.kafka.RefreshMetadata(); err != nil {
		kafkaStatus = "DOWN"
	}

	status := "UP"
	if dbStatus == "DOWN" ||
		redisStatus == "DOWN" ||
		kafkaStatus == "DOWN" {
		status = "DOWN"
	}

	return &entity.HealthCheck{
		Status:    status,
		Service:   "user-service",
		Uptime:    time.Since(startedAt).Round(time.Second).String(),
		Timestamp: time.Now().UTC(),
		Dependencies: map[string]string{
			"database": dbStatus,
			"redis":    redisStatus,
			"kafka":    kafkaStatus,
		},
	}, nil
}
