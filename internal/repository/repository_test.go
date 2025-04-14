package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/senorUVE/pvz_service/internal/dto"
	"github.com/senorUVE/pvz_service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_GetUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &Repository{db: sqlx.NewDb(db, "sqlmock")}

	tests := []struct {
		name         string
		email        string
		mockExpect   func()
		expectedResp func(*testing.T, *models.User, error)
	}{
		{
			name:  "success GetUser",
			email: "testemail@mail.ru",
			mockExpect: func() {
				rows := sqlmock.NewRows([]string{"id", "email", "password_salt"}).
					AddRow("87c17529-99bb-4815-be06-900c4612902a", "testemail@mail.ru", "hashedpassword")
				mock.ExpectQuery(regexp.QuoteMeta(getUserByEmail)).
					WithArgs("testemail@mail.ru").
					WillReturnRows(rows)
			},
			expectedResp: func(t *testing.T, user *models.User, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "testemail@mail.ru", user.Email)
				assert.Equal(t, "hashedpassword", user.Password)
			},
		},
		{
			name:  "user not found",
			email: "nonexistent",
			mockExpect: func() {
				mock.ExpectQuery(regexp.QuoteMeta(getUserByEmail)).
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			expectedResp: func(t *testing.T, user *models.User, err error) {
				assert.Error(t, err)
				assert.Equal(t, ErrUserNotFound, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()
			user, err := repo.GetUser(context.Background(), tt.email)
			tt.expectedResp(t, user, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepository_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &Repository{db: sqlx.NewDb(db, "sqlmock")}
	expectedId := uuid.MustParse("87c17529-99bb-4815-be06-900c4612902a")

	tests := []struct {
		name         string
		email        string
		password     string
		role         string
		mockExpect   func()
		expectedResp func(*testing.T, uuid.UUID, error)
	}{
		{
			name:     "success CreateUser",
			email:    "testemail@mail.ru",
			password: "testpassword",
			role:     "moderator",
			mockExpect: func() {
				mock.ExpectQuery(regexp.QuoteMeta(createUser)).
					WithArgs(sqlmock.AnyArg(), "testemail@mail.ru", "testpassword", "moderator").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedId))
			},
			expectedResp: func(t *testing.T, id uuid.UUID, err error) {
				assert.NoError(t, err)
				assert.Equal(t, expectedId, id)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()
			id, err := repo.CreateUser(context.Background(), tt.email, tt.password, tt.role)
			tt.expectedResp(t, id, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepository_CreatePvz(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &Repository{db: sqlx.NewDb(db, "sqlmock")}
	expectedId := uuid.MustParse("87c17529-99bb-4815-be06-900c4612902a")
	testTime := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name         string
		pvz          models.PVZ
		mockExpect   func()
		expectedResp func(*testing.T, *dto.PvzCreateResponse, error)
	}{
		{
			name: "success CreatePvz",
			pvz: models.PVZ{
				RegistrationDate: testTime,
				City:             models.City("Москва"),
			},
			mockExpect: func() {
				mock.ExpectQuery(regexp.QuoteMeta(createPVZ)).
					WithArgs(sqlmock.AnyArg(), testTime, "Москва").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedId))
			},
			expectedResp: func(t *testing.T, resp *dto.PvzCreateResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, expectedId, resp.Id)
				assert.Equal(t, testTime, resp.RegistrationDate)
				assert.Equal(t, "Москва", resp.City)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()
			resp, err := repo.CreatePvz(context.Background(), tt.pvz)
			tt.expectedResp(t, resp, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepository_CreateReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &Repository{db: sqlx.NewDb(db, "sqlmock")}
	pvzId := uuid.MustParse("87c17529-99bb-4815-be06-900c4612902a")
	testTime := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name         string
		pvzId        uuid.UUID
		mockExpect   func()
		expectedResp func(*testing.T, *dto.CreateReceptionResponse, error)
	}{
		{
			name:  "success CreateReception",
			pvzId: pvzId,
			mockExpect: func() {
				mock.ExpectQuery(regexp.QuoteMeta(createReception)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), pvzId).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "date_time", "status"}).
							AddRow(pvzId, testTime, "in_progress"),
					)
			},
			expectedResp: func(t *testing.T, resp *dto.CreateReceptionResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, pvzId, resp.PvzId)
				assert.Equal(t, testTime, resp.DateTime)
				assert.Equal(t, "in_progress", resp.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()
			resp, err := repo.CreateReception(context.Background(), tt.pvzId)
			tt.expectedResp(t, resp, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepository_CreateProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &Repository{db: sqlx.NewDb(db, "sqlmock")}
	receptionId := uuid.MustParse("87c17529-99bb-4815-be06-900c4612902a")
	testTime := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name         string
		productType  string
		receptionId  uuid.UUID
		mockExpect   func()
		expectedResp func(*testing.T, *dto.AddProductResponse, error)
	}{
		{
			name:        "success CreateProduct",
			productType: "электроника",
			receptionId: receptionId,
			mockExpect: func() {
				mock.ExpectQuery(regexp.QuoteMeta(createProduct)).
					WithArgs(sqlmock.AnyArg(), testTime, "электроника", receptionId).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(receptionId))
			},
			expectedResp: func(t *testing.T, resp *dto.AddProductResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "электроника", resp.Type)
				assert.Equal(t, receptionId, resp.ReceptionId)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()
			resp, err := repo.CreateProduct(context.Background(), tt.productType, tt.receptionId)
			tt.expectedResp(t, resp, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepository_CloseReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &Repository{db: sqlx.NewDb(db, "sqlmock")}
	pvzId := uuid.MustParse("87c17529-99bb-4815-be06-900c4612902a")
	testTime := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name         string
		pvzId        uuid.UUID
		mockExpect   func()
		expectedResp func(*testing.T, *dto.CloseLastReceptionResponse, error)
	}{
		{
			name:  "success CloseReception",
			pvzId: pvzId,
			mockExpect: func() {
				rows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
					AddRow(pvzId, testTime, pvzId, "closed")
				mock.ExpectQuery(regexp.QuoteMeta(closeLastReception)).
					WithArgs(pvzId).
					WillReturnRows(rows)
			},
			expectedResp: func(t *testing.T, resp *dto.CloseLastReceptionResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "closed", resp.Status)
				assert.Equal(t, pvzId, resp.PvzId)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()
			resp, err := repo.CloseReception(context.Background(), tt.pvzId)
			tt.expectedResp(t, resp, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepository_DeleteLastProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &Repository{db: sqlx.NewDb(db, "sqlmock")}
	pvzId := uuid.MustParse("87c17529-99bb-4815-be06-900c4612902a")
	receptionId := uuid.MustParse("97c17529-99bb-4815-be06-900c4612902a")

	tests := []struct {
		name         string
		pvzId        uuid.UUID
		mockExpect   func()
		expectedResp func(*testing.T, error)
	}{
		{
			name:  "success DeleteLastProduct",
			pvzId: pvzId,
			mockExpect: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(getProductFromReception)).
					WithArgs(pvzId).
					WillReturnRows(sqlmock.NewRows([]string{"reception_id"}).AddRow(receptionId))
				mock.ExpectExec(regexp.QuoteMeta(deleteProduct)).
					WithArgs(receptionId).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedResp: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:  "no active reception",
			pvzId: pvzId,
			mockExpect: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(getProductFromReception)).
					WithArgs(pvzId).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectRollback()
			},
			expectedResp: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to find active reception")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()
			err := repo.DeleteLastProduct(context.Background(), tt.pvzId)
			tt.expectedResp(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepository_GetActiveReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &Repository{db: sqlx.NewDb(db, "sqlmock")}
	pvzId := uuid.MustParse("87c17529-99bb-4815-be06-900c4612902a")
	testTime := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name         string
		pvzID        uuid.UUID
		mockExpect   func()
		expectedResp func(*testing.T, *models.Reception, error)
	}{
		{
			name:  "success GetActiveReception",
			pvzID: pvzId,
			mockExpect: func() {
				rows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
					AddRow(pvzId, testTime, pvzId, "in_progress")
				mock.ExpectQuery(regexp.QuoteMeta(getActiveReception)).
					WithArgs(pvzId).
					WillReturnRows(rows)
			},
			expectedResp: func(t *testing.T, resp *models.Reception, err error) {
				assert.NoError(t, err)
				assert.Equal(t, pvzId, resp.PvzId)
				assert.Equal(t, models.Status("in_progress"), resp.Status)
			},
		},
		{
			name:  "reception not found",
			pvzID: pvzId,
			mockExpect: func() {
				mock.ExpectQuery(regexp.QuoteMeta(getActiveReception)).
					WithArgs(pvzId).
					WillReturnError(sql.ErrNoRows)
			},
			expectedResp: func(t *testing.T, resp *models.Reception, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "no active reception")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()
			resp, err := repo.GetActiveReception(context.Background(), tt.pvzID)
			tt.expectedResp(t, resp, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepository_GetPvz(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &Repository{db: sqlx.NewDb(db, "sqlmock")}
	pvzId := uuid.MustParse("87c17529-99bb-4815-be06-900c4612902a")
	testTime := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name         string
		mockExpect   func()
		expectedResp func(*testing.T, []*dto.PVZWithReceptions, error)
	}{
		{
			name: "success GetPvz with receptions",
			mockExpect: func() {
				rows := sqlmock.NewRows([]string{
					"pvz_id", "registration_date", "city",
					"reception_id", "reception_date", "status",
					"product_id", "product_date", "type",
				}).
					AddRow(
						pvzId, testTime, "Москва",
						pvzId, testTime, "closed",
						pvzId, testTime, "электроника",
					)

				mock.ExpectQuery(regexp.QuoteMeta(getPVZWithReceptions)).
					WithArgs(testTime, testTime, 10, 0).
					WillReturnRows(rows)
			},
			expectedResp: func(t *testing.T, resp []*dto.PVZWithReceptions, err error) {
				assert.NoError(t, err)
				assert.Len(t, resp, 1)
				assert.Len(t, resp[0].Receptions, 1)
				assert.Len(t, resp[0].Receptions[0].Products, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()
			resp, err := repo.GetPvz(context.Background(), testTime, testTime, 1, 10)
			tt.expectedResp(t, resp, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepository_DummyLogin(t *testing.T) {
	repo := &Repository{}

	tests := []struct {
		name     string
		role     string
		expected models.Role
	}{
		{"Employee role", "employee", models.RoleEmployee},
		{"Moderator role", "moderator", models.RoleModerator},
		{"Unknown role", "admin", models.Role("admin")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.DummyLogin(context.Background(), tt.role)

			assert.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, user.Id)
			assert.Equal(t, "dummy@example.com", user.Email)
			assert.Equal(t, tt.expected, user.Role)
			assert.Empty(t, user.Password)
		})
	}
}
