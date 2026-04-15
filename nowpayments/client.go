package nowpayments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.nowpayments.io/v1"

type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type InvoiceRequest struct {
	PriceAmount      float64 `json:"price_amount"`
	PriceCurrency    string  `json:"price_currency"`
	OrderID          string  `json:"order_id,omitempty"`
	OrderDescription string  `json:"order_description,omitempty"`
	IPNCallbackURL   string  `json:"ipn_callback_url,omitempty"`
	SuccessURL       string  `json:"success_url,omitempty"`
	CancelURL        string  `json:"cancel_url,omitempty"`
	IsFixedRate      bool    `json:"is_fixed_rate,omitempty"`
	IsFeePaidByUser  bool    `json:"is_fee_paid_by_user,omitempty"`
}

type InvoiceResponse struct {
	ID               string `json:"id"`
	OrderID          string `json:"order_id"`
	OrderDescription string `json:"order_description"`
	PriceAmount      string `json:"price_amount"`
	PriceCurrency    string `json:"price_currency"`
	PayCurrency      string `json:"pay_currency"`
	IPNCallbackURL   string `json:"ipn_callback_url"`
	InvoiceURL       string `json:"invoice_url"`
	SuccessURL       string `json:"success_url"`
	CancelURL        string `json:"cancel_url"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

func (c *Client) CreateInvoice(ctx context.Context, req InvoiceRequest) (*InvoiceResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal invoice request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/invoice", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build invoice request: %w", err)
	}
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send invoice request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read invoice response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("nowpayments create invoice: status %d: %s", resp.StatusCode, string(respBody))
	}

	var out InvoiceResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("unmarshal invoice response: %w", err)
	}
	return &out, nil
}
