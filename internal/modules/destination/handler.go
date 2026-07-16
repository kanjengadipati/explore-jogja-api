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

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateDestinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}

	dest, err := h.Service.Update(id, req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Destination updated", dest, nil)
}

func (h *Handler) UpdateUserDestinationStatus(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, 401, "Unauthorized")
		return
	}
	slug := c.Param("slug")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, 400, "Invalid request")
		return
	}
	if err := h.Service.UpdateUserDestination(userID, slug, req.Status); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Status updated", nil, nil)
}

func (h *Handler) GetUserDestinations(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, 401, "Unauthorized")
		return
	}
	uds, err := h.Service.GetUserDestinations(userID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "User destinations fetched", uds, nil)
}
