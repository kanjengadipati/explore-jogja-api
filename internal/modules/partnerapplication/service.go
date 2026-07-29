package partnerapplication

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"pleco-api/internal/modules/audit"
	"pleco-api/internal/modules/notification"
	"pleco-api/internal/modules/partner"
	"pleco-api/internal/modules/user"
	"pleco-api/internal/utils"
)

type Service struct {
	Repo       Repository
	PartnerSvc *partner.Service
	UserSvc    *user.Service
	AuditSvc   *audit.Service
	NotifSvc   *notification.Service
}

func NewService(repo Repository, partnerSvc *partner.Service, userSvc *user.Service, auditSvc *audit.Service, notifSvc *notification.Service) *Service {
	return &Service{Repo: repo, PartnerSvc: partnerSvc, UserSvc: userSvc, AuditSvc: auditSvc, NotifSvc: notifSvc}
}

type ApplyRequest struct {
	BusinessName string   `json:"business_name" binding:"required"`
	Category     string   `json:"category" binding:"required"`
	Location     string   `json:"location"`
	Locations    []string `json:"locations"`
	Phone        string   `json:"phone"`
	Email        string   `json:"email"`
}

func (s *Service) Apply(req ApplyRequest, applicantUserID uint) (*PartnerApplication, error) {
	now := time.Now()
	locArr := make(utils.JSONArr, len(req.Locations))
	for i, l := range req.Locations {
		locArr[i] = l
	}
	if len(locArr) == 0 && req.Location != "" {
		locArr = utils.JSONArr{req.Location}
	}
	app := PartnerApplication{
		ExternalID:      uuid.New().String(),
		ApplicantUserID: applicantUserID,
		BusinessName:    req.BusinessName,
		Category:        req.Category,
		Location:        req.Location,
		Locations:       locArr,
		Phone:           req.Phone,
		Email:           req.Email,
		Status:          StatusPending,
		SubmittedAt:     &now,
	}
	if err := s.Repo.Create(&app); err != nil {
		return nil, err
	}
	return &app, nil
}

func (s *Service) GetByStatus(status string) ([]PartnerApplication, error) {
	return s.Repo.FindByStatus(status)
}

func (s *Service) GetOwned(userID uint) ([]PartnerApplication, error) {
	return s.Repo.FindByApplicant(userID)
}

func (s *Service) Approve(externalID string, adminUserID uint) (*PartnerApplication, error) {
	app, err := s.Repo.FindByExternalID(externalID)
	if err != nil {
		return nil, errors.New("application not found")
	}
	if app.Status != StatusPending {
		return nil, errors.New("application already reviewed")
	}

	now := time.Now()
	app.Status = StatusApproved
	app.ReviewedAt = &now
	app.ReviewedBy = &adminUserID

	draft := &partner.Partner{
		ExternalID:  uuid.New().String(),
		Name:        app.BusinessName,
		Category:    app.Category,
		Location:    app.Location,
		Phone:       app.Phone,
		OwnerUserID: &app.ApplicantUserID,
		Status:      partner.StatusDraft,
	}
	if err := s.PartnerSvc.Create(draft); err != nil {
		return nil, err
	}
	app.ConvertedPartnerExternalID = &draft.ExternalID

	if err := s.Repo.Update(app); err != nil {
		return nil, err
	}

	if s.UserSvc != nil {
		_ = s.UserSvc.PromoteToPartnerRole(app.ApplicantUserID)
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &adminUserID,
		Action:      "partner_application.approved",
		Resource:    "partner_application",
		Status:      "success",
		Description: "Application " + app.ExternalID + " (" + app.BusinessName + ") approved, draft partner " + draft.ExternalID + " created",
	})
	if s.NotifSvc != nil {
		_ = s.NotifSvc.Notify(app.ApplicantUserID, "Aplikasi Mitra Disetujui",
			"Aplikasi Anda untuk '"+app.BusinessName+"' disetujui. Silakan lengkapi listing Anda untuk mulai tayang.", "application")
	}

	return app, nil
}

func (s *Service) Reject(externalID, reason string, adminUserID uint) (*PartnerApplication, error) {
	app, err := s.Repo.FindByExternalID(externalID)
	if err != nil {
		return nil, errors.New("application not found")
	}
	if app.Status != StatusPending {
		return nil, errors.New("application already reviewed")
	}

	now := time.Now()
	app.Status = StatusRejected
	app.RejectionReason = reason
	app.ReviewedAt = &now
	app.ReviewedBy = &adminUserID

	if err := s.Repo.Update(app); err != nil {
		return nil, err
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &adminUserID,
		Action:      "partner_application.rejected",
		Resource:    "partner_application",
		Status:      "success",
		Description: "Application " + app.ExternalID + " rejected. Reason: " + reason,
	})
	if s.NotifSvc != nil {
		content := "Aplikasi Anda untuk '" + app.BusinessName + "' belum bisa kami setujui."
		if reason != "" {
			content += " Alasan: " + reason
		}
		_ = s.NotifSvc.Notify(app.ApplicantUserID, "Aplikasi Mitra Ditolak", content, "application")
	}

	return app, nil
}
