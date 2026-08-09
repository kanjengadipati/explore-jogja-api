package user

import (
	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
)

// GetMyReferralCode returns the caller's sales referral code, creating one on
// first use. Only reachable by users with Role == "sales" (enforced by the
// "sales:manage-referral" permission at the route level, and again here).
func (h *Handler) GetMyReferralCode(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, 401, "Unauthorized")
		return
	}

	code, err := h.UserService.GetOrCreateReferralCode(userID)
	if err != nil {
		httpx.Error(c, 403, err.Error())
		return
	}

	httpx.Success(c, 200, "OK", gin.H{"referral_code": code}, nil)
}

// RegenerateMyReferralCode issues a new referral code, invalidating the old
// one for future signups.
func (h *Handler) RegenerateMyReferralCode(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, 401, "Unauthorized")
		return
	}

	code, err := h.UserService.RegenerateReferralCode(userID)
	if err != nil {
		httpx.Error(c, 403, err.Error())
		return
	}

	httpx.Success(c, 200, "Kode referral baru berhasil dibuat", gin.H{"referral_code": code}, nil)
}
