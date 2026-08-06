package adcampaign

import (
	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) GetAll(c *gin.Context) {
	campaigns, err := h.Service.GetAll()
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Ad campaigns fetched", campaigns, nil)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	campaign, err := h.Service.GetByID(id)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Ad campaign not found")
		return
	}
	httpx.Success(c, 200, "Ad campaign fetched", campaign, nil)
}

func (h *Handler) GetBanner(c *gin.Context) {
	placement := c.Query("placement")
	if placement == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Query parameter 'placement' is required")
		return
	}

	campaign, err := h.Service.GetActiveBanner(placement, c.Query("category"))
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Banner fetched", campaign, nil)
}

func (h *Handler) GetHouseAd(c *gin.Context) {
	placement := c.Query("placement")
	if placement == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Query parameter 'placement' is required")
		return
	}
	houseAd, err := h.Service.GetEnabledHouseAd(placement)
	if err != nil {
		httpx.Success(c, 200, "House ad fetched", nil, nil)
		return
	}
	httpx.Success(c, 200, "House ad fetched", houseAd, nil)
}

func (h *Handler) Create(c *gin.Context) {
	var campaign AdCampaign
	if err := c.ShouldBindJSON(&campaign); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}
	if campaign.BusinessExternalID == nil || *campaign.BusinessExternalID == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "business_external_id is required")
		return
	}
	if err := h.Service.Create(&campaign); err != nil {
		if err.Error() == "business not found" {
			httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "business not found")
			return
		}
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 201, "Ad campaign created", campaign, nil)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateAdCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}

	campaign, err := h.Service.Update(id, req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Ad campaign updated", campaign, nil)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.Delete(id); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Ad campaign deleted", nil, nil)
}

func (h *Handler) Approve(c *gin.Context) {
	id := c.Param("id")
	campaign, err := h.Service.Approve(id, actorUserID(c))
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Ad campaign approved", campaign, nil)
}

type rejectRequestBody struct {
	Reason string `json:"reason"`
}

func (h *Handler) Reject(c *gin.Context) {
	id := c.Param("id")

	var req rejectRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}

	campaign, err := h.Service.Reject(id, req.Reason, actorUserID(c))
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Ad campaign rejected", campaign, nil)
}

// actorUserID extracts the authenticated user ID from the JWT context set by
// AuthMiddleware. Returns 0 when unavailable so the audit columns stay empty
// rather than recording a bogus actor.
func actorUserID(c *gin.Context) uint {
	if v, exists := c.Get("user_id"); exists {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
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

func (h *Handler) GetAllHouseAds(c *gin.Context) {
	houseAds, err := h.Service.GetAllHouseAds()
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "House ads fetched", houseAds, nil)
}

func (h *Handler) CreateHouseAd(c *gin.Context) {
	var houseAd HouseAd
	if err := c.ShouldBindJSON(&houseAd); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}
	if err := h.Service.CreateHouseAd(&houseAd); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 201, "House ad created", houseAd, nil)
}

func (h *Handler) UpdateHouseAd(c *gin.Context) {
	id := c.Param("id")
	var req UpdateHouseAdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}
	houseAd, err := h.Service.UpdateHouseAd(id, req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "House ad updated", houseAd, nil)
}

func (h *Handler) DeleteHouseAd(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.DeleteHouseAd(id); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "House ad deleted", nil, nil)
}
