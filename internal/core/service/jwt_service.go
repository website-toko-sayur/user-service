package service

import (
	"time"
	"user-service/config"

	"github.com/golang-jwt/jwt/v5"
)

type jwtService struct {
	secretKey string
	issuer    string
}

type JwtServiceInterface interface {
	GenerateToken(userID int64) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
}

func NewJwtService(cfg *config.Config) JwtServiceInterface {
	return &jwtService{
		secretKey: cfg.App.JwtSecretKey,
		issuer:    cfg.App.JwtIssuer,
	}
}

func (j *jwtService) GenerateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"iss":     j.issuer,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *jwtService) ValidateToken(encodetoken string) (*jwt.Token, error) {
	return jwt.Parse(encodetoken, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return []byte(j.secretKey), nil
	})
}
