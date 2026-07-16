package config

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler) {
	cfg := api.Group("/config")
	cfg.GET("/categories", handler.GetCategories)
	cfg.GET("/sub-regions", handler.GetSubRegions)
	cfg.GET("/quotes", handler.GetQuotes)
	cfg.GET("/seo", handler.GetSiteConfig)
	cfg.PUT("/seo", handler.UpdateSiteConfig)
}
