package review

import (
	"fmt"
	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) GetAll(c *gin.Context) {
	reviews, err := h.Service.GetAll()
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Reviews fetched", reviews, nil)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	review, err := h.Service.GetByID(id)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Review not found")
		return
	}
	httpx.Success(c, 200, "Review fetched", review, nil)
}

func (h *Handler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Query parameter 'q' is required")
		return
	}
	reviews, err := h.Service.Search(query)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Search results", reviews, nil)
}

func (h *Handler) Create(c *gin.Context) {
	var review Review
	if err := c.ShouldBindJSON(&review); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}

	// Auto-generate ExternalID if not provided
	if review.ExternalID == "" {
		review.ExternalID = uuid.NewString()
	}

	// Set UserID from JWT context
	if userID, exists := c.Get("user_id"); exists {
		review.UserID = fmt.Sprintf("%v", userID)
	}

	// Default status
	if review.Status == "" {
		review.Status = "published"
	}

	if err := h.Service.Create(&review); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 201, "Review created", review, nil)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}

	review, err := h.Service.Update(id, req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Review updated", review, nil)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.Delete(id); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Review deleted", nil, nil)
}
