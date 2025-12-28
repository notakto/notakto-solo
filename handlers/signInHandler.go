package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/functions"
)

type SignInResponse struct {
	Uid        string `json:"uid"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	ProfilePic string `json:"profile_pic"`
	NewAccount bool   `json:"new_account"`
}
type Handler struct {
	Queries *db.Queries
}

func NewHandler(q *db.Queries) *Handler {
	return &Handler{Queries: q}
}

func (h *Handler) SignInHandler(c echo.Context) error {
	// ✅ Get UID
	uid, ok := c.Get("uid").(string)
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}
	idToken, ok := c.Get("idToken").(string)
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid token")
	}
	log.Printf("SignInHandler called for uid: %s", uid)
	profile_pic, name, email, new, err := functions.EnsureLogin(c.Request().Context(), h.Queries, uid, idToken)
	if err != nil {
		c.Logger().Errorf("EnsurePlayer failed: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := SignInResponse{
		Uid:        uid,
		Name:       name,
		Email:      email,
		ProfilePic: profile_pic,
		NewAccount: new,
	}
	log.Printf("User signed in: %s (new account: %v), name: %s, email %s, profilePic: %s", uid, new, name, email, profile_pic)
	return c.JSON(http.StatusOK, resp)
}
