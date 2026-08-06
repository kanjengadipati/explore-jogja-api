package adcampaign

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"pleco-api/internal/modules/subscription"
	"pleco-api/internal/services"
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
// are derived server-side from the placement).
type SelfServiceCreateRequest struct {
	Placement string     `json:"placement" binding:"required"`
	ImageURL  string     `json:"image_url" binding:"required"`
	TargetURL string     `json:"target_url" binding:"required"`
	Category  string     `json:"category"`
	StartAt   *time.Time `json:"start_at"`
	EndAt     *time.Time `json:"end_at"`
}

// CreateSelfService creates an ad campaign on behalf of a business owner.
// The price comes from the per-placement monthly rate (§ pricing.go) and the
// initial status follows the placement's moderation policy: review-required
// slots start pending_review, the rest start pending_payment and go live
// automatically once the payment webhook marks them paid.
func (s *Service) CreateSelfService(businessExternalID, partnerName string, req SelfServiceCreateRequest) (*AdCampaign, error) {
	if !IsSellablePlacement(req.Placement) {
		return nil, errors.New("unknown or non-sellable placement")
	}
	if strings.TrimSpace(req.ImageURL) == "" || strings.TrimSpace(req.TargetURL) == "" {
		return nil, errors.New("image_url and target_url are required")
	}

	exists, err := s.Repo.BusinessExists(businessExternalID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("business not found")
	}

	canCreate, err := s.SubscriptionSvc.CanCreateAdCampaign(businessExternalID)
	if err != nil {
		return nil, err
	}
	if !canCreate {
		return nil, errors.New("business on free plan cannot create ad campaigns")
	}

	var startAt, endAt time.Time
	if req.StartAt != nil {
		startAt = *req.StartAt
	}
	if req.EndAt != nil {
		endAt = *req.EndAt
	}

	campaign := &AdCampaign{
		ExternalID:         uuid.New().String(),
		BusinessExternalID: &businessExternalID,
		PartnerName:        partnerName,
		Placement:          req.Placement,
		ImageURL:           req.ImageURL,
		TargetURL:          req.TargetURL,
		Category:           req.Category,
		StartAt:            startAt,
		EndAt:              endAt,
		Weight:             1,
		IsActive:           true,
		PriceAmount:        PriceFor(req.Placement, startAt, endAt),
		PriceCurrency:      "IDR",
		PaymentStatus:      InitialPaymentStatus(req.Placement),
	}

	if err := s.Repo.Create(campaign); err != nil {
		return nil, err
	}
	return campaign, nil
}

type UpdateAdCampaignRequest struct {
	PartnerName        *string    `json:"partner_name"`
	BusinessExternalID *string    `json:"business_external_id"`
	Placement          *string    `json:"placement"`
	ImageURL           *string    `json:"image_url"`
	TargetURL          *string    `json:"target_url"`
	Category           *string    `json:"category"`
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
