package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	config "github.com/senorUVE/pvz_service/configs"
	"github.com/senorUVE/pvz_service/internal/auth"
	"github.com/senorUVE/pvz_service/internal/controller"
	"github.com/senorUVE/pvz_service/internal/dto"
	"github.com/senorUVE/pvz_service/internal/handler"
	"github.com/senorUVE/pvz_service/internal/metrics"
	"github.com/senorUVE/pvz_service/internal/repository"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
)

func setupTestServer() (*httptest.Server, func()) {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		logrus.Fatalf("Failed to load Config: %v", err)
	}

	db, err := repository.NewRepository(cfg.DBConfig)
	if err != nil {
		logrus.Fatalf("Failed to init db: %v", err)
	}

	logrus.Info("Database initialized successfully")

	authService := auth.NewAuth(cfg.AuthConfig)

	logrus.Info("Auth service initialized successfully")

	pvzService := controller.NewPvzService(db, authService, cfg.ServiceConfig)

	logrus.Info("Pvz service initialized successfully")

	pvzHandler := handler.NewPvzHandler(pvzService, authService, cfg.AppPort)

	logrus.Info("Pvz handler initialized successfully")

	handler.RegisterRoutes(pvzHandler)

	logrus.Info("Routes registered successfully")

	pvzHandler.GetEcho().GET("/metrics", func(c echo.Context) error {
		metrics.PrometheusHandler().ServeHTTP(c.Response(), c.Request())
		return nil
	})

	server := httptest.NewServer(pvzHandler.GetEcho())

	logrus.Info("Test server started successfully")

	cleanupFunc := func() {
		conn, err := sqlx.Open(
			"postgres", fmt.Sprintf(
				"%s://%s:%s@%s:%s/%s?sslmode=%s",
				cfg.DBConfig.DBDriver,
				cfg.DBConfig.DBUser,
				cfg.DBConfig.DBPass,
				cfg.DBConfig.DBHost,
				cfg.DBConfig.DBPort,
				cfg.DBConfig.DBName,
				cfg.DBConfig.DBSSL,
			),
		)

		if err != nil {
			logrus.Fatalf("Failed to connect to database: %v", err)
		}
		defer conn.Close()

		_, err = conn.Exec("TRUNCATE TABLE users, pvz, product, reception RESTART IDENTITY CASCADE;")
		if err != nil {
			logrus.Fatalf("Failed to truncate tables: %v", err)
		}

		logrus.Info("Database cleanup completed")
	}

	return server, cleanupFunc
}

func TestPVZLifecycle(t *testing.T) {
	server, cleanup := setupTestServer()
	defer server.Close()
	defer cleanup()

	moderatorToken := getDummyToken(t, server.URL, "moderator")

	pvzResp := createPVZ(t, server.URL, moderatorToken, "Москва")

	employeeToken := getDummyToken(t, server.URL, "employee")

	receptionResp := createReception(t, server.URL, employeeToken, pvzResp.Id)
	require.Equal(t, "in_progress", receptionResp.Status, "Should start in progress")

	for i := 0; i < 50; i++ {
		addProduct(t, server.URL, employeeToken, pvzResp.Id)
	}

	closeLastReception(t, server.URL, employeeToken, pvzResp.Id)
}

func getDummyToken(t *testing.T, baseURL, role string) string {
	reqBody := dto.DummyLoginRequest{
		Role: role,
	}
	b, err := json.Marshal(reqBody)
	require.NoError(t, err)

	resp, err := http.Post(baseURL+"/api/dummyLogin", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "dummyLogin should return 200")

	var token string
	err = json.NewDecoder(resp.Body).Decode(&token)
	require.NoError(t, err)

	require.NotEmpty(t, token, "token should not be empty")
	return token
}

func getProducts(t *testing.T, server *httptest.Server, token string, receptionID uuid.UUID) []dto.ProductResponse {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/receptions/%s/products", server.URL, receptionID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var products []dto.ProductResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&products))
	return products
}

func getReception(t *testing.T, server *httptest.Server, token string, receptionID uuid.UUID) dto.ReceptionResponse {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/receptions/%s", server.URL, receptionID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var reception dto.ReceptionResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&reception))
	return reception
}

func createPVZ(t *testing.T, baseURL, token, city string) dto.PvzCreateResponse {
	reqBody := dto.PvzCreateRequest{
		City: city,
	}
	b, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, baseURL+"/pvz", bytes.NewReader(b))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode, "CreatePVZ should return 201")

	var pvzResp dto.PvzCreateResponse
	err = json.NewDecoder(resp.Body).Decode(&pvzResp)
	require.NoError(t, err)

	require.NotEqual(t, uuid.Nil, pvzResp.Id, "pvzId must be generated")
	return pvzResp
}

func createReception(t *testing.T, baseURL, token string, pvzId uuid.UUID) dto.CreateReceptionResponse {
	reqBody := dto.CreateReceptionRequest{
		PvzId: pvzId,
	}
	b, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, baseURL+"/receptions", bytes.NewReader(b))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode, "CreateReception should return 201")

	var recResp dto.CreateReceptionResponse
	err = json.NewDecoder(resp.Body).Decode(&recResp)
	require.NoError(t, err)

	require.NotEqual(t, uuid.Nil, recResp.Id, "receptionId must be generated")
	return recResp
}

func addProduct(t *testing.T, baseURL, token string, pvzId uuid.UUID) {
	reqBody := dto.AddProductRequest{
		Type:  "электроника", // или любой другой тип
		PvzId: pvzId,
	}
	b, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, baseURL+"/products", bytes.NewReader(b))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode, "AddProduct should return 201")
}

func closeLastReception(t *testing.T, baseURL, token string, pvzId uuid.UUID) {
	url := fmt.Sprintf("%s/pvz/%s/close_last_reception", baseURL, pvzId.String())
	req, err := http.NewRequest(http.MethodPost, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "CloseReception should return 200")
}
