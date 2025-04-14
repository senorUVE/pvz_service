package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/senorUVE/pvz_service/internal/models"
)

type AuthService interface {
	GenerateToken(user *models.User) (string, error)
	ParseToken(tokenString string) (*models.User, error)
}

type ServiceAuth struct {
	cfg AuthConfig
}

func NewAuth(cfg AuthConfig) *ServiceAuth {
	return &ServiceAuth{
		cfg: cfg,
	}
}

func (s *ServiceAuth) GenerateToken(user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         user.Id.String(),
		"email":      user.Email,
		"role":       user.Role,
		"expires_at": time.Now().Add(TokenTTL).Unix(),
	})
	return token.SignedString([]byte(s.cfg.SigningKey))
}

func (s *ServiceAuth) ParseToken(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignMethod
		}
		return []byte(s.cfg.SigningKey), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		idStr, ok := claims["id"].(string)
		if !ok {
			return nil, ErrClaimIdFails
		}
		uId, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}

		email, ok := claims["email"].(string)
		if !ok {
			return nil, errors.New("email claim is missing or invalid")
		}
		role, ok := claims["role"].(string)
		if !ok {
			return nil, errors.New("role claim is missing or invalid")
		}
		expiresAt, ok := claims["expires_at"].(float64)
		if !ok || time.Now().After(time.Unix(int64(expiresAt), 0)) {
			return nil, ErrTokenExpired
		}
		return &models.User{
			Id:    uId,
			Email: email,
			Role:  models.Role(role),
		}, nil
	}

	return nil, ErrClaimMissing
}
