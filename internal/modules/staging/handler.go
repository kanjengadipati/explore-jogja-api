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
	dests, err := h.Service.Repo.FindPendingDestinations(c.Query("source"))
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

	results, err := h.Service.ReviewDestinations(c.Request.Context(), input.IDs)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "AI review failed")
		return
	}
	httpx.Success(c, http.StatusOK, "AI review completed", results, nil)
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

func (h *Handler) GetPendingEvents(c *gin.Context) {
	events, err := h.Service.Repo.FindPendingEvents(c.Query("source"))
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "Failed to fetch pending events")
		return
	}
	httpx.Success(c, http.StatusOK, "Pending events fetched", events, nil)
}

func (h *Handler) AIReviewEvents(c *gin.Context) {
	var input struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	results, err := h.Service.ReviewEvents(c.Request.Context(), input.IDs)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "AI review failed")
		return
	}
	httpx.Success(c, http.StatusOK, "AI review completed", results, nil)
}

func (h *Handler) ApproveEvents(c *gin.Context) {
	var input struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	if err := h.Service.Repo.ApproveMultipleEvents(input.IDs); err != nil {
		httpx.Error(c, http.StatusInternalServerError, "Failed to approve events")
		return
	}
	httpx.Success(c, http.StatusOK, "Events approved", nil, nil)
}

func (h *Handler) RejectEvents(c *gin.Context) {
	var input struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	if err := h.Service.Repo.RejectMultipleEvents(input.IDs); err != nil {
		httpx.Error(c, http.StatusInternalServerError, "Failed to reject events")
		return
	}
	httpx.Success(c, http.StatusOK, "Events rejected", nil, nil)
}
