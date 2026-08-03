package midtrans

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
)

type Client struct {
	coreClient coreapi.Client
	serverKey  string
	snapURL    string
}

func New(serverKey, clientKey string, isProduction bool) *Client {
	env := midtrans.Sandbox
	if isProduction {
		env = midtrans.Production
	}
	var c coreapi.Client
	c.New(serverKey, env)
	
	snapURL := "https://app.sandbox.midtrans.com"
	if isProduction {
		snapURL = "https://app.midtrans.com"
	}
	
	return &Client{coreClient: c, serverKey: serverKey, snapURL: snapURL}
}

func (c *Client) CreateSnapTransaction(orderID string, amount int64, itemName, customerName, customerEmail string) (token, redirectURL string, err error) {
	reqBody := map[string]interface{}{
		"transaction_details": map[string]interface{}{
			"order_id":  orderID,
			"gross_amount": amount,
		},
		"item_details": []map[string]interface{}{
			{
				"id":    orderID,
				"price": amount,
				"quantity": 1,
				"name":  itemName,
			},
		},
		"customer_details": map[string]interface{}{
			"first_name": customerName,
			"email":      customerEmail,
		},
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", c.snapURL+"/snap/v1/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.serverKey, "")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("snap request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("snap API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Token       string `json:"token"`
		RedirectURL string `json:"redirect_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("failed to decode snap response: %v", err)
	}

	return result.Token, result.RedirectURL, nil
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
