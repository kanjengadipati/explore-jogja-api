package payment

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"pleco-api/internal/modules/adcampaign"
	"pleco-api/internal/modules/audit"
	"pleco-api/internal/modules/business"
	"pleco-api/internal/modules/subscription"
	"pleco-api/internal/services"
)

// MidtransClient is a narrow interface over the Midtrans provider adapter, keeping
// this module free of SDK imports and making it trivially mockable in tests.
type MidtransClient interface {
	CreateSnapTransaction(orderID string, amount int64, itemName, customerName, customerEmail string) (token, redirectURL string, err error)
	VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool
	GetTransactionStatus(orderID string) (string, error)
}

// CommissionRecorder is notified when a payment is marked paid, so the
// commission module can record the sales commission for it (if the payer was
// referred by a sales user). Optional — nil simply skips commission tracking.
type CommissionRecorder interface {
	RecordFromTransaction(payerUserID, paymentTransactionID uint, orderID, subjectType string, grossAmount float64) error
}

// BonusRecorder is notified when a payment is marked paid, so the bonus module
// can record onboarding (first payment) and milestone bonuses. Optional — nil
// simply skips bonus tracking.
type BonusRecorder interface {
	RecordFromTransaction(payerUserID, paymentTransactionID uint, paidAt time.Time) error
}

type Service struct {
	Repo            Repository
	Midtrans        MidtransClient
	AdCampaignSvc   *adcampaign.Service
	SubscriptionSvc *subscription.Service
	AuditSvc        *audit.Service
	EmailSvc        services.EmailService
	BizRepo         business.Repository
	// CommissionSvc is wired in after construction (see appsetup/router.go) —
	// keeps this module's constructor signature/call order unaffected.
	CommissionSvc CommissionRecorder
	// BonusSvc is wired in after construction, same as CommissionSvc.
	BonusSvc BonusRecorder
}

func NewService(repo Repository, midtrans MidtransClient, adCampaignSvc *adcampaign.Service, subscriptionSvc *subscription.Service, auditSvc *audit.Service, emailSvc services.EmailService, bizRepo business.Repository) *Service {
	return &Service{
		Repo:            repo,
		Midtrans:        midtrans,
		AdCampaignSvc:   adCampaignSvc,
		SubscriptionSvc: subscriptionSvc,
		AuditSvc:        auditSvc,
		EmailSvc:        emailSvc,
		BizRepo:         bizRepo,
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
	fmt.Printf("DEBUGPAY after midtrans token=%q redirect=%q err=%v midtransType=%T\n", token, redirectURL, err, s.Midtrans)
	if err != nil {
		// %v (not %w): midtrans-go v1.3.8 returns an *Error whose Unwrap() panics on a
		// typed-nil RawError — errors.As in httpx.HandleError would nil-deref. Flatten
		// the chain so the error handler never traverses into it.
		return nil, "", fmt.Errorf("midtrans create snap transaction failed: %v", err)
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
		if s.CommissionSvc != nil && tx.CreatedByUserID != nil {
			if err := s.CommissionSvc.RecordFromTransaction(*tx.CreatedByUserID, tx.ID, tx.OrderID, tx.SubjectType, tx.Amount); err != nil {
				// Commission tracking is secondary to the payment itself succeeding —
				// log-and-continue rather than fail the whole webhook over it.
				fmt.Printf("failed to record sales commission for order %s: %v\n", tx.OrderID, err)
			}
		}
		if s.BonusSvc != nil && tx.CreatedByUserID != nil {
			paidAt := time.Now()
			if tx.PaidAt != nil {
				paidAt = *tx.PaidAt
			}
			if err := s.BonusSvc.RecordFromTransaction(*tx.CreatedByUserID, tx.ID, paidAt); err != nil {
				// Same log-and-continue policy as commissions.
				fmt.Printf("failed to record sales bonus for order %s: %v\n", tx.OrderID, err)
			}
		}
		return s.propagatePaidStatus(tx)
	}
	return nil
}

// propagatePaidStatus reuses the existing Update() methods on partner/adcampaign services —
// never writes to those tables directly (§7.7).
func (s *Service) propagatePaidStatus(tx *PaymentTransaction) error {
	paid := StatusPaid
	switch tx.SubjectType {
	case SubjectAdCampaign:
		_, err := s.AdCampaignSvc.Update(tx.SubjectExternalID, adcampaign.UpdateAdCampaignRequest{
			PaymentStatus: &paid,
		})
		return err
	case SubjectSubscription:
		_, err := s.SubscriptionSvc.Upgrade(tx.SubjectExternalID, tx.Amount)
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
