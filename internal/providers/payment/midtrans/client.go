package midtrans

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

type Client struct {
	snapClient snap.Client
	coreClient coreapi.Client
	serverKey  string
}

func New(serverKey, clientKey string, isProduction bool) *Client {
	env := midtrans.Sandbox
	if isProduction {
		env = midtrans.Production
	}
	var s snap.Client
	s.New(serverKey, env)
	var c coreapi.Client
	c.New(serverKey, env)
	return &Client{snapClient: s, coreClient: c, serverKey: serverKey}
}

func (c *Client) CreateSnapTransaction(orderID string, amount int64, itemName, customerName, customerEmail string) (token, redirectURL string, err error) {
	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{OrderID: orderID, GrossAmt: amount},
		Items: &[]midtrans.ItemDetails{
			{ID: orderID, Price: amount, Qty: 1, Name: itemName},
		},
		CustomerDetail: &midtrans.CustomerDetails{FName: customerName, Email: customerEmail},
	}
	resp, err := c.snapClient.CreateTransaction(req)
	if err != nil {
		return "", "", err
	}
	return resp.Token, resp.RedirectURL, nil
}

// VerifySignature is the sole security gate for webhook notifications.
// Never process a webhook body before this passes.
func (c *Client) VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	raw := orderID + statusCode + grossAmount + c.serverKey
	h := sha512.Sum512([]byte(raw))
	return hex.EncodeToString(h[:]) == signatureKey
}

func (c *Client) GetTransactionStatus(orderID string) (*coreapi.TransactionStatusResponse, error) {
	resp, err := c.coreClient.CheckTransaction(orderID)
	if err != nil {
		return nil, fmt.Errorf("check transaction %s: %w", orderID, err)
	}
	return resp, nil
}
