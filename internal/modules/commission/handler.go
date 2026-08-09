package commission

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

// ListMyCommissions — GET /sales/me/commissions (permission: commission:read_own)
func (h *Handler) ListMyCommissions(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, 401, "Unauthorized")
		return
	}

	pagination := httpx.ParsePagination(c)
	items, total, err := h.Service.ListMyCommissions(userID, pagination.Limit, pagination.Offset)
	if err != nil {
		httpx.Error(c, 500, "Failed to list commissions")
		return
	}

	meta := httpx.BuildPaginationMeta(total, pagination.Page(), pagination.Limit)
	httpx.Success(c, 200, "OK", items, meta)
}

// GetSalesPerformanceReport — GET /admin/sales-performance (permission: commission:read_all)
func (h *Handler) GetSalesPerformanceReport(c *gin.Context) {
	report, err := h.Service.GetSalesPerformanceReport()
	if err != nil {
		httpx.Error(c, 500, "Failed to build sales performance report")
		return
	}
	httpx.Success(c, 200, "OK", report, nil)
}

// GetCommissionRate — GET /admin/sales-commission-rate (permission: commission:read_all)
func (h *Handler) GetCommissionRate(c *gin.Context) {
	tier1, tier2 := h.Service.GetCommissionRates()
	httpx.Success(c, 200, "OK", gin.H{
		"tier1_rate":            tier1,
		"tier2_rate":            tier2,
		"tier_threshold_months": TierThresholdMonths,
	}, nil)
}

// UpdateCommissionRate — PUT /admin/sales-commission-rate (permission: commission:manage_rate)
func (h *Handler) UpdateCommissionRate(c *gin.Context) {
	var req UpdateCommissionRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, 400, "Invalid request body — tier1_rate and tier2_rate must be fractions between 0 and 1, e.g. 0.20")
		return
	}
	if err := h.Service.SetCommissionRates(req.Tier1Rate, req.Tier2Rate); err != nil {
		httpx.Error(c, 500, "Failed to update commission rates")
		return
	}
	httpx.Success(c, 200, "Commission rates updated", gin.H{
		"tier1_rate": req.Tier1Rate,
		"tier2_rate": req.Tier2Rate,
	}, nil)
}
