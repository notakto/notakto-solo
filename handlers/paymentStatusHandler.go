package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/usecase"
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

	payment, err := usecase.EnsureGetPaymentStatus(c.Request().Context(), h.Pool, chargeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "payment not found")
		}
		if errors.Is(err, usecase.ErrPaymentForbidden) {
			return echo.NewHTTPError(http.StatusForbidden, "access denied")
		}
		c.Logger().Errorf("EnsureGetPaymentStatus failed: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch payment")
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
