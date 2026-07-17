package tourist

import (
	"pleco-api/internal/ai"
	"pleco-api/internal/modules/destination"
	"pleco-api/internal/modules/event"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler) {
	aiGroup := api.Group("/ai")
	aiGroup.POST("/query", handler.Query)
	aiGroup.POST("/image-search", handler.ImageSearch)
	aiGroup.GET("/recommend", handler.Recommend)
	aiGroup.POST("/journey", handler.Journey)
	aiGroup.GET("/trending", handler.Trending)
}

type Module struct {
	Handler *Handler
}

func BuildModule(aiService *ai.Service, destRepo destination.Repository, eventRepo event.Repository) *Module {
	handler := NewHandler(aiService, destRepo, eventRepo)
	return &Module{Handler: handler}
}
