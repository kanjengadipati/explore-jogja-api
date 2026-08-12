package tourist

import (
	"pleco-api/internal/ai"
	"pleco-api/internal/cache"
	"pleco-api/internal/modules/destination"
	"pleco-api/internal/modules/event"
	"pleco-api/internal/search"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler) {
	aiGroup := api.Group("/ai")
	aiGroup.POST("/query", handler.Query)
	aiGroup.POST("/image-search", handler.ImageSearch)
	aiGroup.GET("/recommend", handler.Recommend)
	aiGroup.GET("/recommend/multi", handler.MultiRecommend)
	aiGroup.POST("/journey", handler.Journey)
	aiGroup.POST("/generate-destination", handler.GenerateDestination)
	aiGroup.POST("/generate-event", handler.GenerateEvent)
	aiGroup.POST("/generate-article", handler.GenerateArticle)
	aiGroup.GET("/trending", handler.Trending)
	aiGroup.GET("/route-timeline", handler.RouteTimeline)
	aiGroup.GET("/next-stop", handler.NextStop)
}

type Module struct {
	Handler *Handler
}

func BuildModule(aiService *ai.Service, destRepo destination.Repository, eventRepo event.Repository, cacheStore cache.Store, searchClient *search.Client) *Module {
	handler := NewHandler(aiService, destRepo, eventRepo, cacheStore, searchClient)
	return &Module{Handler: handler}
}
