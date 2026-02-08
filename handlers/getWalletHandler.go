package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/usecase"
)

type GetWalletResponse struct {
	Coins   int32  `json:"coins"`
	XP      int32  `json:"xp"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (h *Handler) GetWalletHandler(c echo.Context) error {
	uid, ok := contextkey.UIDFromContext(c.Request().Context())
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}
	log.Printf("GetWalletHandler called for uid: %s", uid)
	coins, xp, err := usecase.EnsureGetWallet(c.Request().Context(), h.Pool, uid)
	if err != nil {
		c.Logger().Errorf("EnsureGetWallet failed: %v", err)
		return c.JSON(http.StatusOK, GetWalletResponse{
			Coins:   coins,
			XP:      xp,
			Success: false,
			Error:   err.Error(),
		})
	}

	resp := GetWalletResponse{
		Success: true,
		Coins:   coins,
		XP:      xp,
	}
	log.Printf("GetWalletHandler completed for uid: %s", uid)
	return c.JSON(http.StatusOK, resp)
}
