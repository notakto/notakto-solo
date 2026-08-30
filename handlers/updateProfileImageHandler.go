package handlers

import (
	"errors"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rakshitg600/notakto-solo/contextkey"
	"github.com/rakshitg600/notakto-solo/usecase"
	"github.com/rakshitg600/notakto-solo/utils"
)

var profileImageFileIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,256}$`)

type UpdateProfileImageRequest struct {
	FileID   string  `json:"fileId"`
	FilePath *string `json:"filePath"`
}
type ProfileImageReferenceResponse struct {
	FileID   string `json:"file_id"`
	FilePath string `json:"file_path"`
}
type ProfileImageUpdateResponse struct {
	ProfilePic   string                        `json:"profile_pic"`
	ProfileImage ProfileImageReferenceResponse `json:"profile_image"`
}

func (h *Handler) UpdateProfileImageHandler(c echo.Context) error {
	uid, ok := contextkey.UIDFromContext(c.Request().Context())
	if !ok || uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized: missing or invalid uid")
	}

	var req UpdateProfileImageRequest
	if err := utils.DecodeStrictJSON(utils.DecodeStrictJSONParams{Context: c, Dest: &req}); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if !profileImageFileIDPattern.MatchString(req.FileID) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid profile-image reference")
	}

	filePath := ""
	if req.FilePath != nil {
		filePath = *req.FilePath
		if !isCanonicalAssetPath(filePath) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid profile-image reference")
		}
	}

	result, err := usecase.EnsureUpdateProfileImage(
		c.Request().Context(),
		h.Pool,
		req.FileID,
		filePath,
	)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidProfileImageReference):
			return echo.NewHTTPError(http.StatusBadRequest, "invalid profile-image reference")
		case errors.Is(err, usecase.ErrProfileImagePlayerNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "player profile not found; sign in first")
		case errors.Is(err, usecase.ErrProfileImageProvider):
			c.Logger().Errorf("UpdateProfileImageHandler ImageKit failure for uid %s: %v", uid, err)
			return echo.NewHTTPError(http.StatusBadGateway, "profile-image provider unavailable")
		default:
			c.Logger().Errorf("UpdateProfileImageHandler failed for uid %s: %v", uid, err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update profile image")
		}
	}

	return c.JSON(http.StatusOK, ProfileImageUpdateResponse{
		ProfilePic: result.URL,
		ProfileImage: ProfileImageReferenceResponse{
			FileID:   result.FileID,
			FilePath: result.FilePath,
		},
	})
}

func isCanonicalAssetPath(value string) bool {
	return value != "" && len(value) <= 1024 && strings.HasPrefix(value, "/") &&
		!strings.HasPrefix(value, "//") && !strings.ContainsAny(value, `\?#`) &&
		path.Clean(value) == value && path.Base(value) != "."
}
