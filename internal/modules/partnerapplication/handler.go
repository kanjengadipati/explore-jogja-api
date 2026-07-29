package partnerapplication

import (
	"strings"

	"github.com/gin-gonic/gin"

	"pleco-api/internal/httpx"
)

type Handler struct{ Service *Service }

func NewHandler(service *Service) *Handler { return &Handler{Service: service} }

func (h *Handler) Apply(c *gin.Context) {
	var req ApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ValidationError(c, err)
		return
	}
	userID, _ := httpx.GetUserIDFromContext(c)
	app, err := h.Service.Apply(req, userID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 201, "Application submitted", app, nil)
}

func (h *Handler) GetMine(c *gin.Context) {
	userID, _ := httpx.GetUserIDFromContext(c)
	apps, err := h.Service.GetOwned(userID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Applications fetched", apps, nil)
}

func (h *Handler) AdminGetPending(c *gin.Context) {
	apps, err := h.Service.GetByStatus(StatusPending)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Pending applications fetched", apps, nil)
}

func (h *Handler) AdminApprove(c *gin.Context) {
	id := c.Param("id")
	adminUserID, _ := httpx.GetUserIDFromContext(c)
	app, err := h.Service.Approve(id, adminUserID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Application approved", app, nil)
}

type rejectRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) AdminReject(c *gin.Context) {
	id := c.Param("id")
	adminUserID, _ := httpx.GetUserIDFromContext(c)

	var body rejectRequest
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Reason) == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "reason is required")
		return
	}
	app, err := h.Service.Reject(id, body.Reason, adminUserID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Application rejected", app, nil)
}
