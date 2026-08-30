package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	imagekitsdk "github.com/imagekit-developer/imagekit-go/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakshitg600/notakto-solo/contextkey"
	db "github.com/rakshitg600/notakto-solo/db/generated"
	"github.com/rakshitg600/notakto-solo/imagekitservice"
	"github.com/rakshitg600/notakto-solo/store"
)

const profileImageCleanupTimeout = 2 * time.Second

var (
	ErrProfileImagePlayerNotFound   = errors.New("player profile not found")
	ErrInvalidProfileImageRequest   = errors.New("invalid profile-image request")
	ErrInvalidProfileImageReference = errors.New("invalid profile-image reference")
	ErrProfileImageProvider         = errors.New("profile-image provider unavailable")
)

type ProfileImageResult struct {
	FileID   string
	FilePath string
	URL      string
}

func EnsureUpdateProfileImage(
	ctx context.Context,
	pool *pgxpool.Pool,
	fileID string,
	filePath string,
) (ProfileImageResult, error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return ProfileImageResult{}, errors.New("missing or invalid uid in context")
	}
	if _, err := store.GetPlayerById(ctx, db.New(pool)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ProfileImageResult{}, ErrProfileImagePlayerNotFound
		}
		return ProfileImageResult{}, fmt.Errorf("look up player profile: %w", err)
	}

	asset, err := imagekitservice.GetFile(ctx, fileID)
	if err != nil {
		if errors.Is(err, imagekitservice.ErrInvalidFileID) || errors.Is(err, imagekitservice.ErrInvalidAsset) {
			return ProfileImageResult{}, fmt.Errorf("%w: %v", ErrInvalidProfileImageReference, err)
		}
		var apiErr *imagekitsdk.Error
		if errors.As(err, &apiErr) && (apiErr.StatusCode == http.StatusBadRequest || apiErr.StatusCode == http.StatusNotFound) {
			return ProfileImageResult{}, fmt.Errorf("%w: asset does not exist", ErrInvalidProfileImageReference)
		}
		return ProfileImageResult{}, fmt.Errorf("%w: %v", ErrProfileImageProvider, err)
	}
	if err := imagekitservice.ValidateAsset(asset, uid, filePath); err != nil {
		return ProfileImageResult{}, fmt.Errorf("%w: %v", ErrInvalidProfileImageReference, err)
	}
	deliveryURL, err := imagekitservice.BuildURL(asset.FilePath)
	if err != nil {
		return ProfileImageResult{}, fmt.Errorf("build profile-image URL: %w", err)
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return ProfileImageResult{}, fmt.Errorf("store profile-image reference: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			log.Printf("profile image: failed to roll back transaction: %v", rollbackErr)
		}
	}()

	queries := db.New(pool).WithTx(tx)
	current, err := store.GetPlayerByIdWithLock(ctx, queries)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ProfileImageResult{}, ErrProfileImagePlayerNotFound
		}
		return ProfileImageResult{}, fmt.Errorf("store profile-image reference: %w", err)
	}
	previousFileID := ""
	if current.ProfileImageFileID.Valid {
		previousFileID = current.ProfileImageFileID.String
	}

	if _, err := store.UpdatePlayerProfileImage(ctx, queries, asset.FileID, asset.FilePath); err != nil {
		return ProfileImageResult{}, fmt.Errorf("store profile-image reference: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return ProfileImageResult{}, fmt.Errorf("store profile-image reference: %w", err)
	}
	if previousFileID != "" && previousFileID != asset.FileID {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), profileImageCleanupTimeout)
		defer cleanupCancel()
		if err := imagekitservice.DeleteFile(cleanupCtx, previousFileID); err != nil {
			log.Printf("profile image: failed to delete replaced ImageKit file %s: %v", previousFileID, err)
		}
	}

	return ProfileImageResult{
		FileID:   asset.FileID,
		FilePath: asset.FilePath,
		URL:      deliveryURL,
	}, nil
}
