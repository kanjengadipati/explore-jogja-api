package bonus

import (
	"strconv"

	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

// ListMyBonuses — GET /bonuses/me (permission: bonus:read_own)
func (h *Handler) ListMyBonuses(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, 401, "Unauthorized")
		return
	}

	pagination := httpx.ParsePagination(c)
	items, total, err := h.Service.ListMyBonuses(userID, pagination.Limit, pagination.Offset)
	if err != nil {
		httpx.Error(c, 500, "Failed to list bonuses")
		return
	}

	meta := httpx.BuildPaginationMeta(total, pagination.Page(), pagination.Limit)
	httpx.Success(c, 200, "OK", items, meta)
}

// ListAllBonuses — GET /bonuses (permission: bonus:read_all)
func (h *Handler) ListAllBonuses(c *gin.Context) {
	pagination := httpx.ParsePagination(c)
	items, total, err := h.Service.ListAllBonuses(pagination.Limit, pagination.Offset)
	if err != nil {
		httpx.Error(c, 500, "Failed to list bonuses")
		return
	}

	meta := httpx.BuildPaginationMeta(total, pagination.Page(), pagination.Limit)
	httpx.Success(c, 200, "OK", items, meta)
}

// ListBonusRules — GET /admin/bonus-rules (permission: bonus:read_all)
func (h *Handler) ListBonusRules(c *gin.Context) {
	rules, err := h.Service.ListBonusRules()
	if err != nil {
		httpx.Error(c, 500, "Failed to list bonus rules")
		return
	}
	httpx.Success(c, 200, "OK", rules, nil)
}

// CreateBonusRule — POST /admin/bonus-rules (permission: bonus:manage_rules)
func (h *Handler) CreateBonusRule(c *gin.Context) {
	var req CreateBonusRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, 400, "Invalid request body")
		return
	}
	rule, err := h.Service.CreateBonusRule(req)
	if err != nil {
		httpx.Error(c, 400, err.Error())
		return
	}
	httpx.Success(c, 201, "Bonus rule created", toBonusRuleResponse(rule), nil)
}

// UpdateBonusRule — PUT /admin/bonus-rules/:id (permission: bonus:manage_rules)
func (h *Handler) UpdateBonusRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Error(c, 400, "Invalid bonus rule id")
		return
	}
	var req CreateBonusRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, 400, "Invalid request body")
		return
	}
	rule, err := h.Service.UpdateBonusRule(uint(id), req)
	if err != nil {
		httpx.Error(c, 400, err.Error())
		return
	}
	httpx.Success(c, 200, "Bonus rule updated", toBonusRuleResponse(rule), nil)
}

// UpdateBonusStatus — PUT /admin/bonuses/:id/status (permission: bonus:manage_payout)
func (h *Handler) UpdateBonusStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Error(c, 400, "Invalid bonus id")
		return
	}
	var req UpdateBonusStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, 400, "Invalid request body — status must be 'paid' or 'voided'")
		return
	}
	bonus, err := h.Service.MarkStatus(uint(id), req.Status)
	if err != nil {
		httpx.Error(c, 400, err.Error())
		return
	}
	httpx.Success(c, 200, "Bonus status updated", bonus, nil)
}

// DeleteBonusRule — DELETE /admin/bonus-rules/:id (permission: bonus:manage_rules)
func (h *Handler) DeleteBonusRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Error(c, 400, "Invalid bonus rule id")
		return
	}
	if err := h.Service.DeleteBonusRule(uint(id)); err != nil {
		httpx.Error(c, 404, err.Error())
		return
	}
	httpx.Success(c, 200, "Bonus rule deleted", nil, nil)
}
