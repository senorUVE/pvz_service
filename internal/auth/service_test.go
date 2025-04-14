package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/senorUVE/pvz_service/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestServiceAuth_GenerateToken(t *testing.T) {
	cfg := AuthConfig{
		SigningKey: "test-key",
	}
	service := NewAuth(cfg)

	userID := uuid.New()
	user := &models.User{
		Id:    userID,
		Email: "test@example.com",
		Role:  models.RoleEmployee,
	}

	token, err := service.GenerateToken(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.SigningKey), nil
	})
	assert.NoError(t, err)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, user.Id.String(), claims["id"])
	assert.Equal(t, user.Email, claims["email"])
	assert.Equal(t, string(user.Role), claims["role"])
	assert.InDelta(t, time.Now().Add(TokenTTL).Unix(), claims["expires_at"].(float64), 1)
}

func TestServiceAuth_ParseToken_ValidToken(t *testing.T) {
	cfg := AuthConfig{
		SigningKey: "test-key",
	}
	service := NewAuth(cfg)

	userID := uuid.New()
	user := &models.User{
		Id:    userID,
		Email: "test@example.com",
		Role:  models.RoleModerator,
	}

	token, err := service.GenerateToken(user)
	assert.NoError(t, err)

	parsedUser, err := service.ParseToken(token)
	assert.NoError(t, err)
	assert.Equal(t, user.Id, parsedUser.Id)
	assert.Equal(t, user.Email, parsedUser.Email)
	assert.Equal(t, user.Role, parsedUser.Role)
}

func TestServiceAuth_ParseToken_Invalid(t *testing.T) {
	cfg := AuthConfig{
		SigningKey: "test-key",
	}
	service := NewAuth(cfg)

	invalidToken := "eyJhbGciOiJIUzI1NiIsI"

	_, err := service.ParseToken(invalidToken)
	assert.Error(t, err)
}

func TestServiceAuth_ParseToken_ExpiredToken(t *testing.T) {
	cfg := AuthConfig{
		SigningKey: "test-key",
	}
	service := NewAuth(cfg)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         uuid.New().String(),
		"email":      "test@example.com",
		"role":       "moderator",
		"expires_at": time.Now().Add(-time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(cfg.SigningKey))
	assert.NoError(t, err)

	_, err = service.ParseToken(tokenString)
	assert.ErrorIs(t, err, ErrTokenExpired)
}

func TestServiceAuth_ParseToken_MissingClaims(t *testing.T) {
	cfg := AuthConfig{
		SigningKey: "test-key",
	}
	service := NewAuth(cfg)

	tests := []struct {
		name        string
		claims      jwt.MapClaims
		expectedErr error
	}{
		{
			name: "Missing ID",
			claims: jwt.MapClaims{
				"email":      "test@example.com",
				"role":       "employee",
				"expires_at": time.Now().Add(TokenTTL).Unix(),
			},
			expectedErr: ErrClaimIdFails,
		},
		{
			name: "Missing Email",
			claims: jwt.MapClaims{
				"id":         uuid.New().String(),
				"role":       "employee",
				"expires_at": time.Now().Add(TokenTTL).Unix(),
			},
			expectedErr: errors.New("email claim is missing or invalid"),
		},
		{
			name: "Missing Role",
			claims: jwt.MapClaims{
				"id":         uuid.New().String(),
				"email":      "test@example.com",
				"expires_at": time.Now().Add(TokenTTL).Unix(),
			},
			expectedErr: errors.New("role claim is missing or invalid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, tt.claims)
			tokenString, err := token.SignedString([]byte(cfg.SigningKey))
			assert.NoError(t, err)

			_, err = service.ParseToken(tokenString)
			assert.ErrorContains(t, err, tt.expectedErr.Error())
		})
	}
}
func TestServiceAuth_ParseToken_InvalidClaimTypes(t *testing.T) {
	cfg := AuthConfig{
		SigningKey: "test-key",
	}
	service := NewAuth(cfg)

	tests := []struct {
		name        string
		claims      jwt.MapClaims
		expectedErr error
	}{
		{
			name: "Invalid ID type",
			claims: jwt.MapClaims{
				"id":         123,
				"email":      "test@example.com",
				"role":       "employee",
				"expires_at": time.Now().Add(TokenTTL).Unix(),
			},
			expectedErr: ErrClaimIdFails,
		},
		{
			name: "Invalid Email type",
			claims: jwt.MapClaims{
				"id":         uuid.New().String(),
				"email":      123,
				"role":       "employee",
				"expires_at": time.Now().Add(TokenTTL).Unix(),
			},
			expectedErr: errors.New("email claim is missing or invalid"),
		},
		{
			name: "Invalid Role type",
			claims: jwt.MapClaims{
				"id":         uuid.New().String(),
				"email":      "test@example.com",
				"role":       123,
				"expires_at": time.Now().Add(TokenTTL).Unix(),
			},
			expectedErr: errors.New("role claim is missing or invalid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, tt.claims)
			tokenString, err := token.SignedString([]byte(cfg.SigningKey))
			assert.NoError(t, err)

			_, err = service.ParseToken(tokenString)
			assert.ErrorContains(t, err, tt.expectedErr.Error())
		})
	}
}
func TestServiceAuth_ParseToken_InvalidKey(t *testing.T) {
	cfg := AuthConfig{
		SigningKey: "test-key",
	}
	service := NewAuth(cfg)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    uuid.New().String(),
		"email": "test@example.com",
		"role":  "employee",
	})

	tokenString, err := token.SignedString([]byte("wrong-key"))
	assert.NoError(t, err)

	_, err = service.ParseToken(tokenString)
	assert.Error(t, err)
}
