package imagereport

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"pleco-api/internal/httpx"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

// CreateReport handles POST /destinations/:id/report
func (h *Handler) CreateReport(c *gin.Context) {
	destID := c.Param("id")

	var input struct {
		Reason   string `json:"reason" binding:"required"`
		Details  string `json:"details"`
		ImageURL string `json:"image_url"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	userID, _ := httpx.GetUserIDFromContext(c)
	userName := ""
	if u, ok := c.Get("user_name"); ok {
		userName, _ = u.(string)
	}

	report := &ImageReport{
		DestinationID: destID,
		ImageURL:      input.ImageURL,
		UserID:        userID,
		UserName:      userName,
		Reason:        input.Reason,
		Details:       input.Details,
	}

	if err := h.Service.CreateReport(report); err != nil {
		httpx.Error(c, http.StatusInternalServerError, "Failed to submit report")
		return
	}

	httpx.Success(c, http.StatusCreated, "Report submitted", report, nil)
}

// GetAll handles GET /admin/image-reports
func (h *Handler) GetAll(c *gin.Context) {
	status := c.Query("status")

	var reports []ImageReport
	var err error

	if status != "" {
		reports, err = h.Service.GetByStatus(status)
	} else {
		reports, err = h.Service.GetAll()
	}

	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "Failed to fetch reports")
		return
	}

	httpx.Success(c, http.StatusOK, "Reports fetched", reports, nil)
}

// GetStats handles GET /admin/image-reports/stats
func (h *Handler) GetStats(c *gin.Context) {
	pending, resolved, dismissed, err := h.Service.GetStats()
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "Failed to fetch stats")
		return
	}

	httpx.Success(c, http.StatusOK, "Stats fetched", gin.H{
		"pending":   pending,
		"resolved":  resolved,
		"dismissed": dismissed,
	}, nil)
}

// Resolve handles POST /admin/image-reports/:id/resolve
func (h *Handler) Resolve(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "Invalid report ID")
		return
	}

	if err := h.Service.Resolve(uint(id)); err != nil {
		httpx.Error(c, http.StatusInternalServerError, "Failed to resolve report")
		return
	}

	httpx.Success(c, http.StatusOK, "Report resolved", nil, nil)
}

// Dismiss handles POST /admin/image-reports/:id/dismiss
func (h *Handler) Dismiss(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "Invalid report ID")
		return
	}

	if err := h.Service.Dismiss(uint(id)); err != nil {
		httpx.Error(c, http.StatusInternalServerError, "Failed to dismiss report")
		return
	}

	httpx.Success(c, http.StatusOK, "Report dismissed", nil, nil)
}
