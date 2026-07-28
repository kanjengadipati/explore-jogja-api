package partner

import (
	"pleco-api/internal/httpx"
	"pleco-api/internal/middleware"
	"pleco-api/internal/modules/promotion"
	"pleco-api/internal/modules/review"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service       *Service
	PromoService  *promotion.Service
	ReviewService *review.Service
	PermissionSvc middleware.PermissionChecker
}

func NewHandler(service *Service, promoSvc *promotion.Service, reviewSvc *review.Service) *Handler {
	return &Handler{Service: service, PromoService: promoSvc, ReviewService: reviewSvc}
}

// --- ownerLookup verifies the :id param belongs to the authenticated user ---

func (h *Handler) ownerLookup(c *gin.Context) (*Partner, bool) {
	userID, _ := httpx.GetUserIDFromContext(c)
	externalID := c.Param("id")
	partner, err := h.Service.GetOwnedByID(userID, externalID)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Listing not found")
		return nil, false
	}
	return partner, true
}

// --- Public ---

func (h *Handler) GetAll(c *gin.Context) {
	partners, err := h.Service.GetAllApproved()
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Partners fetched", ToPublicPartnerResponses(partners), nil)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	partner, err := h.Service.GetByID(id)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Partner not found")
		return
	}
	httpx.Success(c, 200, "Partner fetched", ToPublicPartnerResponse(*partner), nil)
}

func (h *Handler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Query parameter 'q' is required")
		return
	}
	partners, err := h.Service.Search(query)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Search results", ToPublicPartnerResponses(partners), nil)
}

func (h *Handler) GetSponsored(c *gin.Context) {
	destinationID := c.Query("destination_id")
	category := c.Query("category")

	partners, err := h.Service.GetSponsored(destinationID, category)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Sponsored partners fetched", ToPublicPartnerResponses(partners), nil)
}

func (h *Handler) TrackImpression(c *gin.Context) {
	id := c.Param("id")
	_ = h.Service.TrackImpression(id)
	httpx.Success(c, 200, "Impression tracked", nil, nil)
}

func (h *Handler) TrackClick(c *gin.Context) {
	id := c.Param("id")
	_ = h.Service.TrackClick(id)
	httpx.Success(c, 200, "Click tracked", nil, nil)
}

// --- Admin manage-all ---

func (h *Handler) AdminGetAll(c *gin.Context) {
	partners, err := h.Service.GetAllAny()
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Partners fetched", partners, nil)
}

func (h *Handler) AdminCreate(c *gin.Context) {
	var partner Partner
	if err := c.ShouldBindJSON(&partner); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}
	partner.Status = StatusApproved
	if err := h.Service.Create(&partner); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 201, "Partner created", partner, nil)
}

func (h *Handler) AdminUpdate(c *gin.Context) {
	id := c.Param("id")

	var req UpdatePartnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}

	partner, err := h.Service.Update(id, req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Partner updated", partner, nil)
}

func (h *Handler) AdminDelete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.Delete(id); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Partner deleted", nil, nil)
}

// --- Self-service partner ---

func (h *Handler) Apply(c *gin.Context) {
	userID, _ := httpx.GetUserIDFromContext(c)

	var req ApplyPartnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}

	partner, err := h.Service.Apply(req, userID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 201, "Partner application submitted", partner, nil)
}

func (h *Handler) GetMyListings(c *gin.Context) {
	userID, _ := httpx.GetUserIDFromContext(c)
	partners, err := h.Service.GetOwned(userID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "My listings fetched", partners, nil)
}

func (h *Handler) GetMyListing(c *gin.Context) {
	userID, _ := httpx.GetUserIDFromContext(c)
	externalID := c.Param("id")

	partner, err := h.Service.GetOwnedByID(userID, externalID)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Listing not found")
		return
	}
	httpx.Success(c, 200, "Listing fetched", partner, nil)
}

func (h *Handler) UpdateMyListing(c *gin.Context) {
	userID, _ := httpx.GetUserIDFromContext(c)
	externalID := c.Param("id")

	var req UpdatePartnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}

	// Strip fields partner cannot set themselves
	req.Rating = nil
	req.Distance = nil
	req.IsSponsored = nil
	req.SponsorTier = nil
	req.SponsorStartAt = nil
	req.SponsorEndAt = nil
	req.TargetDestIDs = nil
	req.SponsorPrice = nil
	req.SponsorPriceCurrency = nil
	req.SponsorPaymentStatus = nil

	updated, err := h.Service.UpdateOwned(userID, externalID, req)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Listing not found")
		return
	}
	httpx.Success(c, 200, "Listing updated", updated, nil)
}

func (h *Handler) DeleteMyListing(c *gin.Context) {
	userID, _ := httpx.GetUserIDFromContext(c)
	externalID := c.Param("id")

	if err := h.Service.DeleteOwned(userID, externalID); err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Listing not found")
		return
	}
	httpx.Success(c, 200, "Listing deleted", nil, nil)
}

// --- Admin approval workflow ---

type reviewDecisionRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) AdminGetPending(c *gin.Context) {
	partners, err := h.Service.GetByStatus(StatusPending)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Pending partners fetched", partners, nil)
}

func (h *Handler) AdminGetByID(c *gin.Context) {
	id := c.Param("id")
	partner, err := h.Service.GetByIDAny(id)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Partner not found")
		return
	}
	httpx.Success(c, 200, "Partner fetched", partner, nil)
}

func (h *Handler) AdminApprove(c *gin.Context) {
	id := c.Param("id")
	adminUserID, _ := httpx.GetUserIDFromContext(c)
	h.reviewDecision(id, StatusApproved, "", adminUserID, c)
}

func (h *Handler) AdminReject(c *gin.Context) {
	id := c.Param("id")
	adminUserID, _ := httpx.GetUserIDFromContext(c)

	var body reviewDecisionRequest
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Reason) == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "reason is required")
		return
	}

	h.reviewDecision(id, StatusRejected, body.Reason, adminUserID, c)
}

func (h *Handler) AdminSuspend(c *gin.Context) {
	id := c.Param("id")
	adminUserID, _ := httpx.GetUserIDFromContext(c)

	var body reviewDecisionRequest
	_ = c.ShouldBindJSON(&body)

	h.reviewDecision(id, StatusSuspended, body.Reason, adminUserID, c)
}

func (h *Handler) reviewDecision(id, status, reason string, adminUserID uint, c *gin.Context) {
	partner, err := h.Service.GetByIDAny(id)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Partner not found")
		return
	}

	now := time.Now()
	partner.Status = status
	partner.RejectionReason = reason
	partner.ReviewedAt = &now
	partner.ReviewedBy = &adminUserID

	if err := h.Service.Save(partner); err != nil {
		httpx.HandleError(c, err)
		return
	}

	httpx.Success(c, 200, "Partner status updated", partner, nil)
}

// --- Partner promotions ---

func (h *Handler) ListMyPromotions(c *gin.Context) {
	partner, ok := h.ownerLookup(c)
	if !ok {
		return
	}

	promos, err := h.PromoService.GetByPartnerID(partner.ExternalID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Promotions fetched", promos, nil)
}

func (h *Handler) CreateMyPromotion(c *gin.Context) {
	partner, ok := h.ownerLookup(c)
	if !ok {
		return
	}

	var promo promotion.Promotion
	if err := c.ShouldBindJSON(&promo); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}
	partnerID := partner.ExternalID
	promo.PartnerID = &partnerID

	if err := h.PromoService.Create(&promo); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 201, "Promotion created", promo, nil)
}

func (h *Handler) UpdateMyPromotion(c *gin.Context) {
	partner, ok := h.ownerLookup(c)
	if !ok {
		return
	}

	pid := c.Param("pid")
	promo, err := h.PromoService.GetByIDAndPartner(pid, partner.ExternalID)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Promotion not found")
		return
	}

	var req promotion.UpdatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}

	// Strip partner_id — cannot reassign
	req.PartnerID = nil

	updated, err := h.PromoService.Update(promo.ExternalID, req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Promotion updated", updated, nil)
}

func (h *Handler) DeleteMyPromotion(c *gin.Context) {
	partner, ok := h.ownerLookup(c)
	if !ok {
		return
	}

	pid := c.Param("pid")
	promo, err := h.PromoService.GetByIDAndPartner(pid, partner.ExternalID)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Promotion not found")
		return
	}

	if err := h.PromoService.Delete(promo.ExternalID); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Promotion deleted", nil, nil)
}

// --- Partner reviews ---

func (h *Handler) ListMyReviews(c *gin.Context) {
	partner, ok := h.ownerLookup(c)
	if !ok {
		return
	}

	reviews, err := h.ReviewService.GetByPartnerID(partner.ExternalID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Reviews fetched", reviews, nil)
}

func (h *Handler) ReplyToReview(c *gin.Context) {
	partner, ok := h.ownerLookup(c)
	if !ok {
		return
	}

	rid := c.Param("rid")
	reviewObj, err := h.ReviewService.GetByID(rid)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Review not found")
		return
	}
	if reviewObj.PartnerID == nil || *reviewObj.PartnerID != partner.ExternalID {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Review not found")
		return
	}

	var body struct {
		Reply string `json:"reply" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Reply) == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "reply is required")
		return
	}

	adminUserID, _ := httpx.GetUserIDFromContext(c)
	now := time.Now()
	reviewObj.Reply = body.Reply
	reviewObj.RepliedAt = &now
	reviewObj.RepliedBy = &adminUserID

	if err := h.ReviewService.Repo.Update(reviewObj); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Reply submitted", reviewObj, nil)
}
