package staging

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"pleco-api/internal/httpx"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) GetPendingDestinations(c *gin.Context) {
	dests, err := h.Service.Repo.FindPendingDestinations()
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "Failed to fetch pending destinations")
		return
	}
	httpx.Success(c, http.StatusOK, "Pending destinations fetched", dests, nil)
}

func (h *Handler) AIReviewDestinations(c *gin.Context) {
	var input struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	if err := h.Service.ReviewDestinations(c.Request.Context(), input.IDs); err != nil {
		httpx.Error(c, http.StatusInternalServerError, "AI review failed")
		return
	}
	httpx.Success(c, http.StatusOK, "AI review completed", nil, nil)
}

func (h *Handler) ApproveDestinations(c *gin.Context) {
	var input struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	if err := h.Service.Repo.ApproveMultipleDestinations(input.IDs); err != nil {
		httpx.Error(c, http.StatusInternalServerError, "Failed to approve destinations")
		return
	}
	httpx.Success(c, http.StatusOK, "Destinations approved", nil, nil)
}

func (h *Handler) RejectDestinations(c *gin.Context) {
	var input struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	if err := h.Service.Repo.RejectMultipleDestinations(input.IDs); err != nil {
		httpx.Error(c, http.StatusInternalServerError, "Failed to reject destinations")
		return
	}
	httpx.Success(c, http.StatusOK, "Destinations rejected", nil, nil)
}
