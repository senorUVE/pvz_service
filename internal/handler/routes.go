package handler

import (
	"github.com/senorUVE/pvz_service/internal/metrics"
	"github.com/senorUVE/pvz_service/internal/models"
)

func RegisterRoutes(h *PvzHandler) {
	h.e.Use(metrics.PrometheusMiddleware)

	authRouter := h.e.Group("/api")
	authRouter.POST("/register", h.Register)
	authRouter.POST("/login", h.Login)
	authRouter.POST("/dummyLogin", h.DummyLogin)
	authRouter.GET("/ping", h.Ping)

	pvzGroup := h.e.Group("/pvz")
	pvzGroup.Use(h.AuthMiddleware())
	{
		pvzGroup.POST("", h.CreatePVZ, h.RoleMiddleware(models.RoleModerator))
		pvzGroup.GET("", h.GetPvz, h.RoleMiddleware(models.RoleModerator, models.RoleEmployee))
		pvzGroup.POST("/:pvzId/close_last_reception", h.CloseReception, h.RoleMiddleware(models.RoleEmployee))
		pvzGroup.POST("/:pvzId/delete_last_product", h.DeleteLastProduct, h.RoleMiddleware(models.RoleEmployee))
	}
	receptionGroup := h.e.Group("/receptions")
	receptionGroup.Use(h.AuthMiddleware(), h.RoleMiddleware(models.RoleEmployee))
	{
		receptionGroup.POST("", h.CreateReception)
	}

	productGroup := h.e.Group("/products")
	productGroup.Use(h.AuthMiddleware(), h.RoleMiddleware(models.RoleEmployee))
	{
		productGroup.POST("", h.AddProduct)
	}
}
