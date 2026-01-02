package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/functions"
)

type SignInResponse struct {
	Uid        string `json:"uid"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	ProfilePic string `json:"profile_pic"`
	NewAccount bool   `json:"new_account"`
}

type SignInResponse struct {
	Uid        string `json:"uid"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	ProfilePic string `json:"profile_pic"`
	NewAccount bool   `json:"new_account"`
}

func (h *Handler) SignInHandler(c echo.Context) error {
	// ✅ Get UID
	uid, ok := c.Get("uid").(string)
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}
	idToken, ok := c.Get("idToken").(string)
	if !ok || idToken == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid token")
	}
	log.Printf("SignInHandler called for uid: %s", uid)
	profile_pic, name, email, isNew, err := functions.EnsureLogin(c.Request().Context(), h.Queries, uid, idToken)
	if err != nil {
		c.Logger().Errorf("EnsurePlayer failed: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := SignInResponse{
		Uid:        uid,
		Name:       name,
		Email:      email,
		ProfilePic: profile_pic,
		NewAccount: isNew,
	}
	log.Printf("User signed in: %s (new account: %v), name: %s, email %s, profilePic: %s", uid, isNew, name, email, profile_pic)
	return c.JSON(http.StatusOK, resp)
}
