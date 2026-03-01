package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/coinbase"
	"github.com/rakshitg600/notakto-solo/usecase"
)

type WebhookEvent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

type WebhookPayload struct {
	Event WebhookEvent `json:"event"`
}

func (h *Handler) WebhookHandler(c echo.Context) error {
	// Read raw body bytes before any parsing — needed for HMAC signature verification against the exact payload Coinbase signed. Using c.Bind() would consume the body without preserving the original bytes.
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("webhook: failed to read body: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	signature := c.Request().Header.Get("X-CC-Webhook-Signature")
	if signature == "" {
		log.Printf("webhook: missing signature header")
		return c.NoContent(http.StatusUnauthorized)
	}

	if !coinbase.VerifyWebhookSignature(h.WebhookSecret, signature, body) {
		log.Printf("webhook: invalid signature")
		return c.NoContent(http.StatusUnauthorized)
	}

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("webhook: failed to parse payload: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	eventType := payload.Event.Type
	chargeID := payload.Event.Data.ID
	if chargeID == "" {
		log.Printf("webhook: missing charge ID in event")
		return c.NoContent(http.StatusBadRequest)
	}

	log.Printf("webhook: received event %s for charge %s", eventType, chargeID)

	if err := usecase.EnsureProcessWebhook(c.Request().Context(), h.Pool, eventType, chargeID); err != nil {
		log.Printf("webhook: processing failed for charge %s: %v", chargeID, err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
