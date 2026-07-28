package payment

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"pleco-api/internal/modules/adcampaign"
	"pleco-api/internal/modules/audit"
	"pleco-api/internal/modules/partner"
	"pleco-api/internal/services"
)

// MidtransClient is a narrow interface over the Midtrans provider adapter, keeping
// this module free of SDK imports and making it trivially mockable in tests.
type MidtransClient interface {
	CreateSnapTransaction(orderID string, amount int64, itemName, customerName, customerEmail string) (token, redirectURL string, err error)
	VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool
	GetTransactionStatus(orderID string) (string, error)
}

type Service struct {
	Repo          Repository
	Midtrans      MidtransClient
	PartnerSvc    *partner.Service
	AdCampaignSvc *adcampaign.Service
	AuditSvc      *audit.Service
	EmailSvc      services.EmailService
}

func NewService(repo Repository, midtrans MidtransClient, partnerSvc *partner.Service, adCampaignSvc *adcampaign.Service, auditSvc *audit.Service, emailSvc services.EmailService) *Service {
	return &Service{
		Repo:          repo,
		Midtrans:      midtrans,
		PartnerSvc:    partnerSvc,
		AdCampaignSvc: adCampaignSvc,
		AuditSvc:      auditSvc,
		EmailSvc:      emailSvc,
	}
}

type CreateTransactionRequest struct {
	SubjectType       string  `json:"subject_type" binding:"required"`
	SubjectExternalID string  `json:"subject_external_id" binding:"required"`
	Amount            float64 `json:"amount" binding:"required,gt=0"`
	ItemName          string  `json:"item_name" binding:"required"`
	CustomerName      string  `json:"customer_name" binding:"required"`
	CustomerEmail     string  `json:"customer_email" binding:"required,email"`
}

type CreateTransactionResponse struct {
	OrderID     string `json:"order_id"`
	SnapToken   string `json:"snap_token"`
	RedirectURL string `json:"redirect_url"`
}

func (s *Service) CreateTransaction(req CreateTransactionRequest, createdByUserID *uint) (*PaymentTransaction, string, error) {
	orderID := "PLC-" + uuid.New().String()

	token, redirectURL, err := s.Midtrans.CreateSnapTransaction(
		orderID, int64(req.Amount), req.ItemName, req.CustomerName, req.CustomerEmail,
	)
	if err != nil {
		return nil, "", err
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	tx := &PaymentTransaction{
		OrderID:           orderID,
		SubjectType:       req.SubjectType,
		SubjectExternalID: req.SubjectExternalID,
		Amount:            req.Amount,
		Status:            StatusPending,
		MidtransToken:     token,
		ExpiresAt:         &expiresAt,
		CreatedByUserID:   createdByUserID,
	}
	if err := s.Repo.Create(tx); err != nil {
		return nil, "", err
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: createdByUserID,
		Action:      "payment.invoice_created",
		Resource:    req.SubjectType,
		Status:      "success",
		Description: "Invoice created for " + req.SubjectExternalID + " amount " + tx.Currency,
	})

	return tx, redirectURL, nil
}

func (s *Service) HandleNotification(payload map[string]any) error {
	orderID, _ := payload["order_id"].(string)
	statusCode, _ := payload["status_code"].(string)
	grossAmount, _ := payload["gross_amount"].(string)
	signatureKey, _ := payload["signature_key"].(string)
	transactionStatus, _ := payload["transaction_status"].(string)
	fraudStatus, _ := payload["fraud_status"].(string)
	paymentType, _ := payload["payment_type"].(string)
	transactionID, _ := payload["transaction_id"].(string)

	if !s.Midtrans.VerifySignature(orderID, statusCode, grossAmount, signatureKey) {
		return errors.New("invalid midtrans signature")
	}

	tx, err := s.Repo.FindByOrderID(orderID)
	if err != nil {
		return err
	}

	newStatus := tx.Status
	switch transactionStatus {
	case "capture":
		if fraudStatus == "accept" {
			newStatus = StatusPaid
		}
	case "settlement":
		newStatus = StatusPaid
	case "deny", "cancel":
		newStatus = StatusFailed
	case "expire":
		newStatus = StatusExpired
	}

	tx.Status = newStatus
	tx.PaymentType = paymentType
	tx.TransactionID = transactionID
	tx.FraudStatus = fraudStatus
	if newStatus == StatusPaid {
		now := time.Now()
		tx.PaidAt = &now
	}
	if err := s.Repo.Update(tx); err != nil {
		return err
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		Action:      "payment.status_changed",
		Resource:    tx.SubjectType,
		Status:      "success",
		Description: "Order " + orderID + " → " + newStatus,
	})

	if newStatus == StatusPaid {
		return s.propagatePaidStatus(tx)
	}
	return nil
}

// propagatePaidStatus reuses the existing Update() methods on partner/adcampaign services —
// never writes to those tables directly (§7.7).
func (s *Service) propagatePaidStatus(tx *PaymentTransaction) error {
	paid := StatusPaid
	switch tx.SubjectType {
	case SubjectPartnerSponsorship:
		_, err := s.PartnerSvc.Update(tx.SubjectExternalID, partner.UpdatePartnerRequest{
			SponsorPaymentStatus: &paid,
		})
		if err == nil && s.EmailSvc != nil {
			_ = s.EmailSvc.SendPaymentConfirmation(tx.SubjectExternalID, tx.SubjectExternalID, tx.Amount, tx.Currency)
		}
		return err
	case SubjectAdCampaign:
		_, err := s.AdCampaignSvc.Update(tx.SubjectExternalID, adcampaign.UpdateAdCampaignRequest{
			PaymentStatus: &paid,
		})
		return err
	}
	return errors.New("unknown subject type: " + tx.SubjectType)
}

func (s *Service) GetAll() ([]PaymentTransaction, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetBySubject(subjectType, subjectExternalID string) ([]PaymentTransaction, error) {
	return s.Repo.FindBySubject(subjectType, subjectExternalID)
}

// PollPendingExpired is called by a scheduler to catch payments that settled at Midtrans
// but whose webhook never arrived (§7.5 point 4).
func (s *Service) PollPendingExpired() error {
	pending, err := s.Repo.FindPendingExpired()
	if err != nil {
		return err
	}
	for _, tx := range pending {
		status, err := s.Midtrans.GetTransactionStatus(tx.OrderID)
		if err != nil {
			continue
		}
		switch status {
		case "settlement", "capture":
			tx.Status = StatusPaid
			_ = s.Repo.Update(&tx)
			_ = s.propagatePaidStatus(&tx)
		case "expire":
			tx.Status = StatusExpired
			_ = s.Repo.Update(&tx)
		}
	}
	return nil
}
