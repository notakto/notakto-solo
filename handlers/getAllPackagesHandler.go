package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/usecase"
)

type CoinPackageResponse struct {
	PackageID   string `json:"packageId"`
	Coins       int32  `json:"coins"`
	AmountCents int32  `json:"amountCents"`
	Currency    string `json:"currency"`
}

type GetAllPackagesResponse struct {
	CoinPackages []CoinPackageResponse `json:"coinPackages"`
}

func (h *Handler) GetAllPackagesHandler(c echo.Context) error {
	uid, ok := contextkey.UIDFromContext(c.Request().Context())
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}

	packages := usecase.EnsureGetAllPackages(c.Request().Context())
	responsePackages := make([]CoinPackageResponse, len(packages))
	for i, pkg := range packages {
		responsePackages[i] = CoinPackageResponse{
			PackageID:   pkg.ID,
			Coins:       pkg.Coins,
			AmountCents: pkg.AmountCents,
			Currency:    pkg.Currency,
		}
	}

	return c.JSON(http.StatusOK, GetAllPackagesResponse{
		CoinPackages: responsePackages,
	})
}
