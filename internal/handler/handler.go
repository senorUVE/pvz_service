package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/senorUVE/pvz_service/internal/auth"
	"github.com/senorUVE/pvz_service/internal/dto"
	"github.com/senorUVE/pvz_service/internal/models"
)

type PvzService interface {
	GetUser(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, request *dto.RegisterRequest) (*models.User, error)
	AuthUser(ctx context.Context, request *dto.AuthRequest) (*dto.AuthResponse, error)
	CreatePVZ(ctx context.Context, request *dto.PvzCreateRequest) (*dto.PvzCreateResponse, error)
	GetPvz(ctx context.Context, request *dto.GetPvzRequest) ([]*dto.PVZWithReceptions, error)
	CloseReception(ctx context.Context, pvzID uuid.UUID) (*dto.CloseLastReceptionResponse, error)
	DeleteLastProduct(ctx context.Context, pvzId uuid.UUID) error
	CreateReception(ctx context.Context, request *dto.CreateReceptionRequest) (*dto.CreateReceptionResponse, error)
	AddProduct(ctx context.Context, request *dto.AddProductRequest) (*dto.AddProductResponse, error)
	DummyLogin(ctx context.Context, role string) (string, error)
}

type PvzHandler struct {
	e          *echo.Echo
	pvzService PvzService
	auth       auth.AuthService
	port       string
}

func NewPvzHandler(srv PvzService, auth auth.AuthService, port string) *PvzHandler {
	e := echo.New()
	return &PvzHandler{
		e:          e,
		pvzService: srv,
		auth:       auth,
		port:       port,
	}

}

func (h *PvzHandler) Start() error {
	RegisterRoutes(h)

	if err := h.e.Start(":" + h.port); err != nil && !errors.Is(err, http.ErrServerClosed) {
		h.e.Logger.Fatal("Shutting down the server")
	}

	return nil
}

func (h *PvzHandler) Close(ctx context.Context) error {
	return h.e.Shutdown(ctx)
}

func (h *PvzHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: "Invalid request"})
	}

	user, err := h.pvzService.CreateUser(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: err.Error()})
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *PvzHandler) Login(c echo.Context) error {
	var req dto.AuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: "Invalid request"})
	}

	response, err := h.pvzService.AuthUser(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Errors: err.Error()})
	}

	return c.JSON(http.StatusOK, response)
}

func (h *PvzHandler) CreatePVZ(c echo.Context) error {
	var req dto.PvzCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: "Invalid request"})
	}

	response, err := h.pvzService.CreatePVZ(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: err.Error()})
	}

	return c.JSON(http.StatusCreated, response)
}

func (h *PvzHandler) GetPvz(c echo.Context) error {
	var req dto.GetPvzRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: "Invalid request"})
	}

	// if startStr := c.QueryParam("startDate"); startStr != "" {
	// 	startDate, err := time.Parse(time.RFC3339, startStr)
	// 	if err != nil {
	// 		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: "validation failed: invalid startDate"})
	// 	}
	// 	req.StartDate = startDate
	// }

	// if endStr := c.QueryParam("endDate"); endStr != "" {
	// 	endDate, err := time.Parse(time.RFC3339, endStr)
	// 	if err != nil {
	// 		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: "validation failed: invalid endDate"})
	// 	}
	// 	req.EndDate = endDate
	// }

	// if p := c.QueryParam("page"); p != "" {
	// 	page, err := strconv.Atoi(p)
	// 	if err != nil {
	// 		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: "validation failed: invalid page"})
	// 	}
	// 	req.Page = page
	// } else {
	// 	req.Page = 1
	// }

	// if l := c.QueryParam("limit"); l != "" {
	// 	limit, err := strconv.Atoi(l)
	// 	if err != nil {
	// 		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: "validation failed: limit must be 1-100"})
	// 	}
	// 	req.Limit = limit
	// } else {
	// 	req.Limit = 10
	// }

	response, err := h.pvzService.GetPvz(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Errors: err.Error()})
	}

	return c.JSON(http.StatusOK, response)
}

func (h *PvzHandler) CreateReception(c echo.Context) error {
	var req dto.CreateReceptionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: "Invalid request"})
	}

	response, err := h.pvzService.CreateReception(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: err.Error()})
	}

	return c.JSON(http.StatusCreated, response)
}

func (h *PvzHandler) AddProduct(c echo.Context) error {
	var req dto.AddProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: "Invalid request"})
	}

	response, err := h.pvzService.AddProduct(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: err.Error()})
	}

	return c.JSON(http.StatusCreated, response)
}

func (h *PvzHandler) CloseReception(c echo.Context) error {
	pvzID, err := uuid.Parse(c.Param("pvzId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: "Invalid PVZ ID"})
	}

	response, err := h.pvzService.CloseReception(c.Request().Context(), pvzID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: err.Error()})
	}

	return c.JSON(http.StatusOK, response)
}

func (h *PvzHandler) DeleteLastProduct(c echo.Context) error {
	pvzID, err := uuid.Parse(c.Param("pvzId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: "Invalid PVZ ID"})
	}

	if err := h.pvzService.DeleteLastProduct(c.Request().Context(), pvzID); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: err.Error()})
	}

	return c.NoContent(http.StatusOK)
}

func (h *PvzHandler) DummyLogin(c echo.Context) error {
	var req dto.DummyLoginRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Errors: "Invalid request"})
	}

	token, err := h.pvzService.DummyLogin(c.Request().Context(), req.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Errors: err.Error()})
	}
	return c.JSON(http.StatusOK, token)
}

func (h *PvzHandler) Ping(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "pong")
}

func (h *PvzHandler) GetEcho() *echo.Echo {
	return h.e
}
