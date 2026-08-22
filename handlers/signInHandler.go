package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/contextkey"
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
	profilePic, name, email, username, isNew, err := usecase.EnsureLogin(
		c.Request().Context(),
		h.Pool,
		h.AuthClient,
	)

	if err != nil {
		c.Logger().Errorf("EnsurePlayer failed: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := SignInResponse{
		Uid:        uid,
		Username:   username,
		Name:       name,
		Email:      email,
		ProfilePic: profilePic,
		NewAccount: isNew,
	}
	log.Printf("User signed in: %s (new account: %v), username: %s, name: %s, email %s, profilePic: %s", uid, isNew, username, name, email, profilePic)
	return c.JSON(http.StatusOK, resp)
}
