package business

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"pleco-api/internal/httpx"
	"pleco-api/internal/modules/adcampaign"
	"pleco-api/internal/modules/audit"
	"pleco-api/internal/modules/notification"
	"pleco-api/internal/modules/promotion"
	"pleco-api/internal/modules/review"
	"pleco-api/internal/modules/subscription"
	"pleco-api/internal/modules/user"
	"pleco-api/internal/services"
	"pleco-api/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	Service       *Service
	PromoService  *promotion.Service
	ReviewService *review.Service
	AuditSvc      *audit.Service
	NotifSvc      *notification.Service
	SubSvc        *subscription.Service
	AdCampaignSvc *adcampaign.Service
	UserRepo      user.Repository
	EmailSvc      services.EmailService
}

func NewHandler(service *Service, promoSvc *promotion.Service, reviewSvc *review.Service, auditSvc *audit.Service, notifSvc *notification.Service, subSvc *subscription.Service, adCampaignSvc *adcampaign.Service, userRepo user.Repository, emailSvc services.EmailService) *Handler {
	return &Handler{
		Service:       service,
		PromoService:  promoSvc,
		ReviewService: reviewSvc,
		AuditSvc:      auditSvc,
		NotifSvc:      notifSvc,
		SubSvc:        subSvc,
		AdCampaignSvc: adCampaignSvc,
		UserRepo:      userRepo,
		EmailSvc:      emailSvc,
	}
}

// ownerLookup verifies the :id param belongs to the authenticated user.
func (h *Handler) ownerLookup(c *gin.Context) (*Business, bool) {
	userID, _ := httpx.GetUserIDFromContext(c)
	b, err := h.Service.GetOwnedByID(userID, c.Param("id"))
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Business not found")
		return nil, false
	}
	return b, true
}

// requireOwnerRole resolves the business via ownerLookup, then insists the
// caller holds the owner role (admins cannot manage members).
func (h *Handler) requireOwnerRole(c *gin.Context) (*Business, bool) {
	b, ok := h.ownerLookup(c)
	if !ok {
		return nil, false
	}
	userID, _ := httpx.GetUserIDFromContext(c)
	isOwner, err := h.Service.Repo.IsOwnerRole(b.ID, userID)
	if err != nil {
		httpx.ErrorWithCode(c, 500, "INTERNAL", "Failed to check role")
		return nil, false
	}
	if !isOwner {
		httpx.ErrorWithCode(c, 403, "FORBIDDEN", "Only the business owner can manage members")
		return nil, false
	}
	return b, true
}

// --- Self-service ---

func (h *Handler) CreateMyBusiness(c *gin.Context) {
	var req CreateBusinessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Nama bisnis wajib diisi")
		return
	}
	if strings.TrimSpace(req.Category) == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Kategori bisnis wajib diisi")
		return
	}
	if strings.TrimSpace(req.Phone) == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "No. Telepon/WhatsApp wajib diisi")
		return
	}
	if strings.TrimSpace(req.Address) == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Alamat usaha/kantor wajib diisi")
		return
	}
	if len(req.Regions) == 0 {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Pilih minimal 1 wilayah layanan")
		return
	}
	for _, region := range req.Regions {
		if !IsValidServiceAreaRegion(region) {
			httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Wilayah tidak dikenal: "+region)
			return
		}
	}
	userID, _ := httpx.GetUserIDFromContext(c)
	b, err := h.Service.CreateOwned(userID, req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 201, "Business created", b, nil)
}

// CheckNameSimilar returns approved businesses whose names are similar to the
// given query. Used by the frontend name-dedup step at registration time.
// Returns an empty list (not an error) for short queries (<3 chars).
func (h *Handler) CheckNameSimilar(c *gin.Context) {
	q := c.Query("q")
	similar, _ := h.Service.FindSimilarApprovedName(q) // fail-open
	if similar == nil {
		similar = []Business{}
	}
	httpx.Success(c, 200, "Name check done", similar, nil)
}

func (h *Handler) GetMyBusinesses(c *gin.Context) {
	userID, _ := httpx.GetUserIDFromContext(c)
	bs, err := h.Service.GetOwned(userID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	if bs == nil {
		bs = make([]Business, 0)
	}
	httpx.Success(c, 200, "My businesses fetched", bs, nil)
}

func (h *Handler) GetMyBusiness(c *gin.Context) {
	b, ok := h.ownerLookup(c)
	if !ok {
		return
	}
	httpx.Success(c, 200, "Business fetched", b, nil)
}

func (h *Handler) UpdateMyBusiness(c *gin.Context) {
	var req UpdateBusinessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}
	userID, _ := httpx.GetUserIDFromContext(c)
	b, err := h.Service.UpdateOwned(userID, c.Param("id"), req)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Business not found")
		return
	}
	httpx.Success(c, 200, "Business updated", b, nil)
}

func (h *Handler) GetMyListings(c *gin.Context) {
	b, ok := h.ownerLookup(c)
	if !ok {
		return
	}
	listings, err := h.Service.GetListings(b.ID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Listings fetched", listings, nil)
}

// --- Business team members (owner & admin roles) ---

// teamMember is the member projection returned to the dashboard.
type teamMember struct {
	UserID        uint   `json:"user_id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	Role          string `json:"role"`
	InvitedBy     *uint  `json:"invited_by,omitempty"`
	IsCurrentUser bool   `json:"is_current_user"`
}

func (h *Handler) GetMyMembers(c *gin.Context) {
	b, ok := h.ownerLookup(c)
	if !ok {
		return
	}
	owners, err := h.Service.Repo.ListOwners(b.ID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	currentUserID, _ := httpx.GetUserIDFromContext(c)
	members := make([]teamMember, 0, len(owners))
	for _, o := range owners {
		u, err := h.UserRepo.FindByID(o.UserID)
		if err != nil {
			continue
		}
		members = append(members, teamMember{
			UserID:        o.UserID,
			Name:          u.Name,
			Email:         u.Email,
			AvatarURL:     u.AvatarURL,
			Role:          o.Role,
			InvitedBy:     o.InvitedBy,
			IsCurrentUser: o.UserID == currentUserID,
		})
	}
	httpx.Success(c, 200, "Members fetched", members, nil)
}

type inviteMemberRequest struct {
	Email string `json:"email" binding:"required"`
	Role  string `json:"role" binding:"required,oneof=owner admin"`
}

func (h *Handler) InviteMember(c *gin.Context) {
	b, ok := h.requireOwnerRole(c)
	if !ok {
		return
	}
	var req inviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ValidationError(c, err)
		return
	}
	inviterID, _ := httpx.GetUserIDFromContext(c)
	email := strings.ToLower(strings.TrimSpace(req.Email))

	target, err := h.UserRepo.FindByEmail(email)
	switch {
	case err == nil:
		// Registered user: add them directly to the team.
		if target.ID == inviterID {
			httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "You are already a member of this business")
			return
		}
		if err := h.Service.Repo.UpsertOwner(b.ID, target.ID, req.Role); err != nil {
			httpx.HandleError(c, err)
			return
		}
		if err := h.Service.Repo.SetInvitedBy(b.ID, target.ID, inviterID); err != nil {
			httpx.HandleError(c, err)
			return
		}
		_ = h.NotifSvc.Notify(target.ID, "Akses Bisnis", "Anda telah ditambahkan sebagai "+req.Role+" di bisnis '"+b.Name+"'", "application")
		httpx.Success(c, 201, "Member invited", gin.H{"type": "existing", "user_id": target.ID, "role": req.Role}, nil)
		return
	case !errors.Is(err, gorm.ErrRecordNotFound):
		httpx.HandleError(c, err)
		return
	}

	// Unknown email: create a pending invitation the recipient accepts after
	// registering/logging in. A still-pending invite for the same business +
	// email is reissued (new token, refreshed expiry) instead of stacked.
	invite, findErr := h.Service.Repo.FindPendingInviteByBusinessAndEmail(b.ID, email)
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		httpx.HandleError(c, findErr)
		return
	}
	if findErr != nil {
		invite = &BusinessMemberInvite{
			BusinessID: b.ID,
			Email:      email,
			Role:       req.Role,
			InvitedBy:  &inviterID,
			Status:     InviteStatusPending,
			ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		}
		token, genErr := newInviteToken()
		if genErr != nil {
			httpx.HandleError(c, genErr)
			return
		}
		invite.TokenHash = utils.HashToken(token)
		if createErr := h.Service.Repo.CreateInvite(invite); createErr != nil {
			httpx.HandleError(c, createErr)
			return
		}
		// Best-effort email; the inviter can always copy the link from the UI.
		inviteURL := h.EmailSvc.BusinessInviteURL(token)
		if sendErr := h.EmailSvc.SendBusinessInvite(email, inviteURL); sendErr != nil {
			_ = sendErr
		}
		httpx.Success(c, 201, "Invitation created", gin.H{
			"type":       "invite",
			"token":      token,
			"invite_url": inviteURL,
			"email":      email,
			"role":       invite.Role,
			"expires_at": invite.ExpiresAt,
		}, nil)
		return
	}

	// Live pending invite: reissue with a fresh token and refreshed expiry.
	invite.Role = req.Role
	invite.InvitedBy = &inviterID
	invite.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	token, genErr := newInviteToken()
	if genErr != nil {
		httpx.HandleError(c, genErr)
		return
	}
	invite.TokenHash = utils.HashToken(token)
	if updateErr := h.Service.Repo.UpdateInvite(invite); updateErr != nil {
		httpx.HandleError(c, updateErr)
		return
	}
	inviteURL := h.EmailSvc.BusinessInviteURL(token)
	if sendErr := h.EmailSvc.SendBusinessInvite(email, inviteURL); sendErr != nil {
		_ = sendErr
	}
	httpx.Success(c, 200, "Invitation already exists", gin.H{
		"type":       "invite",
		"token":      token,
		"invite_url": inviteURL,
		"email":      email,
		"role":       invite.Role,
		"expires_at": invite.ExpiresAt,
	}, nil)
}

// invitePreview is the safe projection of an invitation for the accept page.
type invitePreview struct {
	Email         string    `json:"email"`
	Role          string    `json:"role"`
	Status        string    `json:"status"`
	ExpiresAt     time.Time `json:"expires_at"`
	BusinessID    string    `json:"business_external_id"`
	BusinessName  string    `json:"business_name"`
	InvitedByName string    `json:"invited_by_name"`
}

func (h *Handler) GetInvitePreview(c *gin.Context) {
	token := c.Param("token")
	invite, err := h.Service.Repo.FindInviteByTokenHash(utils.HashToken(token))
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Undangan tidak ditemukan atau sudah tidak berlaku")
		return
	}
	if invite.Status == InviteStatusAccepted {
		httpx.ErrorWithCode(c, 410, "ALREADY_ACCEPTED", "Undangan sudah diterima")
		return
	}
	if invite.Status == InviteStatusRevoked {
		httpx.ErrorWithCode(c, 410, "REVOKED", "Undangan sudah dibatalkan oleh pemilik bisnis")
		return
	}
	if invite.IsExpired() {
		httpx.ErrorWithCode(c, 410, "EXPIRED", "Undangan sudah kedaluwarsa")
		return
	}

	b, err := h.Service.Repo.FindByID(fmt.Sprintf("%d", invite.BusinessID))
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Bisnis tidak ditemukan")
		return
	}
	preview := invitePreview{
		Email:        invite.Email,
		Role:         invite.Role,
		Status:       invite.Status,
		ExpiresAt:    invite.ExpiresAt,
		BusinessID:   b.ExternalID,
		BusinessName: b.Name,
	}
	if invite.InvitedBy != nil {
		if inviter, uErr := h.UserRepo.FindByID(*invite.InvitedBy); uErr == nil {
			preview.InvitedByName = inviter.Name
		}
	}
	httpx.Success(c, 200, "Invitation found", preview, nil)
}

func (h *Handler) AcceptInvite(c *gin.Context) {
	token := c.Param("token")
	userID, _ := httpx.GetUserIDFromContext(c)

	invite, err := h.Service.Repo.FindInviteByTokenHash(utils.HashToken(token))
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Undangan tidak ditemukan atau sudah tidak berlaku")
		return
	}
	if invite.Status == InviteStatusAccepted {
		httpx.ErrorWithCode(c, 410, "ALREADY_ACCEPTED", "Undangan sudah diterima")
		return
	}
	if invite.Status == InviteStatusRevoked {
		httpx.ErrorWithCode(c, 410, "REVOKED", "Undangan sudah dibatalkan oleh pemilik bisnis")
		return
	}
	if invite.IsExpired() {
		httpx.ErrorWithCode(c, 410, "EXPIRED", "Undangan sudah kedaluwarsa")
		return
	}

	user, err := h.UserRepo.FindByID(userID)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Akun tidak ditemukan")
		return
	}
	if !strings.EqualFold(user.Email, invite.Email) {
		httpx.ErrorWithCode(c, 403, "EMAIL_MISMATCH", "Undangan ini ditujukan untuk "+invite.Email+". Silakan masuk dengan akun tersebut.")
		return
	}

	b, err := h.Service.Repo.FindByID(fmt.Sprintf("%d", invite.BusinessID))
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Bisnis tidak ditemukan")
		return
	}

	if err := h.Service.Repo.UpsertOwner(b.ID, userID, invite.Role); err != nil {
		httpx.HandleError(c, err)
		return
	}
	now := time.Now()
	invite.Status = InviteStatusAccepted
	invite.AcceptedAt = &now
	invite.AcceptedBy = &userID
	if err := h.Service.Repo.UpdateInvite(invite); err != nil {
		httpx.HandleError(c, err)
		return
	}

	_ = h.NotifSvc.Notify(userID, "Akses Bisnis", "Selamat! Anda sekarang menjadi "+invite.Role+" di bisnis '"+b.Name+"'", "application")
	if invite.InvitedBy != nil {
		_ = h.NotifSvc.Notify(*invite.InvitedBy, "Akses Bisnis", user.Name+" telah menerima undangan Anda untuk bisnis '"+b.Name+"'", "application")
	}

	httpx.Success(c, 200, "Invitation accepted", gin.H{
		"business_external_id": b.ExternalID,
		"business_name":        b.Name,
		"role":                 invite.Role,
	}, nil)
}

type pendingInviteResponse struct {
	ID            uint      `json:"id"`
	Email         string    `json:"email"`
	Role          string    `json:"role"`
	Status        string    `json:"status"`
	ExpiresAt     time.Time `json:"expires_at"`
	InvitedBy     *uint     `json:"invited_by,omitempty"`
	InvitedByName string    `json:"invited_by_name,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

func (h *Handler) ListPendingInvites(c *gin.Context) {
	b, ok := h.requireOwnerRole(c)
	if !ok {
		return
	}
	invites, err := h.Service.Repo.ListPendingInvites(b.ID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	out := make([]pendingInviteResponse, 0, len(invites))
	for _, inv := range invites {
		item := pendingInviteResponse{
			ID:        inv.ID,
			Email:     inv.Email,
			Role:      inv.Role,
			Status:    inv.Status,
			ExpiresAt: inv.ExpiresAt,
			InvitedBy: inv.InvitedBy,
			CreatedAt: inv.CreatedAt,
		}
		if inv.InvitedBy != nil {
			if inviter, uErr := h.UserRepo.FindByID(*inv.InvitedBy); uErr == nil {
				item.InvitedByName = inviter.Name
			}
		}
		out = append(out, item)
	}
	httpx.Success(c, 200, "Invitations fetched", out, nil)
}

func (h *Handler) RevokeInvite(c *gin.Context) {
	b, ok := h.requireOwnerRole(c)
	if !ok {
		return
	}
	var inviteID uint
	if _, err := fmt.Sscan(c.Param("inviteId"), &inviteID); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid invite id")
		return
	}
	invite, err := h.Service.Repo.FindInviteByID(inviteID)
	if err != nil || invite.BusinessID != b.ID {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Invitation not found")
		return
	}
	if invite.Status != InviteStatusPending {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invitation is not pending")
		return
	}
	invite.Status = InviteStatusRevoked
	if err := h.Service.Repo.UpdateInvite(invite); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Invitation revoked", gin.H{"id": invite.ID, "status": invite.Status}, nil)
}

// newInviteToken generates a cryptographically random invite token (48 hex chars).
func newInviteToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (h *Handler) RemoveMember(c *gin.Context) {
	b, ok := h.requireOwnerRole(c)
	if !ok {
		return
	}
	actorID, _ := httpx.GetUserIDFromContext(c)
	var userID uint
	if _, err := fmt.Sscan(c.Param("userId"), &userID); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid member id")
		return
	}
	if userID == actorID {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "You cannot remove yourself")
		return
	}
	role, err := h.Service.Repo.GetRole(b.ID, userID)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Member not found")
		return
	}
	if role == "" {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Member not found")
		return
	}
	if role == RoleOwner {
		ownerIDs, _ := h.Service.Repo.FindOwnerUserIDs(b.ID)
		ownerCount := 0
		for _, id := range ownerIDs {
			r, _ := h.Service.Repo.GetRole(b.ID, id)
			if r == RoleOwner {
				ownerCount++
			}
		}
		if ownerCount <= 1 {
			httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "At least one owner must remain")
			return
		}
	}
	if err := h.Service.Repo.RemoveOwner(b.ID, userID); err != nil {
		httpx.HandleError(c, err)
		return
	}
	_ = h.NotifSvc.Notify(userID, "Akses Bisnis", "Anda telah dihapus dari bisnis '"+b.Name+"'", "application")
	httpx.Success(c, 200, "Member removed", nil, nil)
}

// --- Admin approval workflow ---

func (h *Handler) AdminGetAll(c *gin.Context) {
	bs, err := h.Service.GetAllAny()
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Businesses fetched", bs, nil)
}

func (h *Handler) AdminGetPending(c *gin.Context) {
	bs, err := h.Service.GetByStatus(StatusPending)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Pending businesses fetched", bs, nil)
}

func (h *Handler) AdminGetByID(c *gin.Context) {
	b, err := h.Service.GetByIDAny(c.Param("id"))
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Business not found")
		return
	}
	httpx.Success(c, 200, "Business fetched", b, nil)
}

type reviewDecisionRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) AdminApprove(c *gin.Context) {
	adminUserID, _ := httpx.GetUserIDFromContext(c)
	h.reviewDecision(c.Param("id"), StatusApproved, "", adminUserID, c)
}

func (h *Handler) AdminReject(c *gin.Context) {
	adminUserID, _ := httpx.GetUserIDFromContext(c)
	var body reviewDecisionRequest
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Reason) == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "reason is required")
		return
	}
	h.reviewDecision(c.Param("id"), StatusRejected, body.Reason, adminUserID, c)
}

func (h *Handler) AdminSuspend(c *gin.Context) {
	adminUserID, _ := httpx.GetUserIDFromContext(c)
	var body reviewDecisionRequest
	_ = c.ShouldBindJSON(&body)
	h.reviewDecision(c.Param("id"), StatusSuspended, body.Reason, adminUserID, c)
}

func (h *Handler) reviewDecision(id, status, reason string, adminUserID uint, c *gin.Context) {
	b, err := h.Service.SetStatus(id, status, reason, adminUserID)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Business not found")
		return
	}

	h.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &adminUserID,
		Action:      "business." + status,
		Resource:    "business",
		Status:      "success",
		Description: "Business " + b.ExternalID + " (" + b.Name + ") " + status + ". Reason: " + reason,
		IPAddress:   c.ClientIP(),
		UserAgent:   c.GetHeader("User-Agent"),
	})

	ownerIDs, _ := h.Service.Repo.FindOwnerUserIDs(b.ID)
	for _, uid := range ownerIDs {
		title := "Status Bisnis"
		content := "Status bisnis Anda '" + b.Name + "' telah diubah menjadi: " + status
		if reason != "" {
			content += ". Alasan: " + reason
		}
		_ = h.NotifSvc.Notify(uid, title, content, "application")
	}

	httpx.Success(c, 200, "Business status updated", b, nil)
}

// --- Business promotions (Phase 3 additive: linked via legacy partner id) ---

func (h *Handler) ListMyPromotions(c *gin.Context) {
	b, ok := h.ownerLookup(c)
	if !ok {
		return
	}
	promos, err := h.PromoService.GetByBusinessID(b.ExternalID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Promotions fetched", promos, nil)
}

func (h *Handler) CreateMyPromotion(c *gin.Context) {
	b, ok := h.ownerLookup(c)
	if !ok {
		return
	}
	if b.ExternalID == "" {
		httpx.ErrorWithCode(c, 400, "NOT_SUPPORTED", "This business is not linked to a legacy partner listing yet")
		return
	}

	var promo promotion.Promotion
	if err := c.ShouldBindJSON(&promo); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}
	businessID := b.ExternalID
	promo.BusinessExternalID = &businessID
	// Keep the legacy link during the transition so the old partner dashboard
	// still shows it; retired when the dual-write is removed (Phase 6).
	if b.LegacyPartnerExternalID != nil {
		legacyID := *b.LegacyPartnerExternalID
		promo.LegacyPartnerExternalID = &legacyID
	}
	promo.Status = "pending"

	if err := h.PromoService.Create(&promo); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 201, "Promotion submitted for review", promo, nil)
}

func (h *Handler) UpdateMyPromotion(c *gin.Context) {
	b, ok := h.ownerLookup(c)
	if !ok {
		return
	}

	pid := c.Param("pid")
	promo, err := h.PromoService.GetByIDAndBusiness(pid, b.ExternalID)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Promotion not found")
		return
	}

	var req promotion.UpdatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}
	// Strip partner id — cannot reassign
	req.PartnerID = nil

	updated, err := h.PromoService.Update(promo.ExternalID, req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Promotion updated", updated, nil)
}

func (h *Handler) DeleteMyPromotion(c *gin.Context) {
	b, ok := h.ownerLookup(c)
	if !ok {
		return
	}

	pid := c.Param("pid")
	promo, err := h.PromoService.GetByIDAndBusiness(pid, b.ExternalID)
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

// --- Business reviews (Phase 3 additive: linked via legacy partner id) ---

func (h *Handler) ListMyReviews(c *gin.Context) {
	b, ok := h.ownerLookup(c)
	if !ok {
		return
	}
	if b.LegacyPartnerExternalID == nil {
		httpx.Success(c, 200, "Reviews fetched", []any{}, nil)
		return
	}
	reviews, err := h.ReviewService.GetByPartnerID(*b.LegacyPartnerExternalID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Reviews fetched", reviews, nil)
}

func (h *Handler) ReplyToReview(c *gin.Context) {
	b, ok := h.ownerLookup(c)
	if !ok {
		return
	}

	rid := c.Param("rid")
	reviewObj, err := h.ReviewService.GetByID(rid)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Review not found")
		return
	}
	if reviewObj.PartnerID == nil || b.LegacyPartnerExternalID == nil || *reviewObj.PartnerID != *b.LegacyPartnerExternalID {
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

func (h *Handler) GetMySubscription(c *gin.Context) {
	b, ok := h.ownerLookup(c)
	if !ok {
		return
	}

	sub, err := h.SubSvc.GetByBusinessExternalID(b.ExternalID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Subscription fetched", sub, nil)
}

func (h *Handler) GetMyAdCampaigns(c *gin.Context) {
	b, ok := h.ownerLookup(c)
	if !ok {
		return
	}
	campaigns, err := h.AdCampaignSvc.GetAllByBusiness(b.ExternalID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Ad campaigns fetched", campaigns, nil)
}

func (h *Handler) CreateMyAdCampaign(c *gin.Context) {
	b, ok := h.ownerLookup(c)
	if !ok {
		return
	}
	var req adcampaign.SelfServiceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ValidationError(c, err)
		return
	}
	campaign, err := h.AdCampaignSvc.CreateSelfService(b.ExternalID, b.Name, req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 201, "Ad campaign created", campaign, nil)
}
