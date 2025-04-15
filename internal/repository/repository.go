package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/senorUVE/pvz_service/internal/dto"
	"github.com/senorUVE/pvz_service/internal/models"
	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

type Repository struct {
	db  *sqlx.DB
	cfg DBConfig
}

func NewRepository(config DBConfig) (*Repository, error) {
	db, err := sqlx.Open(
		"postgres", fmt.Sprintf(
			"%s://%s:%s@%s:%s/%s?sslmode=%s",
			config.DBDriver,
			config.DBUser,
			config.DBPass,
			config.DBHost,
			config.DBPort,
			config.DBName,
			config.DBSSL,
		),
	)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &Repository{
		db:  db,
		cfg: config,
	}, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) GetUser(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowxContext(ctx, getUserByEmail, email).StructScan(&user)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) CreateUser(ctx context.Context, email, password, role string) (uuid.UUID, error) {
	newUUID := uuid.New()
	err := r.db.QueryRowxContext(ctx, createUser, newUUID, email, password, role).Scan(&newUUID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create user: %w", err)
	}
	return newUUID, nil
}

func (r *Repository) CreatePvz(ctx context.Context, pvz models.PVZ) (*dto.PvzCreateResponse, error) {
	newUUID := uuid.New()
	err := r.db.QueryRowxContext(ctx, createPVZ, newUUID, pvz.RegistrationDate, pvz.City).Scan(&newUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pvz: %w", err)
	}
	return &dto.PvzCreateResponse{
		Id:               newUUID,
		RegistrationDate: pvz.RegistrationDate,
		City:             string(pvz.City),
	}, nil
}

func (r *Repository) CreateReception(ctx context.Context, pvzId uuid.UUID) (*dto.CreateReceptionResponse, error) {
	newUUID := uuid.New()
	currentTime := time.Now().UTC().Truncate(time.Second)
	var createdReception struct {
		ID       uuid.UUID
		DateTime time.Time
		Status   string
	}
	err := r.db.QueryRowxContext(ctx, createReception, newUUID, currentTime, pvzId).Scan(&createdReception.ID, &createdReception.DateTime, &createdReception.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to create reception: %w", err)
	}
	return &dto.CreateReceptionResponse{
		Id:       newUUID,
		DateTime: currentTime,
		PvzId:    pvzId,
		Status:   createdReception.Status,
	}, nil
}

func (r *Repository) CreateProduct(ctx context.Context, typeOf string, receptionId uuid.UUID) (*dto.AddProductResponse, error) {
	newUUID := uuid.New()
	currentTime := time.Now().UTC().Truncate(time.Second)
	err := r.db.QueryRowxContext(ctx, createProduct, newUUID, currentTime, typeOf, receptionId).Scan(&newUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}
	return &dto.AddProductResponse{
		Id:          newUUID,
		DateTime:    currentTime,
		Type:        typeOf,
		ReceptionId: receptionId,
	}, nil
}

func (r *Repository) GetActiveReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	var reception models.Reception
	err := r.db.GetContext(ctx, &reception, getActiveReception, pvzID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no active reception for pvz %s: %w", pvzID, ErrNoActiveReception)
		}
		return nil, err
	}

	return &reception, nil
}

func (r *Repository) CloseReception(ctx context.Context, pvzId uuid.UUID) (*dto.CloseLastReceptionResponse, error) {
	//newUUID := uuid.New()
	//currentTime := time.Now().UTC()
	var closedReception struct {
		Id       uuid.UUID
		DateTime time.Time
		pvzId    uuid.UUID
		Status   string
	}
	err := r.db.QueryRowxContext(ctx, closeLastReception, pvzId).Scan(&closedReception.Id, &closedReception.DateTime, &closedReception.pvzId, &closedReception.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to close reception: %w", err)
	}
	return &dto.CloseLastReceptionResponse{
		Id:       closedReception.Id,
		DateTime: closedReception.DateTime,
		PvzId:    closedReception.pvzId,
		Status:   closedReception.Status,
	}, nil
}

func (r *Repository) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	const ab = "internal.repository.DeleteLastProduct"
	tx, err := r.db.BeginTxx(ctx, nil)
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			logrus.WithFields(logrus.Fields{"event": ab}).Error(err)
		}
	}()

	if err != nil {
		return err
	}

	var receptionId uuid.UUID
	err = tx.QueryRowxContext(ctx, getProductFromReception, pvzID).Scan(&receptionId)
	if err != nil {
		return fmt.Errorf("failed to find active reception: %w", err)
	}
	_, err = tx.ExecContext(ctx, deleteProduct, receptionId)

	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *Repository) GetPvz(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*dto.PVZWithReceptions, error) {
	offset := (page - 1) * limit
	rows, err := r.db.QueryxContext(ctx, getPVZWithReceptions, startDate, endDate, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query pvz list: %w", err)
	}
	defer rows.Close()
	pvzMap := make(map[uuid.UUID]*dto.PVZWithReceptions)
	for rows.Next() {
		var (
			pvzId            uuid.UUID
			registrationDate time.Time
			city             string

			recId       uuid.NullUUID
			recDateTime sql.NullTime
			recStatus   sql.NullString

			prodId       uuid.NullUUID
			prodDateTime sql.NullTime
			prodType     sql.NullString
		)
		err = rows.Scan(&pvzId, &registrationDate, &city, &recId, &recDateTime, &recStatus, &prodId, &prodDateTime, &prodType)
		if err != nil {
			return nil, err
		}

		pvzRec, exists := pvzMap[pvzId]
		if !exists {
			pvzRec = &dto.PVZWithReceptions{
				PVZ: dto.PVZResponse{
					Id:               pvzId,
					RegistrationDate: registrationDate,
					City:             city,
				},
				Receptions: []dto.ReceptionWithProducts{},
			}
			pvzMap[pvzId] = pvzRec
		}
		if recId.Valid {
			var recPtr *dto.ReceptionWithProducts
			for i := range pvzRec.Receptions {
				if pvzRec.Receptions[i].Reception.Id == recId.UUID {
					recPtr = &pvzRec.Receptions[i]
					break
				}
			}

			if recPtr == nil {
				newRec := dto.ReceptionWithProducts{
					Reception: dto.ReceptionResponse{
						Id:       recId.UUID,
						DateTime: recDateTime.Time,
						PvzId:    pvzId,
						Status:   recStatus.String,
					},
					Products: []dto.ProductResponse{},
				}
				pvzRec.Receptions = append(pvzRec.Receptions, newRec)
				recPtr = &pvzRec.Receptions[len(pvzRec.Receptions)-1]
			}
			if prodId.Valid {
				product := dto.ProductResponse{
					Id:          prodId.UUID,
					DateTime:    prodDateTime.Time,
					Type:        prodType.String,
					ReceptionId: recId.UUID,
				}
				recPtr.Products = append(recPtr.Products, product)
			}
		}
	}
	var result []*dto.PVZWithReceptions
	for _, pvz := range pvzMap {
		result = append(result, pvz)
	}
	return result, nil
}

func (r *Repository) DummyLogin(ctx context.Context, role string) (*models.User, error) {
	dummyUser := &models.User{
		Id:       uuid.New(),
		Email:    "dummy@example.com",
		Password: "",
		Role:     models.Role(role),
	}
	return dummyUser, nil
}
