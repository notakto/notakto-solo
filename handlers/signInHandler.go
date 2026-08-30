package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/logic"
	"github.com/rakshitg600/notakto-solo/usecase"
)

type SignInResponse struct {
	Uid        string `json:"uid"`
	Username   string `json:"username"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	ProfilePic string `json:"profile_pic"`
	NewAccount bool   `json:"new_account"`
}

func (h *Handler) SignInHandler(c echo.Context) error {
	uid, ok := contextkey.UIDFromContext(c.Request().Context())
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}
	log.Printf("SignInHandler called for uid: %s", uid)
	result, err := usecase.EnsureLogin(
		c.Request().Context(),
		h.Pool,
		h.AuthClient,
		h.ValkeyClient,
	)

	if err != nil {
		if errors.Is(err, logic.ErrRedisLockAlreadyHeld) {
			return echo.NewHTTPError(http.StatusTooManyRequests, "Could not acquire lock, try again later")
		}
		c.Logger().Errorf("EnsurePlayer failed: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := SignInResponse{
		Uid:        uid,
		Username:   result.Username,
		Name:       result.Name,
		Email:      result.Email,
		ProfilePic: result.ProfilePic,
		NewAccount: result.IsNew,
	}
	log.Printf("User signed in: %s (new account: %v), username: %s, name: %s, email %s, profilePic: %s", uid, result.IsNew, result.Username, result.Name, result.Email, result.ProfilePic)
	return c.JSON(http.StatusOK, resp)
}
