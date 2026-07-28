package notification

import (
	"pleco-api/internal/httpx"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) GetNotifications(c *gin.Context) {
	userID, _ := httpx.GetUserIDFromContext(c)
	notifs, err := h.Service.GetUserNotifications(userID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Notifications fetched", notifs, nil)
}

func (h *Handler) MarkRead(c *gin.Context) {
	userID, _ := httpx.GetUserIDFromContext(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid ID")
		return
	}
	if err := h.Service.MarkAsRead(uint(id), userID); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Notification marked as read", nil, nil)
}
