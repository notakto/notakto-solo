package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/imagekitservice"
	"github.com/rakshitg600/notakto-solo/store"
)

var (
	ErrProfileImagePlayerNotFound = errors.New("player profile not found")
	ErrInvalidProfileImageRequest = errors.New("invalid profile-image request")
)

func EnsureProfileImageUploadAuth(
	ctx context.Context,
	pool *pgxpool.Pool,
	imagekitClient *imagekitservice.Client,
	originalFilename string,
) (imagekitservice.UploadAuth, error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return imagekitservice.UploadAuth{}, errors.New("missing or invalid uid in context")
	}
	if _, err := store.GetPlayerById(ctx, db.New(pool)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return imagekitservice.UploadAuth{}, ErrProfileImagePlayerNotFound
		}
		return imagekitservice.UploadAuth{}, fmt.Errorf("look up player profile: %w", err)
	}

	auth, err := imagekitClient.GenerateUploadAuth(uid, originalFilename)
	if err != nil {
		if errors.Is(err, imagekitservice.ErrInvalidFilename) ||
			errors.Is(err, imagekitservice.ErrUnsupportedExtension) {
			return imagekitservice.UploadAuth{}, fmt.Errorf("%w: %v", ErrInvalidProfileImageRequest, err)
		}
		return imagekitservice.UploadAuth{}, fmt.Errorf("generate ImageKit upload auth: %w", err)
	}
	return auth, nil
}
