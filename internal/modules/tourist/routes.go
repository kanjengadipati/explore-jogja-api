package tourist

import (
	"pleco-api/internal/ai"
	"pleco-api/internal/modules/destination"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler) {
	aiGroup := api.Group("/ai")
	aiGroup.POST("/query", handler.Query)
	aiGroup.POST("/image-search", handler.ImageSearch)
	aiGroup.GET("/recommend", handler.Recommend)
}

type Module struct {
	Handler *Handler
}

func BuildModule(aiService *ai.Service, destRepo destination.Repository) *Module {
	handler := NewHandler(aiService, destRepo)
	return &Module{Handler: handler}
}
