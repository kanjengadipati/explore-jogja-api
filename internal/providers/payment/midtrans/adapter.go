package midtrans

// Adapter bridges *Client (which returns SDK types) to the payment.MidtransClient
// interface (which uses only primitive types), keeping the payment module free of
// third-party SDK imports.

type Adapter struct{ client *Client }

func NewAdapter(c *Client) *Adapter { return &Adapter{client: c} }

func (a *Adapter) CreateSnapTransaction(orderID string, amount int64, itemName, customerName, customerEmail string) (string, string, error) {
	return a.client.CreateSnapTransaction(orderID, amount, itemName, customerName, customerEmail)
}

func (a *Adapter) VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	return a.client.VerifySignature(orderID, statusCode, grossAmount, signatureKey)
}

func (a *Adapter) GetTransactionStatus(orderID string) (string, error) {
	resp, err := a.client.GetTransactionStatus(orderID)
	if err != nil {
		return "", err
	}
	return resp.TransactionStatus, nil
}
