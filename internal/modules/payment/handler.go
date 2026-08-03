package payment

import (
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"

	"pleco-api/internal/httpx"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) CreateTransaction(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ValidationError(c, err)
		return
	}
	userID, _ := httpx.GetUserIDFromContext(c)
	tx, redirectURL, err := h.Service.CreateTransaction(req, &userID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 201, "Invoice created", CreateTransactionResponse{
		OrderID:     tx.OrderID,
		SnapToken:   tx.MidtransToken,
		RedirectURL: redirectURL,
	}, nil)
}

func (h *Handler) ListTransactions(c *gin.Context) {
	txs, err := h.Service.GetAll()
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Transactions fetched", txs, nil)
}

func (h *Handler) CreateSubscriptionUpgrade(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ValidationError(c, err)
		return
	}

	userID, _ := httpx.GetUserIDFromContext(c)

	// Resolve the caller's own business so the invoice is always created
	// against their own subscription, never someone else's.
	b, err := h.Service.BizRepo.FindByIDAndOwner(c.Param("id"), userID)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Business not found or not owned")
		return
	}

	sub, err := h.Service.SubscriptionSvc.GetByBusinessExternalID(b.ExternalID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	req.SubjectType = SubjectSubscription
	req.SubjectExternalID = sub.ExternalID

	tx, redirectURL, err := h.Service.CreateTransaction(req, &userID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 201, "Invoice created", CreateTransactionResponse{
		OrderID:     tx.OrderID,
		SnapToken:   tx.MidtransToken,
		RedirectURL: redirectURL,
	}, nil)
}

func (h *Handler) HandleNotification(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		httpx.ErrorWithCode(c, 400, "INVALID_BODY", "Cannot read request body")
		return
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		httpx.ErrorWithCode(c, 400, "INVALID_JSON", "Invalid notification payload")
		return
	}
	if err := h.Service.HandleNotification(payload); err != nil {
		if err.Error() == "invalid midtrans signature" {
			// Return 401 only for spoofed webhooks — Midtrans will not retry 401.
			httpx.ErrorWithCode(c, 401, "INVALID_SIGNATURE", "Signature verification failed")
			return
		}
		// Return 200 for all other errors so Midtrans doesn't trigger retry storms.
		httpx.Success(c, 200, err.Error(), nil, nil)
		return
	}
	httpx.Success(c, 200, "Notification processed", nil, nil)
}
