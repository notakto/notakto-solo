package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/usecase"
)

type LeaderboardEntryResponse struct {
	Rank     int32  `json:"rank"`
	Username string `json:"username"`
	XP       int32  `json:"xp"`
}

type LeaderboardResponse struct {
	Leaderboard []LeaderboardEntryResponse `json:"leaderboard"`
	CurrentUser LeaderboardEntryResponse   `json:"current_user"`
}

func (h *Handler) LeaderboardHandler(c echo.Context) error {
	uid, ok := contextkey.UIDFromContext(c.Request().Context())
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}

	log.Printf("LeaderboardHandler called for uid: %s", uid)

	leaderboard, err := usecase.EnsureGetLeaderboard(c.Request().Context(), h.Pool)
	if err != nil {
		c.Logger().Errorf("EnsureGetLeaderboard failed: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get leaderboard")
	}

	resp := LeaderboardResponse{
		Leaderboard: make([]LeaderboardEntryResponse, 0, len(leaderboard.Leaderboard)),
		CurrentUser: LeaderboardEntryResponse{
			Rank:     leaderboard.CurrentUser.Rank,
			Username: leaderboard.CurrentUser.Username,
			XP:       leaderboard.CurrentUser.XP,
		},
	}
	for _, entry := range leaderboard.Leaderboard {
		resp.Leaderboard = append(resp.Leaderboard, LeaderboardEntryResponse{
			Rank:     entry.Rank,
			Username: entry.Username,
			XP:       entry.XP,
		})
	}

	return c.JSON(http.StatusOK, resp)
}
