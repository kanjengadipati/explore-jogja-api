package destination

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler) {
	dest := api.Group("/destinations")
	dest.GET("", handler.GetAll)
	dest.GET("/search", handler.Search)
	dest.GET("/:id", handler.GetByID)
	dest.GET("/category/:category", handler.GetByCategory)
}
