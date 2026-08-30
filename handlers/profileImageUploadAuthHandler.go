package handlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/usecase"
	"github.com/rakshitg600/notakto-solo/utils"
)

type ProfileImageUploadAuthRequest struct {
	FileName string `json:"fileName"`
}

func (h *Handler) ProfileImageUploadAuthHandler(c echo.Context) error {
	uid, ok := contextkey.UIDFromContext(c.Request().Context())
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}

	var req ProfileImageUploadAuthRequest
	if err := utils.DecodeStrictJSON(utils.DecodeStrictJSONParams{Context: c, Dest: &req}); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	auth, err := usecase.EnsureProfileImageUploadAuth(
		c.Request().Context(),
		h.Pool,
		h.ImageKitClient,
		req.FileName,
	)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidProfileImageRequest):
			return echo.NewHTTPError(http.StatusBadRequest, "fileName must end in .jpg, .jpeg, .png, or .webp")
		case errors.Is(err, usecase.ErrProfileImagePlayerNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "player profile not found; sign in first")
		default:
			c.Logger().Errorf("ProfileImageUploadAuthHandler failed for uid %s: %v", uid, err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create profile-image upload credentials")
		}
	}

	return c.JSON(http.StatusOK, auth)
}
