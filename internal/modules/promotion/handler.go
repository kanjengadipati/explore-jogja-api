package promotion

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
	promotions, err := h.Service.GetAll()
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Promotions fetched", promotions, nil)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	promotion, err := h.Service.GetByID(id)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Promotion not found")
		return
	}
	httpx.Success(c, 200, "Promotion fetched", promotion, nil)
}

func (h *Handler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Query parameter 'q' is required")
		return
	}
	promotions, err := h.Service.Search(query)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Search results", promotions, nil)
}

func (h *Handler) Create(c *gin.Context) {
	var promotion Promotion
	if err := c.ShouldBindJSON(&promotion); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}
	if err := h.Service.Create(&promotion); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 201, "Promotion created", promotion, nil)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}

	promotion, err := h.Service.Update(id, req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Promotion updated", promotion, nil)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.Delete(id); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Promotion deleted", nil, nil)
}

// --- Admin promotion approval ---

func (h *Handler) AdminApprove(c *gin.Context) {
	id := c.Param("id")
	approved := "approved"
	promo, err := h.Service.Update(id, UpdatePromotionRequest{Status: &approved})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Promotion approved", promo, nil)
}

func (h *Handler) AdminReject(c *gin.Context) {
	id := c.Param("id")
	rejected := "rejected"
	promo, err := h.Service.Update(id, UpdatePromotionRequest{Status: &rejected})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Promotion rejected", promo, nil)
}

func (h *Handler) AdminGetPending(c *gin.Context) {
	promotions, err := h.Service.GetByStatus("pending")
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Pending promotions fetched", promotions, nil)
}
