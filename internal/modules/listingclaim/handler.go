package listingclaim

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"pleco-api/internal/httpx"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) Submit(c *gin.Context) {
	var req SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ValidationError(c, err)
		return
	}
	userID, _ := httpx.GetUserIDFromContext(c)

	claim, err := h.Service.Submit(req, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotBusinessOwner):
			httpx.ErrorWithCode(c, http.StatusForbidden, "FORBIDDEN", "user is not an owner of this business")
		case errors.Is(err, ErrListingNotOwnable):
			httpx.ErrorWithCode(c, http.StatusBadRequest, "INVALID_LISTING_TYPE", "unsupported listing type")
		default:
			httpx.HandleError(c, err)
		}
		return
	}
	httpx.Success(c, http.StatusCreated, "Claim submitted", claim, nil)
}

func (h *Handler) GetMine(c *gin.Context) {
	userID, _ := httpx.GetUserIDFromContext(c)

	var businessIDs []uint
	if h.Service.BizRepo != nil {
		businesses, err := h.Service.BizRepo.FindByOwner(userID)
		if err != nil {
			httpx.HandleError(c, err)
			return
		}
		for _, b := range businesses {
			businessIDs = append(businessIDs, b.ID)
		}
	}

	claims, err := h.Service.GetOwned(businessIDs)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	if claims == nil {
		claims = make([]ListingClaim, 0)
	}
	httpx.Success(c, http.StatusOK, "Claims fetched", claims, nil)
}

func (h *Handler) AdminGetPending(c *gin.Context) {
	claims, err := h.Service.GetPending()
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, "Pending claims fetched", claims, nil)
}

func (h *Handler) AdminApprove(c *gin.Context) {
	id := c.Param("id")
	adminUserID, _ := httpx.GetUserIDFromContext(c)

	if err := h.Service.Approve(id, adminUserID); err != nil {
		switch {
		case errors.Is(err, ErrClaimNotFound):
			httpx.ErrorWithCode(c, http.StatusNotFound, "NOT_FOUND", "pending claim not found")
		case errors.Is(err, ErrListingAlreadyOwned):
			httpx.ErrorWithCode(c, http.StatusConflict, "CONFLICT", "listing is already claimed by another business")
		default:
			httpx.HandleError(c, err)
		}
		return
	}
	httpx.Success(c, http.StatusOK, "Claim approved", nil, nil)
}

type rejectRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) AdminReject(c *gin.Context) {
	id := c.Param("id")
	adminUserID, _ := httpx.GetUserIDFromContext(c)

	var body rejectRequest
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Reason) == "" {
		httpx.ErrorWithCode(c, http.StatusBadRequest, "VALIDATION_FAILED", "reason is required")
		return
	}

	if err := h.Service.Reject(id, body.Reason, adminUserID); err != nil {
		if errors.Is(err, ErrClaimNotFound) {
			httpx.ErrorWithCode(c, http.StatusNotFound, "NOT_FOUND", "pending claim not found")
			return
		}
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, "Claim rejected", nil, nil)
}

func (h *Handler) SearchListings(c *gin.Context) {
	query := c.Query("q")
	results, err := h.Service.SearchListings(query)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, "Search results", results, nil)
}
