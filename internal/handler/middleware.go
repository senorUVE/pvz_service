package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/senorUVE/pvz_service/internal/dto"
	"github.com/senorUVE/pvz_service/internal/models"
	"github.com/sirupsen/logrus"
)

const (
	authorizationHeader = "Authorization"
)

func (h *PvzHandler) AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			const op = "internal.handler.AuthMiddleware"
			header := ctx.Request().Header.Get(authorizationHeader)
			if header == "" {
				return ctx.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Errors: ErrEmptyToken.Error()})
			}

			headerSplit := strings.Split(header, " ")
			if len(headerSplit) != 2 {
				return ctx.JSON(http.StatusUnauthorized, dto.UnauthorizedResponse{Errors: ErrInvalidAuthHeader.Error()})
			}

			user, err := h.auth.ParseToken(headerSplit[1])
			if err != nil {
				logrus.WithFields(logrus.Fields{"event": op}).Error(err)

				return ctx.JSON(http.StatusInternalServerError, dto.InternalServerErrorResponse{Errors: ErrInternalServer.Error()})
			}

			ctx.Set("user", user)
			return next(ctx)
		}
	}
}

func (h *PvzHandler) RoleMiddleware(roles ...models.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get("user").(*models.User)
			if !ok {
				return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Errors: "User not found in context"})
			}

			for _, role := range roles {
				if user.Role == role {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, dto.ErrorResponse{Errors: "Insufficient permissions"})
		}
	}
}
