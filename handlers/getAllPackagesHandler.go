package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/config"
	"github.com/rakshitg600/notakto-solo/contextkey"

	"github.com/rakshitg600/notakto-solo/usecase"
)

type GetAllPackagesResponse struct {
	CoinPackages []config.CoinPackage `json:"coinPackages"`
}

func (h *Handler) GetAllPackagesHandler(c echo.Context) error {
	uid, ok := contextkey.UIDFromContext(c.Request().Context())
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}

	packages, err := usecase.EnsureGetAllPackages(c.Request().Context(), h.Pool)
	if err != nil {
		c.Logger().Errorf("EnsureGetAllPackages failed: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get all packages")
	}
	responsePackages := make([]config.CoinPackage, len(packages))
	for i, pkg := range packages {
		responsePackages[i] = config.CoinPackage{
			PackageID:      pkg.PackageID,
			PackageName:    pkg.PackageName,
			Coins:          pkg.Coins,
			VisualCoins:    pkg.VisualCoins,
			AmountCents:    pkg.AmountCents,
			Currency:       pkg.Currency,
			DefaultPackage: pkg.DefaultPackage,
		}
	}

	return c.JSON(http.StatusOK, GetAllPackagesResponse{
		CoinPackages: responsePackages,
	})
}
