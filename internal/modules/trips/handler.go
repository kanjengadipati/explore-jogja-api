package trips

import (
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

// GetAll returns all trips belonging to the authenticated user.
func (h *Handler) GetAll(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.ErrorWithCode(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	trips, err := h.Service.GetByUser(userID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Trips fetched", trips, nil)
}

// GetByID returns a single trip owned by the authenticated user.
func (h *Handler) GetByID(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.ErrorWithCode(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	id := c.Param("id")
	trip, err := h.Service.GetByID(id, userID)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Trip not found")
		return
	}
	httpx.Success(c, 200, "Trip fetched", trip, nil)
}

// Create saves a new trip for the authenticated user.
func (h *Handler) Create(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.ErrorWithCode(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	var trip Trip
	if err := c.ShouldBindJSON(&trip); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}

	trip.ExternalID = uuid.NewString()
	trip.UserID = userID

	if trip.Title == "" {
		trip.Title = "My Trip"
	}
	if trip.Status == "" {
		trip.Status = "draft"
	}
	if trip.DurationDays == 0 {
		trip.DurationDays = len(trip.Days)
		if trip.DurationDays == 0 {
			trip.DurationDays = 1
		}
	}

	if err := h.Service.Create(&trip); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 201, "Trip created", trip, nil)
}

// Update applies a partial update to a trip owned by the authenticated user.
func (h *Handler) Update(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.ErrorWithCode(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	id := c.Param("id")

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}

	trip, err := h.Service.Update(id, userID, req)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", err.Error())
		return
	}
	httpx.Success(c, 200, "Trip updated", trip, nil)
}

// Delete soft-deletes a trip owned by the authenticated user.
func (h *Handler) Delete(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.ErrorWithCode(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	id := c.Param("id")
	if err := h.Service.Delete(id, userID); err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Trip not found")
		return
	}
	httpx.Success(c, 200, "Trip deleted", nil, nil)
}
