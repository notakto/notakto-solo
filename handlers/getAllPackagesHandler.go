package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/usecase"
)

type CoinPackageResponse struct {
	PackageID   string `json:"packageId"`
	PackageName string `json:"packageName"`
	Coins       int32  `json:"coins"`
	VisualCoins int32  `json:"visualCoins"`
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
			PackageName: pkg.PackageName,
			Coins:       pkg.Coins,
			VisualCoins: pkg.VisualCoins,
			AmountCents: pkg.AmountCents,
			Currency:    pkg.Currency,
		}
	}

	return c.JSON(http.StatusOK, GetAllPackagesResponse{
		CoinPackages: responsePackages,
	})
}
