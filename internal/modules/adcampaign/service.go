package adcampaign

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"pleco-api/internal/domain"
	"pleco-api/internal/modules/subscription"
	"pleco-api/internal/services"
	"pleco-api/internal/utils"
)

type Service struct {
	Repo            Repository
	SubscriptionSvc *subscription.Service
	EmailSvc        services.EmailService
}

func NewService(repo Repository, subscriptionSvc *subscription.Service, emailSvc services.EmailService) *Service {
	return &Service{Repo: repo, SubscriptionSvc: subscriptionSvc, EmailSvc: emailSvc}
}

func (s *Service) GetAll() ([]AdCampaign, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetAllByBusiness(businessExternalID string) ([]AdCampaign, error) {
	return s.Repo.FindAllByBusinessExternalID(businessExternalID)
}

func (s *Service) GetByID(externalID string) (*AdCampaign, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) GetActiveBanner(placement, category string) (*AdCampaign, error) {
	candidates, err := s.Repo.FindActiveCandidates(placement, category)
	if err != nil {
		return nil, err
	}
	return WeightedPick(candidates), nil
}

func (s *Service) Create(campaign *AdCampaign) error {
	if campaign.BusinessExternalID == nil || *campaign.BusinessExternalID == "" {
		return errors.New("business_external_id is required")
	}
	exists, err := s.Repo.BusinessExists(*campaign.BusinessExternalID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("business not found")
	}

	canCreate, err := s.SubscriptionSvc.CanCreateAdCampaign(*campaign.BusinessExternalID)
	if err != nil {
		return err
	}
	if !canCreate {
		return errors.New("business on free plan cannot create ad campaigns")
	}

	// Approval is mandatory: every sellable campaign starts under review
	// regardless of the source (admin-created or self-service).
	if IsSellablePlacement(campaign.Placement) {
		campaign.PaymentStatus = PaymentStatusPendingReview
	}

	return s.Repo.Create(campaign)
}

// SelfServiceCreateRequest is the payload the business portal submits when a
// business owner creates their own ad campaign (no price/status fields — those
// are derived server-side from the placement). The creative image and target
// URL are required for banner-style placements; ecosystem placements derive
// both from the promoted listing instead.
type SelfServiceCreateRequest struct {
	Placement string     `json:"placement" binding:"required"`
	ImageURL  string     `json:"image_url"`
	TargetURL string     `json:"target_url"`
	Category  string     `json:"category"`
	StartAt   *time.Time `json:"start_at"`
	EndAt     *time.Time `json:"end_at"`

	// Ecosystem rail fields — required when placement is ecosystem_*.
	ListingType       string   `json:"listing_type"`
	ListingExternalID string   `json:"listing_external_id"`
	TargetDestIDs     []string `json:"target_dest_ids"`
	SortOrder         int      `json:"sort_order"`
}

// CreateSelfService creates an ad campaign on behalf of a business owner.
// The price comes from the per-placement monthly rate (§ pricing.go) and the
// initial status follows the placement's moderation policy: review-required
// slots start pending_review, the rest start pending_payment and go live
// automatically once the payment webhook marks them paid.
func (s *Service) CreateSelfService(businessExternalID, partnerName string, req SelfServiceCreateRequest) (*AdCampaign, error) {
	if !IsSellablePlacement(req.Placement) {
		return nil, domain.NewAPIError(http.StatusBadRequest, domain.CodeValidationFailed,
			"unknown or non-sellable placement", nil)
	}

	// Ecosystem placements promote a real listing owned by the business. The
	// creative image and target URL are derived from the listing record itself,
	// so the owner does not have to upload a separate creative for these slots.
	if IsEcosystemPlacement(req.Placement) {
		if strings.TrimSpace(req.ListingType) == "" || strings.TrimSpace(req.ListingExternalID) == "" {
			return nil, domain.NewAPIError(http.StatusBadRequest, domain.CodeValidationFailed,
				"listing_type and listing_external_id are required for ecosystem placements", nil)
		}
		if err := s.validateListingOwnership(businessExternalID, req.ListingType, req.ListingExternalID); err != nil {
			return nil, err
		}
		listing, err := s.Repo.FindEcosystemListing(req.ListingType, req.ListingExternalID)
		if err != nil {
			return nil, err
		}
		if listing == nil {
			return nil, domain.NewAPIError(http.StatusNotFound, domain.CodeNotFound,
				"listing not found or not owned by this business", nil)
		}
		req.ImageURL = listing.Image
		req.TargetURL = listing.Website
	} else if strings.TrimSpace(req.ImageURL) == "" || strings.TrimSpace(req.TargetURL) == "" {
		return nil, domain.NewAPIError(http.StatusBadRequest, domain.CodeValidationFailed,
			"image_url and target_url are required", nil)
	}

	exists, err := s.Repo.BusinessExists(businessExternalID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.NewAPIError(http.StatusNotFound, domain.CodeNotFound, "business not found", nil)
	}

	canCreate, err := s.SubscriptionSvc.CanCreateAdCampaign(businessExternalID)
	if err != nil {
		return nil, err
	}
	if !canCreate {
		return nil, domain.NewAPIError(http.StatusBadRequest, domain.CodeValidationFailed,
			"business on free plan cannot create ad campaigns", nil)
	}

	var startAt, endAt time.Time
	if req.StartAt != nil {
		startAt = *req.StartAt
	}
	if req.EndAt != nil {
		endAt = *req.EndAt
	}

	targetDestIDs := make(utils.JSONArr, 0, len(req.TargetDestIDs))
	for _, id := range req.TargetDestIDs {
		targetDestIDs = append(targetDestIDs, id)
	}

	campaign := &AdCampaign{
		ExternalID:         uuid.New().String(),
		BusinessExternalID: &businessExternalID,
		PartnerName:        partnerName,
		Placement:          req.Placement,
		ImageURL:           req.ImageURL,
		TargetURL:          req.TargetURL,
		Category:           req.Category,
		ListingType:        req.ListingType,
		ListingExternalID:  req.ListingExternalID,
		TargetDestIDs:      targetDestIDs,
		SortOrder:          req.SortOrder,
		StartAt:            startAt,
		EndAt:              endAt,
		Weight:             1,
		IsActive:           true,
		PriceAmount:        s.PriceForPlacement(req.Placement, startAt, endAt),
		PriceCurrency:      "IDR",
		PaymentStatus:      InitialPaymentStatus(req.Placement),
	}

	if err := s.Repo.Create(campaign); err != nil {
		return nil, err
	}
	return campaign, nil
}

// validateListingOwnership ensures a business can only sponsor listings it
// actually owns (business_id FK set by the listing-claim approval flow).
func (s *Service) validateListingOwnership(businessExternalID, listingType, listingExternalID string) error {
	ok, err := s.Repo.ListingBelongsToBusiness(listingType, listingExternalID, businessExternalID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.NewAPIError(http.StatusNotFound, domain.CodeNotFound,
			"listing not found or not owned by this business", nil)
	}
	return nil
}

type UpdateAdCampaignRequest struct {
	PartnerName        *string    `json:"partner_name"`
	BusinessExternalID *string    `json:"business_external_id"`
	Placement          *string    `json:"placement"`
	ImageURL           *string    `json:"image_url"`
	TargetURL          *string    `json:"target_url"`
	Category           *string    `json:"category"`
	ListingType        *string    `json:"listing_type"`
	ListingExternalID  *string    `json:"listing_external_id"`
	TargetDestIDs      *[]string  `json:"target_dest_ids"`
	SortOrder          *int       `json:"sort_order"`
	StartAt            *time.Time `json:"start_at"`
	EndAt              *time.Time `json:"end_at"`
	Weight             *int       `json:"weight"`
	IsActive           *bool      `json:"is_active"`

	PriceAmount   *float64 `json:"price_amount"`
	PriceCurrency *string  `json:"price_currency"`
	PaymentStatus *string  `json:"payment_status"`
}

func (s *Service) Update(externalID string, req UpdateAdCampaignRequest) (*AdCampaign, error) {
	campaign, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("ad campaign not found")
	}

	if req.PartnerName != nil {
		campaign.PartnerName = *req.PartnerName
	}
	if req.BusinessExternalID != nil {
		if *req.BusinessExternalID == "" {
			return nil, errors.New("business_external_id cannot be empty")
		}
		exists, err := s.Repo.BusinessExists(*req.BusinessExternalID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.New("business not found")
		}
		campaign.BusinessExternalID = req.BusinessExternalID
	}
	if req.Placement != nil {
		campaign.Placement = *req.Placement
	}
	if req.ImageURL != nil {
		campaign.ImageURL = *req.ImageURL
	}
	if req.TargetURL != nil {
		campaign.TargetURL = *req.TargetURL
	}
	if req.Category != nil {
		campaign.Category = *req.Category
	}
	if req.ListingType != nil {
		campaign.ListingType = *req.ListingType
	}
	if req.ListingExternalID != nil {
		campaign.ListingExternalID = *req.ListingExternalID
	}
	if req.TargetDestIDs != nil {
		ids := make(utils.JSONArr, 0, len(*req.TargetDestIDs))
		for _, id := range *req.TargetDestIDs {
			ids = append(ids, id)
		}
		campaign.TargetDestIDs = ids
	}
	if req.SortOrder != nil {
		campaign.SortOrder = *req.SortOrder
	}
	if req.StartAt != nil {
		campaign.StartAt = *req.StartAt
	}
	if req.EndAt != nil {
		campaign.EndAt = *req.EndAt
	}
	if req.Weight != nil {
		campaign.Weight = *req.Weight
	}
	if req.IsActive != nil {
		campaign.IsActive = *req.IsActive
	}
	if req.PriceAmount != nil {
		campaign.PriceAmount = *req.PriceAmount
	}
	if req.PriceCurrency != nil {
		campaign.PriceCurrency = *req.PriceCurrency
	}
	if req.PaymentStatus != nil {
		if !validPaymentTransition(campaign.PaymentStatus, *req.PaymentStatus) {
			return nil, errors.New("invalid payment status transition; use the approve/reject endpoints for review-gated campaigns")
		}
		campaign.PaymentStatus = *req.PaymentStatus
	}

	if err := s.Repo.Update(campaign); err != nil {
		return nil, err
	}
	return campaign, nil
}

// validPaymentTransition guards direct payment_status writes through the
// generic update endpoint. Approvals and rejections must go through Approve /
// Reject so the notification email is always sent; the payment webhook is the
// only path allowed to flip a payable campaign to paid.
func validPaymentTransition(current, next string) bool {
	if current == next {
		return true
	}
	switch next {
	case PaymentStatusPaid:
		return current == PaymentStatusPendingPayment || current == "pending"
	case PaymentStatusPendingPayment:
		return current == "pending"
	default:
		// pending_review, rejected and unknown statuses cannot be set directly.
		return false
	}
}

// Approve opens payment for a campaign that is pending review. It is the only
// path from pending_review → pending_payment, and it triggers the approval
// email to the campaign owner.
func (s *Service) Approve(externalID string, actorID uint) (*AdCampaign, error) {
	campaign, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("ad campaign not found")
	}
	if campaign.PaymentStatus != PaymentStatusPendingReview {
		return nil, errors.New("campaign is not pending review")
	}
	now := time.Now()
	campaign.PaymentStatus = PaymentStatusPendingPayment
	campaign.ApprovedAt = &now
	campaign.ApprovedBy = s.actorLabel(actorID)
	if err := s.Repo.Update(campaign); err != nil {
		return nil, err
	}
	s.sendApprovalEmail(campaign, true, "")
	return campaign, nil
}

// Reject denies a campaign that is pending review, recording the reason and the
// acting admin. It is the only path from pending_review → rejected and
// notifies the campaign owner.
func (s *Service) Reject(externalID, reason string, actorID uint) (*AdCampaign, error) {
	campaign, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("ad campaign not found")
	}
	if campaign.PaymentStatus != PaymentStatusPendingReview {
		return nil, errors.New("campaign is not pending review")
	}
	now := time.Now()
	campaign.PaymentStatus = PaymentStatusRejected
	campaign.RejectionReason = reason
	campaign.RejectedAt = &now
	campaign.RejectedBy = s.actorLabel(actorID)
	if err := s.Repo.Update(campaign); err != nil {
		return nil, err
	}
	s.sendApprovalEmail(campaign, false, reason)
	return campaign, nil
}

// actorLabel resolves a human-readable identity (the admin's email when
// available, otherwise "admin#<id>") for the approved_by / rejected_by audit
// columns.
func (s *Service) actorLabel(userID uint) string {
	if userID == 0 {
		return ""
	}
	if email, err := s.Repo.UserEmailByID(userID); err == nil && email != "" {
		return email
	}
	return fmt.Sprintf("admin#%d", userID)
}

// sendApprovalEmail notifies the campaign owner(s) after an approve/reject
// action. Legacy campaigns without a business reference have no reliable
// recipient, so the email is skipped and logged instead of failing the action.
func (s *Service) sendApprovalEmail(campaign *AdCampaign, approved bool, reason string) {
	if s.EmailSvc == nil {
		return
	}
	if campaign.BusinessExternalID == nil || *campaign.BusinessExternalID == "" {
		log.Printf("adcampaign: skipping %s email for %s: legacy campaign without business_external_id",
			emailActionLabel(approved), campaign.ExternalID)
		return
	}
	emails, err := s.Repo.OwnerEmailsForBusiness(*campaign.BusinessExternalID)
	if err != nil {
		log.Printf("adcampaign: failed to resolve owner emails for %s: %v", campaign.ExternalID, err)
		return
	}
	if len(emails) == 0 {
		log.Printf("adcampaign: no owner email for campaign %s; skipping %s notification",
			campaign.ExternalID, emailActionLabel(approved))
		return
	}

	name := campaign.BusinessName
	if name == "" {
		name = campaign.PartnerName
	}
	placement := PlacementName(campaign.Placement)
	for _, email := range emails {
		var err error
		if approved {
			err = s.EmailSvc.SendAdCampaignApproved(email, name, placement, campaign.StartAt, campaign.EndAt)
		} else {
			err = s.EmailSvc.SendAdCampaignRejected(email, name, placement, reason)
		}
		if err != nil {
			log.Printf("adcampaign: failed to send %s email to %s for %s: %v",
				emailActionLabel(approved), email, campaign.ExternalID, err)
		}
	}
}

func emailActionLabel(approved bool) string {
	if approved {
		return "approval"
	}
	return "rejection"
}

func (s *Service) Delete(externalID string) error {
	return s.Repo.Delete(externalID)
}

func (s *Service) TrackImpression(externalID string) error {
	return s.Repo.IncrementImpression(externalID)
}

func (s *Service) TrackClick(externalID string) error {
	return s.Repo.IncrementClick(externalID)
}

func (s *Service) GetAllHouseAds() ([]HouseAd, error) {
	return s.Repo.FindAllHouseAds()
}

func (s *Service) GetEnabledHouseAd(placement string) (*HouseAd, error) {
	return s.Repo.FindEnabledHouseAdByPlacement(placement)
}

type UpdateHouseAdRequest struct {
	Placement  *string `json:"placement"`
	Headline   *string `json:"headline"`
	HeadlineEn *string `json:"headline_en"`
	Subline    *string `json:"subline"`
	SublineEn  *string `json:"subline_en"`
	CTALabel   *string `json:"cta_label"`
	CTALabelEn *string `json:"cta_label_en"`
	ImageURL   *string `json:"image_url"`
	TargetURL  *string `json:"target_url"`
	IsEnabled  *bool   `json:"is_enabled"`
}

func (s *Service) CreateHouseAd(houseAd *HouseAd) error {
	if houseAd.ExternalID == "" {
		houseAd.ExternalID = uuid.New().String()
	}
	return s.Repo.CreateHouseAd(houseAd)
}

func (s *Service) UpdateHouseAd(externalID string, req UpdateHouseAdRequest) (*HouseAd, error) {
	houseAd, err := s.Repo.FindHouseAdByID(externalID)
	if err != nil {
		return nil, errors.New("house ad not found")
	}
	if req.Placement != nil {
		houseAd.Placement = *req.Placement
	}
	if req.Headline != nil {
		houseAd.Headline = *req.Headline
	}
	if req.HeadlineEn != nil {
		houseAd.HeadlineEn = *req.HeadlineEn
	}
	if req.Subline != nil {
		houseAd.Subline = *req.Subline
	}
	if req.SublineEn != nil {
		houseAd.SublineEn = *req.SublineEn
	}
	if req.CTALabel != nil {
		houseAd.CTALabel = *req.CTALabel
	}
	if req.CTALabelEn != nil {
		houseAd.CTALabelEn = *req.CTALabelEn
	}
	if req.ImageURL != nil {
		houseAd.ImageURL = *req.ImageURL
	}
	if req.TargetURL != nil {
		houseAd.TargetURL = *req.TargetURL
	}
	if req.IsEnabled != nil {
		houseAd.IsEnabled = *req.IsEnabled
	}
	if err := s.Repo.UpdateHouseAd(houseAd); err != nil {
		return nil, err
	}
	return houseAd, nil
}

func (s *Service) DeleteHouseAd(externalID string) error {
	return s.Repo.DeleteHouseAd(externalID)
}

// PlacementPriceView is the admin/API-facing representation of a placement's
// pricing: the stored monthly rate, the effective rate after an active promo,
// and the promo metadata.
type PlacementPriceView struct {
	Placement           string     `json:"placement"`
	Name                string     `json:"name"`
	MonthlyRate         float64    `json:"monthly_rate"`
	EffectiveMonthlyRate float64   `json:"effective_monthly_rate"`
	Currency            string     `json:"currency"`
	PromoPct            float64    `json:"promo_pct"`
	PromoLabel          string     `json:"promo_label"`
	PromoActive         bool       `json:"promo_active"`
	PromoStartAt        *time.Time `json:"promo_start_at,omitempty"`
	PromoEndAt          *time.Time `json:"promo_end_at,omitempty"`
}

// GetPlacementPrices returns the pricing for every sellable placement, merging
// the DB rows (customizable, migration 000087) with the code-map defaults as
// fallback. VolumeDiscounts is included so clients can mirror the tiered
// long-term discount without hardcoding it.
func (s *Service) GetPlacementPrices() ([]PlacementPriceView, map[string]float64) {
	rows, err := s.Repo.FindAllPlacementPrices()
	if err != nil {
		rows = nil
	}
	byPlacement := make(map[string]*AdPlacementPrice, len(rows))
	for i := range rows {
		byPlacement[rows[i].Placement] = &rows[i]
	}

	now := time.Now()
	prices := make([]PlacementPriceView, 0, len(MonthlyPrices))
	for placement := range MonthlyPrices {
		view := PlacementPriceView{
			Placement: placement,
			Name:      PlacementName(placement),
			Currency:  "IDR",
		}
		if row, ok := byPlacement[placement]; ok {
			view.MonthlyRate = row.MonthlyRate
			view.Currency = row.Currency
			view.PromoPct = row.PromoPct
			view.PromoLabel = row.PromoLabel
			view.PromoStartAt = row.PromoStartAt
			view.PromoEndAt = row.PromoEndAt
			view.PromoActive = row.PromoActive(now)
			view.EffectiveMonthlyRate = row.EffectiveMonthlyRate(now)
		} else {
			view.MonthlyRate = MonthlyPrices[placement]
			view.EffectiveMonthlyRate = view.MonthlyRate
		}
		prices = append(prices, view)
	}

	discounts := make(map[string]float64, len(VolumeDiscounts))
	for months, pct := range VolumeDiscounts {
		discounts[fmt.Sprintf("%d", months)] = pct
	}
	return prices, discounts
}

// UpdatePlacementPriceRequest is the admin payload for customizing a placement's
// monthly rate and/or promo. All fields optional; nil fields keep current value.
type UpdatePlacementPriceRequest struct {
	MonthlyRate  *float64   `json:"monthly_rate"`
	Currency     *string    `json:"currency"`
	PromoPct     *float64   `json:"promo_pct"`
	PromoLabel   *string    `json:"promo_label"`
	PromoStartAt *time.Time `json:"promo_start_at"`
	PromoEndAt   *time.Time `json:"promo_end_at"`
}

// UpdatePlacementPrice upserts the price row for a placement. It refuses
// unknown (non-sellable) placements so the table can't drift from the code map.
func (s *Service) UpdatePlacementPrice(placement string, req UpdatePlacementPriceRequest) (*AdPlacementPrice, error) {
	if !IsSellablePlacement(placement) {
		return nil, domain.NewAPIError(http.StatusBadRequest, domain.CodeValidationFailed,
			"unknown or non-sellable placement", nil)
	}

	price, err := s.Repo.FindPlacementPrice(placement)
	if err != nil {
		price = &AdPlacementPrice{Placement: placement, MonthlyRate: MonthlyPrices[placement], Currency: "IDR"}
	}
	if req.MonthlyRate != nil {
		if *req.MonthlyRate < 0 {
			return nil, domain.NewAPIError(http.StatusBadRequest, domain.CodeValidationFailed,
				"monthly_rate cannot be negative", nil)
		}
		price.MonthlyRate = *req.MonthlyRate
	}
	if req.Currency != nil {
		price.Currency = *req.Currency
	}
	if req.PromoPct != nil {
		if *req.PromoPct < 0 || *req.PromoPct > 100 {
			return nil, domain.NewAPIError(http.StatusBadRequest, domain.CodeValidationFailed,
				"promo_pct must be between 0 and 100", nil)
		}
		price.PromoPct = *req.PromoPct
	}
	if req.PromoLabel != nil {
		price.PromoLabel = *req.PromoLabel
	}
	if req.PromoStartAt != nil {
		price.PromoStartAt = req.PromoStartAt
	}
	if req.PromoEndAt != nil {
		price.PromoEndAt = req.PromoEndAt
	}

	if err := s.Repo.UpsertPlacementPrice(price); err != nil {
		return nil, err
	}
	return price, nil
}
