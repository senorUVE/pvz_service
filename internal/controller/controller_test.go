package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/senorUVE/pvz_service/internal/dto"
	"github.com/senorUVE/pvz_service/internal/models"
	"github.com/senorUVE/pvz_service/internal/repository"
	"github.com/senorUVE/pvz_service/test/mocks"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestPvzService_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := uuid.MustParse("87c17529-99bb-4815-be06-900c4612902a")
	mockRepo := mocks.NewMockRepository(ctrl)
	mockAuth := mocks.NewMockAuthService(ctrl)

	service := NewPvzService(mockRepo, mockAuth, ServiceConfig{Salt: "test-salt", Cost: bcrypt.DefaultCost})

	ctx := context.Background()
	req := &dto.RegisterRequest{Email: "testemail@mail.ru", Password: "password1", Role: "moderator"}
	expectedUser := &models.User{Id: id, Email: "testemail@mail.ru", Password: "password1", Role: "moderator "}

	mockRepo.EXPECT().CreateUser(ctx, req.Email, gomock.Any(), req.Role).Return(id, nil)

	user, err := service.CreateUser(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.Id, user.Id)
	assert.Equal(t, expectedUser.Email, user.Email)
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password+service.cfg.Salt),
	)
}

func TestPvzService_AuthUser_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := uuid.MustParse("87c17529-99bb-4815-be06-900c4612902a")
	mockRepo := mocks.NewMockRepository(ctrl)
	mockAuth := mocks.NewMockAuthService(ctrl)

	service := NewPvzService(mockRepo, mockAuth, ServiceConfig{Salt: "test-salt"})

	ctx := context.Background()
	req := &dto.AuthRequest{Email: "testemail@mail.ru", Password: "wrong-password"}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password1test-salt"), bcrypt.DefaultCost)
	existingUser := &models.User{Id: id, Email: "testemail@mail.ru", Password: string(hashedPassword)}

	mockRepo.EXPECT().GetUser(ctx, req.Email).Return(existingUser, nil)

	_, err := service.AuthUser(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPasswd, err)
}

func TestPvzService_CreatePVZ_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockAuth := mocks.NewMockAuthService(ctrl)
	service := NewPvzService(mockRepo, mockAuth, ServiceConfig{})

	ctx := context.Background()
	req := &dto.PvzCreateRequest{City: "Москва"}
	expected := &dto.PvzCreateResponse{
		Id:               uuid.New(),
		RegistrationDate: time.Now().UTC(),
		City:             "Москва",
	}

	mockRepo.EXPECT().CreatePvz(ctx, gomock.Any()).Return(expected, nil)

	resp, err := service.CreatePVZ(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, expected.City, resp.City)
	assert.NotEqual(t, uuid.Nil, resp.Id)
}

func TestPvzService_CreateReception_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewPvzService(mockRepo, nil, ServiceConfig{})
	ctx := context.Background()
	pvzId := uuid.New()
	req := &dto.CreateReceptionRequest{PvzId: pvzId}
	expected := &dto.CreateReceptionResponse{
		Id:       uuid.New(),
		DateTime: time.Now().UTC(),
		PvzId:    pvzId,
		Status:   "in_progress",
	}

	mockRepo.EXPECT().CreateReception(ctx, pvzId).Return(expected, nil)

	resp, err := service.CreateReception(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, pvzId, resp.PvzId)
	assert.Equal(t, "in_progress", resp.Status)
}

func TestPvzService_DeleteLastProduct_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewPvzService(mockRepo, nil, ServiceConfig{})
	pvzID := uuid.New()

	ctx := context.Background()
	mockRepo.EXPECT().
		GetActiveReception(ctx, pvzID).
		Return(&models.Reception{Id: uuid.New()}, nil)

	mockRepo.EXPECT().DeleteLastProduct(ctx, pvzID).Return(nil)

	err := service.DeleteLastProduct(ctx, pvzID)

	assert.NoError(t, err)
}

func TestPvzService_GetPvz_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := NewPvzService(nil, nil, ServiceConfig{})
	invalidReq := &dto.GetPvzRequest{
		Page:  -1,
		Limit: 100,
	}

	_, err := service.GetPvz(context.Background(), invalidReq)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestPvzService_CloseReception_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewPvzService(mockRepo, nil, ServiceConfig{})
	ctx := context.Background()
	pvzID := uuid.New()
	expectedErr := errors.New("database error")

	mockRepo.EXPECT().CloseReception(ctx, pvzID).Return(nil, expectedErr)

	_, err := service.CloseReception(ctx, pvzID)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestPvzService_AddProduct_NoActiveReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewPvzService(mockRepo, nil, ServiceConfig{})
	ctx := context.Background()
	pvzId := uuid.New()
	req := &dto.AddProductRequest{PvzId: pvzId, Type: "электроника"}

	mockRepo.EXPECT().GetActiveReception(ctx, pvzId).Return(nil, repository.ErrNoActiveReception)

	_, err := service.AddProduct(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active reception")
}

func TestPvzService_CreateReception_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewPvzService(mockRepo, nil, ServiceConfig{})
	ctx := context.Background()
	pvzID := uuid.New()
	req := &dto.CreateReceptionRequest{PvzId: pvzID}
	expectedErr := errors.New("database error")

	mockRepo.EXPECT().CreateReception(ctx, pvzID).Return(nil, expectedErr)

	_, err := service.CreateReception(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestPvzService_GetUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewPvzService(mockRepo, nil, ServiceConfig{})
	ctx := context.Background()
	email := "test@email.ru"
	expectedUser := &models.User{
		Id:    uuid.New(),
		Email: email,
		Role:  models.RoleModerator,
	}

	mockRepo.EXPECT().
		GetUser(ctx, email).
		Return(expectedUser, nil)

	user, err := service.GetUser(ctx, email)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestPvzService_GetUser_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewPvzService(mockRepo, nil, ServiceConfig{})
	ctx := context.Background()
	email := "notfound@mail.ru"

	mockRepo.EXPECT().
		GetUser(ctx, email).
		Return(nil, repository.ErrUserNotFound)

	_, err := service.GetUser(ctx, email)

	assert.ErrorIs(t, err, repository.ErrUserNotFound)
}

func TestPvzService_GetPvz_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewPvzService(mockRepo, nil, ServiceConfig{})
	ctx := context.Background()
	req := &dto.GetPvzRequest{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
		Page:      1,
		Limit:     10,
	}
	expected := []*dto.PVZWithReceptions{
		{
			PVZ: dto.PVZResponse{
				Id:               uuid.New(),
				RegistrationDate: time.Now(),
				City:             "Москва",
			},
		},
	}

	mockRepo.EXPECT().GetPvz(ctx, req.StartDate, req.EndDate, req.Page, req.Limit).Return(expected, nil)

	result, err := service.GetPvz(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestPvzService_CloseReception_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewPvzService(mockRepo, nil, ServiceConfig{})
	ctx := context.Background()
	pvzID := uuid.New()
	expected := &dto.CloseLastReceptionResponse{
		Status: "closed",
	}

	mockRepo.EXPECT().
		CloseReception(ctx, pvzID).
		Return(expected, nil)

	resp, err := service.CloseReception(ctx, pvzID)

	assert.NoError(t, err)
	assert.Equal(t, "closed", resp.Status)
}

func TestPvzService_AddProduct_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewPvzService(mockRepo, nil, ServiceConfig{})
	ctx := context.Background()
	pvzID := uuid.New()
	req := &dto.AddProductRequest{
		PvzId: pvzID,
		Type:  "электроника",
	}
	reception := &models.Reception{Id: uuid.New()}
	expected := &dto.AddProductResponse{
		Type: "электроника",
	}

	mockRepo.EXPECT().
		GetActiveReception(ctx, pvzID).
		Return(reception, nil)

	mockRepo.EXPECT().
		CreateProduct(ctx, req.Type, reception.Id).
		Return(expected, nil)

	resp, err := service.AddProduct(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, "электроника", resp.Type)
}

func TestValidateRegisterRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *dto.RegisterRequest
		wantErr bool
	}{
		{
			name:    "valid",
			req:     &dto.RegisterRequest{Email: "test@email.ru", Password: "pass#1234", Role: "moderator"},
			wantErr: false,
		},
		{
			name:    "invalid email",
			req:     &dto.RegisterRequest{Email: "bad-email", Password: "pass", Role: "moderator"},
			wantErr: true,
		},
		{
			name:    "short password",
			req:     &dto.RegisterRequest{Email: "test@email.ru", Password: "123", Role: "moderator"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegisterRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPvzService_AuthUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockAuth := mocks.NewMockAuthService(ctrl)
	service := NewPvzService(mockRepo, mockAuth, ServiceConfig{Salt: "test-salt"})

	ctx := context.Background()
	req := &dto.AuthRequest{Email: "test@email.ru", Password: "valid-password"}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("valid-passwordtest-salt"), bcrypt.DefaultCost)
	user := &models.User{
		Email:    "test@email.ru",
		Password: string(hashedPassword),
		Role:     models.RoleModerator,
	}

	mockRepo.EXPECT().GetUser(ctx, req.Email).Return(user, nil)
	mockAuth.EXPECT().GenerateToken(user).Return("valid-token", nil)

	resp, err := service.AuthUser(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, "valid-token", resp.Token)
}
func TestPvzService_AuthUser_TokenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockAuth := mocks.NewMockAuthService(ctrl)
	service := NewPvzService(mockRepo, mockAuth, ServiceConfig{Salt: "test-salt"})

	ctx := context.Background()
	req := &dto.AuthRequest{Email: "test@email.ru", Password: "valid-password"}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("valid-passwordtest-salt"), bcrypt.DefaultCost)
	user := &models.User{
		Email:    "test@email.ru",
		Password: string(hashedPassword),
	}

	mockRepo.EXPECT().GetUser(ctx, req.Email).Return(user, nil)
	mockAuth.EXPECT().GenerateToken(user).Return("", errors.New("token error"))

	_, err := service.AuthUser(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token error")
}

func TestPvzService_CreatePVZ_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewPvzService(mockRepo, nil, ServiceConfig{})
	ctx := context.Background()
	req := &dto.PvzCreateRequest{City: "Москва"}
	expectedErr := errors.New("database error")

	mockRepo.EXPECT().
		CreatePvz(ctx, gomock.Any()).
		Return(nil, expectedErr)

	_, err := service.CreatePVZ(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestValidatePVZ(t *testing.T) {
	tests := []struct {
		name    string
		pvz     *dto.PVZResponse
		wantErr bool
	}{
		{
			name:    "valid",
			pvz:     &dto.PVZResponse{City: "Москва"},
			wantErr: false,
		},
		{
			name:    "invalid city",
			pvz:     &dto.PVZResponse{City: "Новосибирск"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePVZ(tt.pvz)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPvzService_DeleteLastProduct_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewPvzService(mockRepo, nil, ServiceConfig{})
	pvzID := uuid.New()
	expectedErr := errors.New("delete error")

	ctx := context.Background()
	mockRepo.EXPECT().
		GetActiveReception(ctx, pvzID).
		Return(&models.Reception{Id: uuid.New()}, nil)
	mockRepo.EXPECT().
		DeleteLastProduct(ctx, pvzID).
		Return(expectedErr)

	err := service.DeleteLastProduct(ctx, pvzID)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestValidateDeleteProductRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *dto.DeleteProductRequest
		wantErr bool
	}{
		{
			name:    "nil uuid",
			req:     &dto.DeleteProductRequest{PvzId: uuid.Nil},
			wantErr: true,
		},
		{
			name:    "valid uuid",
			req:     &dto.DeleteProductRequest{PvzId: uuid.New()},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeleteProductRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAuth(t *testing.T) {
	tests := []struct {
		name    string
		req     *dto.AuthRequest
		wantErr error
	}{
		{
			name:    "valid",
			req:     &dto.AuthRequest{Email: "test@mail.ru", Password: "password123"},
			wantErr: nil,
		},
		{
			name:    "invalid email",
			req:     &dto.AuthRequest{Email: "invalid-email", Password: "password123"},
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "short password",
			req:     &dto.AuthRequest{Email: "test@mail.ru", Password: "short"},
			wantErr: ErrShortPassword,
		},
		{
			name:    "empty password",
			req:     &dto.AuthRequest{Email: "test@mail.ru", Password: ""},
			wantErr: ErrShortPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAuth(tt.req)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestValidateCloseLastReceptionRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *dto.CloseLastReceptionRequest
		wantErr error
	}{
		{
			name:    "valid",
			req:     &dto.CloseLastReceptionRequest{PvzId: uuid.New()},
			wantErr: nil,
		},
		{
			name:    "nil uuid",
			req:     &dto.CloseLastReceptionRequest{PvzId: uuid.Nil},
			wantErr: ErrInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCloseLastReceptionRequest(tt.req)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestValidatePvzCreateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *dto.PvzCreateRequest
		wantErr error
	}{
		{
			name:    "valid moscow",
			req:     &dto.PvzCreateRequest{City: "Москва"},
			wantErr: nil,
		},
		{
			name:    "invalid city",
			req:     &dto.PvzCreateRequest{City: "Новосибирск"},
			wantErr: ErrInvalidCity,
		},
		{
			name:    "future date",
			req:     &dto.PvzCreateRequest{City: "Москва", RegistrationDate: time.Now().Add(24 * time.Hour)},
			wantErr: ErrFutureDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePvzCreateRequest(tt.req)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestValidateAddProductRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *dto.AddProductRequest
		wantErr error
	}{
		{
			name:    "valid electronics",
			req:     &dto.AddProductRequest{PvzId: uuid.New(), Type: "электроника"},
			wantErr: nil,
		},
		{
			name:    "invalid type",
			req:     &dto.AddProductRequest{PvzId: uuid.New(), Type: "неизвестный"},
			wantErr: ErrInvalidProductType,
		},
		{
			name:    "nil uuid",
			req:     &dto.AddProductRequest{PvzId: uuid.Nil, Type: "электроника"},
			wantErr: ErrInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddProductRequest(tt.req)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestPvzService_DummyLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockAuth := mocks.NewMockAuthService(ctrl)
	service := &PvzService{repo: mockRepo, auth: mockAuth}

	t.Run("Success", func(t *testing.T) {
		expectedUser := &models.User{Role: models.RoleModerator}
		expectedToken := "generated.token"

		mockRepo.EXPECT().DummyLogin(gomock.Any(), "moderator").Return(expectedUser, nil)
		mockAuth.EXPECT().GenerateToken(expectedUser).Return(expectedToken, nil)

		token, err := service.DummyLogin(context.Background(), "moderator")

		assert.NoError(t, err)
		assert.Equal(t, expectedToken, token)
	})

	t.Run("Repository error", func(t *testing.T) {
		expectedErr := errors.New("repo error")
		mockRepo.EXPECT().DummyLogin(gomock.Any(), "moderator").Return(nil, expectedErr)

		_, err := service.DummyLogin(context.Background(), "moderator")
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("Auth error", func(t *testing.T) {
		expectedErr := errors.New("auth error")
		mockRepo.EXPECT().DummyLogin(gomock.Any(), "moderator").Return(&models.User{}, nil)
		mockAuth.EXPECT().GenerateToken(gomock.Any()).Return("", expectedErr)

		_, err := service.DummyLogin(context.Background(), "moderator")
		assert.ErrorIs(t, err, expectedErr)
	})
}
