package repository

import (
	"context"
	"errors"
	"time"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/model"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type verificationTokenRepository struct {
	db *gorm.DB
}

type VerificationTokenRepositoryInterface interface {
	CreateVerificationToken(ctx context.Context, req entity.VerificationTokenEntity) error
	GetDataByToken(ctx context.Context, token string) (*entity.VerificationTokenEntity, error)
}

func NewVerificationTokenRepository(db *gorm.DB) VerificationTokenRepositoryInterface {
	return &verificationTokenRepository{db: db}
}

func (v *verificationTokenRepository) GetDataByToken(ctx context.Context, token string) (*entity.VerificationTokenEntity, error) {
	modelToken := model.VerificationToken{}

	if err := v.db.Where("token =?", token).First(&modelToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Info().
				Str("source", "internal.adapter.verificationTokenRepository.GetDataByToken").
				Msg("Role not found")
			return nil, err
		}
		log.Error().
			Err(err).
			Str("source", "internal.adapter.verificationTokenRepository.GetDataByToken")
		return nil, err
	}

	currentTime := time.Now()
	if currentTime.After(modelToken.ExpiresAt) {
		err := errors.New("401")
		log.Error().
			Err(err).
			Str("source", "internal.adapter.verificationTokenRepository.GetDataByToken")
		return nil, err
	}

	return &entity.VerificationTokenEntity{
		ID:        modelToken.ID,
		UserID:    modelToken.UserID,
		Token:     token,
		TokenType: modelToken.TokenType,
		ExpiresAt: modelToken.ExpiresAt,
	}, nil
}

func (v *verificationTokenRepository) CreateVerificationToken(ctx context.Context, req entity.VerificationTokenEntity) error {
	modelVerificationToken := model.VerificationToken{
		UserID:    req.UserID,
		Token:     req.Token,
		TokenType: req.TokenType,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	if err := v.db.Create(&modelVerificationToken).Error; err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.verificationTokenRepository.CreateVerificationToken")
		return err
	}

	return nil
}
