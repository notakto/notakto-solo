package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/nowpayments"
	"github.com/rakshitg600/notakto-solo/usecase"
)

// maxWebhookBodySize caps IPN payloads. NOWPayments bodies are ~1 KB; the
// limit exists to prevent unbounded allocation from a hostile or misbehaving
// caller, not to match any protocol field.
const maxWebhookBodySize = 64 << 10 // 64 KiB

// WebhookRequest is the flat JSON body NOWPayments posts to our callback.
// Only the fields we actually use are typed; the rest are ignored. We use
// json.Number for numeric identifiers to preserve precision and to cover
// NOWPayments emitting them either as quoted strings or bare numbers across
// endpoints.
type WebhookRequest struct {
	PaymentID     json.Number `json:"payment_id"`
	PaymentStatus string      `json:"payment_status"`
	OrderID       string      `json:"order_id"`
	InvoiceID     json.Number `json:"invoice_id"`
}

func (h *Handler) WebhookHandler(c echo.Context) error {
	// Read raw body bytes before any parsing — needed for HMAC signature
	// verification against the exact payload NOWPayments signed. Wrap with
	// MaxBytesReader so a hostile caller can't force unbounded allocation.
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, maxWebhookBodySize)
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			log.Printf("webhook: body exceeded %d bytes", maxWebhookBodySize)
			return c.NoContent(http.StatusRequestEntityTooLarge)
		}
		log.Printf("webhook: failed to read body: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	signature := c.Request().Header.Get("x-nowpayments-sig")
	if signature == "" {
		log.Printf("webhook: missing signature header")
		return c.NoContent(http.StatusUnauthorized)
	}

	if !nowpayments.VerifyIPNSignature(h.IPNSecret, signature, body) {
		log.Printf("webhook: invalid signature")
		return c.NoContent(http.StatusUnauthorized)
	}

	var payload WebhookRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("webhook: failed to parse payload: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	if payload.OrderID == "" {
		log.Printf("webhook: missing order_id in payload (payment_id=%s)", payload.PaymentID.String())
		return c.NoContent(http.StatusBadRequest)
	}
	if payload.PaymentStatus == "" {
		log.Printf("webhook: missing payment_status for order %s", payload.OrderID)
		return c.NoContent(http.StatusBadRequest)
	}

	log.Printf("webhook: received status %s for order %s (payment_id=%s)",
		payload.PaymentStatus, payload.OrderID, payload.PaymentID.String())

	if err := usecase.EnsureProcessWebhook(c.Request().Context(), h.Pool, payload.PaymentStatus, payload.OrderID); err != nil {
		log.Printf("webhook: processing failed for order %s: %v", payload.OrderID, err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
