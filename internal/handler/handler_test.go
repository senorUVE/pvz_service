package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/senorUVE/pvz_service/internal/dto"
	"github.com/senorUVE/pvz_service/internal/models"
	"github.com/senorUVE/pvz_service/test/mocks"
	"github.com/stretchr/testify/assert"
)

func TestShopHandlerPing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShopService := mocks.NewMockPvzService(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)

	e := echo.New()
	handler := NewPvzHandler(mockShopService, mockAuthService, "8080")

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Ping(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "pong", rec.Body.String())
}

func TestShopHandlerAuthMiddleware_InvalidAuthHeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShopService := mocks.NewMockPvzService(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)

	e := echo.New()
	handler := NewPvzHandler(mockShopService, mockAuthService, "8080")

	e.Use(handler.AuthMiddleware())

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(authorizationHeader, "InvalidHeader")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), ErrInvalidAuthHeader.Error())
}

func TestShopHandlerAuthMiddleware_EmptyToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShopService := mocks.NewMockPvzService(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)

	e := echo.New()
	handler := NewPvzHandler(mockShopService, mockAuthService, "8080")

	e.Use(handler.AuthMiddleware())

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), ErrEmptyToken.Error())
}

func TestShopHandlerAuthMiddleware_ValidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShopService := mocks.NewMockPvzService(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)

	e := echo.New()
	handler := NewPvzHandler(mockShopService, mockAuthService, "8080")

	e.Use(handler.AuthMiddleware())

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	mockAuthService.EXPECT().ParseToken("valid-token").Return(&models.User{Role: models.RoleEmployee}, nil)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(authorizationHeader, "Bearer valid-token")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}

func TestRegisterHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPvzService(ctrl)
	mockAuth := mocks.NewMockAuthService(ctrl)
	handler := NewPvzHandler(mockService, mockAuth, "8080")

	reqBody := `{"email":"test@email.ru","password":"password123","role":"moderator"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	expectedUser := &models.User{Id: uuid.New(), Email: "test@email.ru"}
	mockService.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(expectedUser, nil)

	err := handler.Register(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), `"email":"test@email.ru"`)
}

func TestRegisterHandler_InvalidRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPvzService(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	handler := NewPvzHandler(mockService, authService, "8080")

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader("invalid json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	err := handler.Register(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid request")
}

func TestCreatePVZHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPvzService(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	handler := NewPvzHandler(mockService, authService, "8080")

	reqBody := `{"city":"Москва"}`
	req := httptest.NewRequest(http.MethodPost, "/pvz", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	expected := &dto.PvzCreateResponse{City: "Москва"}
	mockService.EXPECT().CreatePVZ(gomock.Any(), gomock.Any()).Return(expected, nil)

	err := handler.CreatePVZ(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), `"city":"Москва"`)
}

func TestGetPvzHandler_WithFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPvzService(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	handler := NewPvzHandler(mockService, authService, "8080")

	req := httptest.NewRequest(http.MethodGet, "/pvz?startDate=2023-01-01T00:00:00Z&endDate=2023-01-31T23:59:59Z&page=2&limit=20", nil)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	expected := []*dto.PVZWithReceptions{{PVZ: dto.PVZResponse{City: "Москва"}}}
	mockService.EXPECT().GetPvz(gomock.Any(), gomock.Any()).Return(expected, nil)

	err := handler.GetPvz(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"city":"Москва"`)
}

func TestCloseReceptionHandler_InvalidUUID(t *testing.T) {
	handler := NewPvzHandler(nil, nil, "8080")

	req := httptest.NewRequest(http.MethodPost, "/pvz/invalid_uuid/close_last_reception", nil)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)
	c.SetParamNames("pvzId")
	c.SetParamValues("invalid_uuid")

	err := handler.CloseReception(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid PVZ ID")
}

func TestDeleteLastProductHandler_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPvzService(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	handler := NewPvzHandler(mockService, authService, "8080")
	pvzID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/pvz/"+pvzID.String()+"/delete_last_product", nil)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)
	c.SetParamNames("pvzId")
	c.SetParamValues(pvzID.String())

	expectedErr := errors.New("database error")
	mockService.EXPECT().DeleteLastProduct(gomock.Any(), pvzID).Return(expectedErr)

	err := handler.DeleteLastProduct(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), expectedErr.Error())
}

func TestRoleMiddleware_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuthService(ctrl)
	mockService := mocks.NewMockPvzService(ctrl)
	handler := NewPvzHandler(mockService, mockAuth, "8080")

	e := echo.New()

	e.Use(handler.AuthMiddleware())
	e.Use(handler.RoleMiddleware(models.RoleEmployee))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer valid_token")
	rec := httptest.NewRecorder()

	mockAuth.EXPECT().ParseToken("valid_token").Return(&models.User{Role: models.RoleModerator}, nil)

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "Insufficient permissions")
}

func TestCreateReceptionHandler_InvalidRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := NewPvzHandler(nil, nil, "8080")

	// Неверный формат тела запроса
	req := httptest.NewRequest(http.MethodPost, "/receptions", strings.NewReader("{invalid json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	err := handler.CreateReception(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid request")
}

func TestAddProductHandler_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authService := mocks.NewMockAuthService(ctrl)
	mockService := mocks.NewMockPvzService(ctrl)
	handler := NewPvzHandler(mockService, authService, "8080")

	reqBody := `{"pvzId":"aa6c3be3-945d-43b3-b0a5-18a1d8d5a5b3","type":"электроника"}`
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	expectedErr := errors.New("service error")
	mockService.EXPECT().AddProduct(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

	err := handler.AddProduct(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), expectedErr.Error())
}

func TestRoleMiddleware_InvalidUserContext(t *testing.T) {
	handler := NewPvzHandler(nil, nil, "8080")

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Неправильный тип в контексте
			c.Set("user", "invalid-type")
			return next(c)
		}
	})
	e.Use(handler.RoleMiddleware(models.RoleEmployee))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "User not found in context")
}

func TestCloseReceptionHandler_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authService := mocks.NewMockAuthService(ctrl)
	mockService := mocks.NewMockPvzService(ctrl)
	handler := NewPvzHandler(mockService, authService, "8080")
	pvzID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/pvz/"+pvzID.String()+"/close_last_reception", nil)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)
	c.SetParamNames("pvzId")
	c.SetParamValues(pvzID.String())

	expectedErr := errors.New("database error")
	mockService.EXPECT().CloseReception(gomock.Any(), pvzID).Return(nil, expectedErr)

	err := handler.CloseReception(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), expectedErr.Error())
}

func TestGetPvzHandler_InvalidDates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPvzService(ctrl)
	authService := mocks.NewMockAuthService(ctrl)
	handler := NewPvzHandler(mockService, authService, "8080")

	tests := []struct {
		name        string
		url         string
		expectError string
	}{
		{
			name:        "InvalidStartDate",
			url:         "/pvz?startDate=invalid",
			expectError: "Invalid request",
		},
		{
			name:        "InvalidEndDate",
			url:         "/pvz?endDate=invalid",
			expectError: "Invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			c := handler.e.NewContext(req, rec)

			err := handler.GetPvz(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectError)
		})
	}
}

func TestCreatePVZHandler_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authService := mocks.NewMockAuthService(ctrl)
	mockService := mocks.NewMockPvzService(ctrl)
	handler := NewPvzHandler(mockService, authService, "8080")

	reqBody := `{"city":"Москва"}`
	req := httptest.NewRequest(http.MethodPost, "/pvz", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	expectedErr := errors.New("service error")
	mockService.EXPECT().CreatePVZ(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

	err := handler.CreatePVZ(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), expectedErr.Error())
}

func TestLoginHandler_InvalidRequest(t *testing.T) {
	handler := NewPvzHandler(nil, nil, "8080")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("invalid json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	err := handler.Login(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid request")
}

func TestLoginHandler_AuthError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authService := mocks.NewMockAuthService(ctrl)
	mockService := mocks.NewMockPvzService(ctrl)
	handler := NewPvzHandler(mockService, authService, "8080")

	reqBody := `{"email":"test@test.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	expectedErr := errors.New("invalid credentials")
	mockService.EXPECT().AuthUser(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

	err := handler.Login(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), expectedErr.Error())
}

func TestCloseReceptionHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authService := mocks.NewMockAuthService(ctrl)
	mockService := mocks.NewMockPvzService(ctrl)
	handler := NewPvzHandler(mockService, authService, "8080")
	pvzID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/pvz/"+pvzID.String()+"/close_last_reception", nil)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)
	c.SetParamNames("pvzId")
	c.SetParamValues(pvzID.String())

	expected := &dto.CloseLastReceptionResponse{Status: "closed"}
	mockService.EXPECT().CloseReception(gomock.Any(), pvzID).Return(expected, nil)

	err := handler.CloseReception(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"closed"`)
}

func TestRoleMiddleware_Allowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPvzService(ctrl)
	mockAuth := mocks.NewMockAuthService(ctrl)
	handler := NewPvzHandler(mockService, mockAuth, "8080")

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user", &models.User{Role: models.RoleEmployee})
			return next(c)
		}
	})
	e.Use(handler.RoleMiddleware(models.RoleEmployee))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRegisterHandler_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authService := mocks.NewMockAuthService(ctrl)
	mockService := mocks.NewMockPvzService(ctrl)
	handler := NewPvzHandler(mockService, authService, "8080")

	reqBody := `{"email":"test@test.com","password":"password123","role":"moderator"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	expectedErr := errors.New("email already exists")
	mockService.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

	err := handler.Register(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), expectedErr.Error())
}

func TestAuthMiddleware_ParseTokenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceMock := mocks.NewMockPvzService(ctrl)
	mockAuth := mocks.NewMockAuthService(ctrl)
	handler := NewPvzHandler(serviceMock, mockAuth, "8080")

	e := echo.New()
	e.Use(handler.AuthMiddleware())
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	mockAuth.EXPECT().ParseToken("invalid-token").Return(nil, errors.New("token expired"))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(authorizationHeader, "Bearer invalid-token")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "internal error")
}

func TestRoleMiddleware_MissingUser(t *testing.T) {
	handler := NewPvzHandler(nil, nil, "8080")

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			return next(c)
		}
	})
	e.Use(handler.RoleMiddleware(models.RoleEmployee))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetPvzHandler_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPvzService(ctrl)
	handler := NewPvzHandler(mockService, nil, "8080")

	req := httptest.NewRequest(http.MethodGet, "/pvz", nil)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	expectedErr := errors.New("database error")
	mockService.EXPECT().GetPvz(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

	err := handler.GetPvz(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), expectedErr.Error())
}

func TestCreatePVZHandler_InvalidRequest(t *testing.T) {
	handler := NewPvzHandler(nil, nil, "8080")

	req := httptest.NewRequest(http.MethodPost, "/pvz", strings.NewReader(`{"invalid": "data"`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	err := handler.CreatePVZ(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDeleteLastProduct_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPvzService(ctrl)
	handler := NewPvzHandler(mockService, nil, "8080")
	pvzID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/pvz/"+pvzID.String()+"/delete_last_product", nil)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)
	c.SetParamNames("pvzId")
	c.SetParamValues(pvzID.String())

	mockService.EXPECT().DeleteLastProduct(gomock.Any(), pvzID).Return(nil)

	err := handler.DeleteLastProduct(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCreateReception_InvalidUUID(t *testing.T) {
	handler := NewPvzHandler(nil, nil, "8080")

	reqBody := `{"pvzId": "invalid-uuid"}`
	req := httptest.NewRequest(http.MethodPost, "/receptions", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	err := handler.CreateReception(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetPvzHandler_MaxLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPvzService(ctrl)
	handler := NewPvzHandler(mockService, nil, "8080")

	req := httptest.NewRequest(http.MethodGet, "/pvz?limit=100", nil)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	mockService.EXPECT().GetPvz(gomock.Any(), gomock.Any()).Return([]*dto.PVZWithReceptions{}, nil)

	err := handler.GetPvz(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRoleMiddleware_AllRoles(t *testing.T) {
	roles := []models.Role{models.RoleEmployee, models.RoleModerator}

	for _, role := range roles {
		t.Run(role.String(), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := NewPvzHandler(nil, nil, "8080")
			e := echo.New()

			e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c echo.Context) error {
					c.Set("user", &models.User{Role: role})
					return next(c)
				}
			})

			e.Use(handler.RoleMiddleware(role))
			e.GET("/test", func(c echo.Context) error {
				return c.String(http.StatusOK, "success")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestGetPvzHandler_ValidRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPvzService(ctrl)
	handler := NewPvzHandler(mockService, nil, "8080")

	expected := []*dto.PVZWithReceptions{{PVZ: dto.PVZResponse{City: "Москва"}}}
	mockService.EXPECT().GetPvz(gomock.Any(), gomock.Any()).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/pvz?page=1&limit=10", nil)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	err := handler.GetPvz(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Москва")
}

func TestCreateReceptionHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPvzService(ctrl)
	handler := NewPvzHandler(mockService, nil, "8080")

	pvzID := uuid.New()
	currentTime := time.Now().UTC()

	tests := []struct {
		name         string
		requestBody  string
		mockResponse *dto.CreateReceptionResponse
		mockError    error
		expectCode   int
		expectBody   string
	}{
		{
			name:        "Success",
			requestBody: `{"pvzId":"` + pvzID.String() + `"}`,
			mockResponse: &dto.CreateReceptionResponse{
				Id:       uuid.New(),
				DateTime: currentTime,
				PvzId:    pvzID,
				Status:   "in_progress",
			},
			expectCode: http.StatusCreated,
			expectBody: `"status":"in_progress"`,
		},
		{
			name:        "Invalid UUID format",
			requestBody: `{"pvzId":"invalid-uuid"}`,
			expectCode:  http.StatusBadRequest,
			expectBody:  "Invalid request",
		},
		{
			name:        "Service error",
			requestBody: `{"pvzId":"` + pvzID.String() + `"}`,
			mockError:   errors.New("database connection failed"),
			expectCode:  http.StatusBadRequest,
			expectBody:  "database connection failed",
		},
		{
			name:        "Invalid request body",
			requestBody: `{"invalid": "data"`,
			expectCode:  http.StatusBadRequest,
			expectBody:  "Invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/receptions", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := handler.e.NewContext(req, rec)

			if tt.mockResponse != nil || tt.mockError != nil {
				mockService.EXPECT().CreateReception(gomock.Any(), gomock.Any()).Return(tt.mockResponse, tt.mockError)
			}

			err := handler.CreateReception(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectCode, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectBody)
		})
	}
}

func TestLoginHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPvzService(ctrl)
	handler := NewPvzHandler(mockService, nil, "8080")

	reqBody := `{"email":"test@test.com","password":"correct"}`
	expectedToken := "valid.token.123"
	mockResponse := &dto.AuthResponse{Token: expectedToken}

	mockService.EXPECT().AuthUser(gomock.Any(), gomock.Any()).Return(mockResponse, nil)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	err := handler.Login(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), expectedToken)
}

func TestAddProduct_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPvzService(ctrl)
	handler := NewPvzHandler(mockService, nil, "8080")

	pvzID := uuid.New()
	reqBody := `{"pvzId":"` + pvzID.String() + `","type":"electronics"}`
	mockResponse := &dto.AddProductResponse{Id: uuid.New()}

	mockService.EXPECT().AddProduct(gomock.Any(), gomock.Any()).Return(mockResponse, nil)

	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	err := handler.AddProduct(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), mockResponse.Id.String())
}

func TestAddProduct_InvalidJSON(t *testing.T) {
	handler := NewPvzHandler(nil, nil, "8080")

	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader("{invalid json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := handler.e.NewContext(req, rec)

	err := handler.AddProduct(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid request")
}

func TestPvzHandler_DummyLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPvzService(ctrl)
	handler := NewPvzHandler(mockService, nil, "8080")

	t.Run("Success", func(t *testing.T) {
		expectedToken := "test.token"
		reqBody := `{"role":"moderator"}`

		mockService.EXPECT().DummyLogin(gomock.Any(), "moderator").Return(expectedToken, nil)

		req := httptest.NewRequest(http.MethodPost, "/dummyLogin", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := handler.e.NewContext(req, rec)

		err := handler.DummyLogin(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), expectedToken)
	})

	t.Run("Invalid request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/dummyLogin", strings.NewReader("{invalid"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := handler.e.NewContext(req, rec)

		err := handler.DummyLogin(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid request")
	})

	t.Run("Service error", func(t *testing.T) {
		reqBody := `{"role":"moderator"}`
		expectedErr := errors.New("service error")

		mockService.EXPECT().DummyLogin(gomock.Any(), "moderator").Return("", expectedErr)

		req := httptest.NewRequest(http.MethodPost, "/dummyLogin", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := handler.e.NewContext(req, rec)

		err := handler.DummyLogin(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), expectedErr.Error())
	})
}
