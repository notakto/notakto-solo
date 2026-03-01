package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"

	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/store"
)

type PaymentStatusResponse struct {
	ChargeID    string `json:"chargeId"`
	PackageID   string `json:"packageId"`
	Coins       int32  `json:"coins"`
	AmountCents int32  `json:"amountCents"`
	Status      string `json:"status"`
	HostedURL   string `json:"hostedUrl"`
}

func (h *Handler) PaymentStatusHandler(c echo.Context) error {
	uid, ok := contextkey.UIDFromContext(c.Request().Context())
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}

	chargeID := c.QueryParam("chargeId")
	if chargeID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "chargeId query parameter is required")
	}

	log.Printf("PaymentStatusHandler called for uid: %s, chargeId: %s", uid, chargeID)

	queries := db.New(h.Pool)
	payment, err := store.GetPaymentById(c.Request().Context(), queries, chargeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "payment not found")
		}
		c.Logger().Errorf("GetPaymentById failed: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch payment")
	}

	if payment.Uid != uid {
		return echo.NewHTTPError(http.StatusForbidden, "access denied")
	}

	return c.JSON(http.StatusOK, PaymentStatusResponse{
		ChargeID:    payment.ID,
		PackageID:   payment.PackageID,
		Coins:       payment.Coins,
		AmountCents: payment.AmountCents,
		Status:      payment.Status,
		HostedURL:   payment.HostedUrl,
	})
}
