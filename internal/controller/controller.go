package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/senorUVE/pvz_service/internal/auth"
	"github.com/senorUVE/pvz_service/internal/dto"
	"github.com/senorUVE/pvz_service/internal/models"
	"github.com/senorUVE/pvz_service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	GetUser(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, email, password, role string) (uuid.UUID, error)
	CreatePvz(ctx context.Context, pvz models.PVZ) (*dto.PvzCreateResponse, error)
	CreateReception(ctx context.Context, pvzId uuid.UUID) (*dto.CreateReceptionResponse, error)
	CreateProduct(ctx context.Context, typeOf string, receptionId uuid.UUID) (*dto.AddProductResponse, error)
	CloseReception(ctx context.Context, pvzId uuid.UUID) (*dto.CloseLastReceptionResponse, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
	GetPvz(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*dto.PVZWithReceptions, error)
	GetActiveReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	DummyLogin(ctx context.Context, role string) (*models.User, error)
}

type PvzService struct {
	repo Repository
	auth auth.AuthService
	cfg  ServiceConfig
}

func NewPvzService(repo Repository, auth auth.AuthService, cfg ServiceConfig) *PvzService {
	return &PvzService{
		repo: repo,
		auth: auth,
		cfg:  cfg,
	}
}

func (p *PvzService) GetUser(ctx context.Context, email string) (*models.User, error) {
	return p.repo.GetUser(ctx, email)
}

func (p *PvzService) generatePasswordHash(password string) (string, error) {
	var passwordBytes = []byte(password + p.cfg.Salt)

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword(passwordBytes, p.cfg.Cost)

	return string(hashedPasswordBytes), err
}

func (p *PvzService) CreateUser(ctx context.Context, request *dto.RegisterRequest) (*models.User, error) {
	password_hash, err := p.generatePasswordHash(request.Password)
	if err != nil {
		return nil, err
	}
	if err = ValidateRegisterRequest(request); err != nil {
		return nil, err
	}
	userResponse, err := p.repo.CreateUser(ctx, request.Email, password_hash, request.Role)
	if err != nil {
		return nil, err
	}
	user := models.User{
		Id:    userResponse,
		Email: request.Email,
		Role:  models.Role(request.Role),
	}
	return &user, nil
}

func (p *PvzService) AuthUser(ctx context.Context, request *dto.AuthRequest) (*dto.AuthResponse, error) {
	if err := ValidateAuth(request); err != nil {
		return nil, err
	}

	user, err := p.repo.GetUser(ctx, request.Email)
	if err != nil && errors.Is(err, repository.ErrUserNotFound) {
		token, err := p.auth.GenerateToken(user)
		if err != nil {
			return nil, err
		}
		return &dto.AuthResponse{
			Token: token,
		}, nil
	}

	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password+p.cfg.Salt)); err != nil {
		return nil, ErrInvalidPasswd
	}
	token, err := p.auth.GenerateToken(user)
	if err != nil {
		return nil, err
	}
	return &dto.AuthResponse{Token: token}, nil
}

// func (p *PvzService) CreatePvz(ctx context.Context, )

// func (p *PvzService) Register(ctx context.Context, request *dto.RegisterRequest) (*dto.RegisterResponse, error) {
// 	if err := ValidateRegisterRequest(request); err != nil {
// 		return nil, err
// 	}

// 	hashedPassword, err := bcrypt.GenerateFromPassword(
// 		[]byte(request.Password+p.cfg.Salt),
// 		p.cfg.Cost,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	user := &models.User{
// 		Id:       uuid.New(),
// 		Email:    request.Email,
// 		Password: string(hashedPassword),
// 		Role:     models.Role(request.Role),
// 	}

// 	if _, err := p.repo.CreateUser(ctx, user); err != nil {
// 		return nil, err
// 	}

// 	return &dto.RegisterResponse{
// 		Id:    user.Id,
// 		Email: user.Email,
// 		Role:  string(user.Role),
// 	}, nil
// }

func (p *PvzService) CreatePVZ(ctx context.Context, request *dto.PvzCreateRequest) (*dto.PvzCreateResponse, error) {
	if err := ValidatePvzCreateRequest(request); err != nil {
		return nil, err
	}

	pvz := models.PVZ{
		Id:               uuid.New(),
		RegistrationDate: time.Now().UTC(),
		City:             models.City(request.City),
	}

	created, err := p.repo.CreatePvz(ctx, pvz)
	if err != nil {
		return nil, err
	}

	return &dto.PvzCreateResponse{
		Id:               created.Id,
		RegistrationDate: created.RegistrationDate,
		City:             created.City,
	}, nil
}

func (p *PvzService) GetPvz(ctx context.Context, request *dto.GetPvzRequest) ([]*dto.PVZWithReceptions, error) {
	if err := ValidateGetPvzRequest(request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	filter := dto.GetPvzRequest{
		StartDate: request.StartDate,
		EndDate:   request.EndDate,
		Page:      request.Page,
		Limit:     request.Limit,
	}

	result, err := p.repo.GetPvz(ctx, filter.StartDate, filter.EndDate, filter.Page, filter.Limit)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *PvzService) CreateReception(ctx context.Context, request *dto.CreateReceptionRequest) (*dto.CreateReceptionResponse, error) {

	reception := &models.Reception{
		Id:       uuid.New(),
		DateTime: time.Now().UTC(),
		PvzId:    request.PvzId,
		Status:   models.StatusInProgress,
	}

	created, err := p.repo.CreateReception(ctx, reception.PvzId)
	if err != nil {
		return nil, err
	}

	return &dto.CreateReceptionResponse{
		Id:       created.Id,
		DateTime: created.DateTime,
		PvzId:    request.PvzId,
		Status:   created.Status,
	}, nil
}

func (p *PvzService) CloseReception(ctx context.Context, pvzID uuid.UUID) (*dto.CloseLastReceptionResponse, error) {
	reception, err := p.repo.CloseReception(ctx, pvzID)
	if err != nil {
		return nil, err
	}

	return &dto.CloseLastReceptionResponse{
		Id:       reception.Id,
		DateTime: reception.DateTime,
		PvzId:    reception.PvzId,
		Status:   string(reception.Status),
	}, nil
}

func (p *PvzService) AddProduct(ctx context.Context, request *dto.AddProductRequest) (*dto.AddProductResponse, error) {
	if err := ValidateAddProductRequest(request); err != nil {
		return nil, err
	}

	activeReception, err := p.repo.GetActiveReception(ctx, request.PvzId)
	if err != nil {
		return nil, err
	}

	product := &models.Product{
		Id:       uuid.New(),
		DateTime: time.Now().UTC(),
		Type:     models.Type(request.Type),
	}

	created, err := p.repo.CreateProduct(ctx, string(product.Type), activeReception.Id)
	if err != nil {
		return nil, err
	}

	return &dto.AddProductResponse{
		Id:          created.Id,
		DateTime:    created.DateTime,
		Type:        created.Type,
		ReceptionId: created.ReceptionId,
	}, nil
}

func (p *PvzService) DeleteLastProduct(ctx context.Context, pvzId uuid.UUID) error {
	activeReception, err := p.repo.GetActiveReception(ctx, pvzId)
	if err != nil {
		if activeReception == nil {
			return fmt.Errorf("no active reception: %w", err)
		}

	}
	return p.repo.DeleteLastProduct(ctx, pvzId)
}

func (p *PvzService) DummyLogin(ctx context.Context, role string) (string, error) {
	dummyUser, err := p.repo.DummyLogin(ctx, role)

	if err != nil {
		return "", err
	}

	token, err := p.auth.GenerateToken(dummyUser)
	if err != nil {
		return "", err
	}
	return token, nil
}
