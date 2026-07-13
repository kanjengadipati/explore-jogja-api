package destination

import (
	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) GetAll(c *gin.Context) {
	dests, err := h.Service.GetAll()
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Destinations fetched", dests, nil)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	dest, err := h.Service.GetByID(id)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Destination not found")
		return
	}
	httpx.Success(c, 200, "Destination fetched", dest, nil)
}

func (h *Handler) GetByCategory(c *gin.Context) {
	category := c.Param("category")
	dests, err := h.Service.GetByCategory(category)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Destinations fetched", dests, nil)
}

func (h *Handler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Query parameter 'q' is required")
		return
	}
	dests, err := h.Service.Search(query)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Search results", dests, nil)
}
