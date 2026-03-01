package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/usecase"
)

type CreateChargeRequest struct {
	PackageID string `json:"packageId"`
}

type CreateChargeResponse struct {
	ChargeID  string `json:"chargeId"`
	HostedURL string `json:"hostedUrl"`
}

func (h *Handler) CreateChargeHandler(c echo.Context) error {
	uid, ok := contextkey.UIDFromContext(c.Request().Context())
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}

	var req CreateChargeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.PackageID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "packageId is required")
	}

	log.Printf("CreateChargeHandler called for uid: %s, package: %s", uid, req.PackageID)

	chargeID, hostedURL, err := usecase.EnsureCreateCharge(c.Request().Context(), h.Pool, h.CommerceClient, req.PackageID)
	if err != nil {
		c.Logger().Errorf("EnsureCreateCharge failed: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create charge")
	}

	log.Printf("CreateChargeHandler completed for uid: %s, chargeId: %s", uid, chargeID)
	return c.JSON(http.StatusOK, CreateChargeResponse{
		ChargeID:  chargeID,
		HostedURL: hostedURL,
	})
}
