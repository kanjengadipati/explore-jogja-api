package review

import (
	"errors"
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

func (h *Handler) callerIdentity(c *gin.Context) (userID string, isAdmin bool) {
	uid, _ := httpx.GetUserIDFromContext(c)
	userID = fmt.Sprintf("%v", uid)
	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	isAdmin = roleStr == "admin" || roleStr == "superadmin"
	return userID, isAdmin
}

func (h *Handler) GetAll(c *gin.Context) {
	if destID := c.Query("destination_id"); destID != "" {
		reviews, err := h.Service.GetByDestinationID(destID)
		if err != nil {
			httpx.HandleError(c, err)
			return
		}
		httpx.Success(c, 200, "Reviews fetched", reviews, nil)
		return
	}
	if userID := c.Query("user_id"); userID != "" {
		reviews, err := h.Service.GetByUserID(userID)
		if err != nil {
			httpx.HandleError(c, err)
			return
		}
		httpx.Success(c, 200, "Reviews fetched", reviews, nil)
		return
	}
	reviews, err := h.Service.GetAll()
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Reviews fetched", reviews, nil)
}

func (h *Handler) GetAllAdmin(c *gin.Context) {
	reviews, err := h.Service.GetAllAdmin()
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

	callerID, isAdmin := h.callerIdentity(c)
	review, err := h.Service.Update(id, callerID, isAdmin, req)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			httpx.ErrorWithCode(c, 403, "FORBIDDEN", "You can only edit your own review")
			return
		}
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Review updated", review, nil)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	callerID, isAdmin := h.callerIdentity(c)
	if err := h.Service.Delete(id, callerID, isAdmin); err != nil {
		if errors.Is(err, ErrForbidden) {
			httpx.ErrorWithCode(c, 403, "FORBIDDEN", "You can only delete your own review")
			return
		}
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Review deleted", nil, nil)
}
