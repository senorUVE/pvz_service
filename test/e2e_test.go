package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	config "github.com/senorUVE/pvz_service/configs"
	"github.com/senorUVE/pvz_service/internal/auth"
	"github.com/senorUVE/pvz_service/internal/controller"
	"github.com/senorUVE/pvz_service/internal/dto"
	"github.com/senorUVE/pvz_service/internal/handler"
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

	logrus.SetLevel(logrus.DebugLevel)
	moderToken := getDummyToken(t, server, "moderator")
	empToken := getDummyToken(t, server, "employee")

	pvzID := createPVZ(t, server, moderToken, "Moscow")
	t.Logf("Created PVZ ID: %s", pvzID)

	receptionID := createReception(t, server, empToken, pvzID)
	t.Logf("Created Reception ID: %s", receptionID)

	addProducts(t, server, empToken, pvzID, receptionID, 50)
	t.Log("Products added")

	closeReception(t, server, empToken, pvzID)
	t.Log("Reception closed")

	verifyReceptionState(t, server, empToken, pvzID, receptionID)
	t.Log("Verification complete")
}

func getDummyToken(t *testing.T, server *httptest.Server, role string) string {
	reqBody := dto.DummyLoginRequest{Role: role}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	resp, err := http.Post(server.URL+"/api/dummyLogin", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var token string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&token))

	return token
}

func createPVZ(t *testing.T, server *httptest.Server, token string, city string) uuid.UUID {
	reqBody := dto.PvzCreateRequest{
		City:             city,
		RegistrationDate: time.Now().UTC(),
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", server.URL+"/pvz", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var response dto.PVZResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response))

	return response.Id
}

func createReception(t *testing.T, server *httptest.Server, token string, pvzID uuid.UUID) uuid.UUID {
	reqBody := dto.CreateReceptionRequest{
		PvzId: pvzID,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", server.URL+"/receptions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var response dto.ReceptionResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response))

	return response.Id
}

func addProducts(t *testing.T, server *httptest.Server, token string, pvzID uuid.UUID, receptionID uuid.UUID, count int) {
	for i := 0; i < count; i++ {
		reqBody := dto.AddProductRequest{
			Type:  fmt.Sprintf("Type-%d", i+1),
			PvzId: pvzID,
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", server.URL+"/products", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}
}

func closeReception(t *testing.T, server *httptest.Server, token string, pvzID uuid.UUID) {
	url := fmt.Sprintf("%s/pvz/%s/close_last_reception", server.URL, pvzID.String())
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func verifyReceptionState(t *testing.T, server *httptest.Server, token string, pvzID uuid.UUID, receptionID uuid.UUID) {
	// Получение информации о ПВЗ
	req, _ := http.NewRequest("GET", server.URL+"/pvz?pvzId="+pvzID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var pvzData dto.PVZWithReceptions
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pvzData))

	require.NotEmpty(t, pvzData, "PVZ not found")

	// Поиск нужной приёмки
	var foundReception *dto.ReceptionWithProducts
	for _, reception := range pvzData.Receptions {
		if reception.Reception.Id == receptionID {
			foundReception = &reception
			break
		}
	}

	require.NotNil(t, foundReception, "Reception not found")
	assert.Equal(t, "closed", foundReception.Reception.Status)
	assert.Len(t, foundReception.Products, 50)
}
