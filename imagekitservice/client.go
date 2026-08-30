package imagekitservice

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	imagekitsdk "github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
	"github.com/imagekit-developer/imagekit-go/v2/shared"
)

const (
	UploadURL            = "https://upload.imagekit.io/api/v2/files/upload"
	UploadAuthTTL        = 5 * time.Minute
	MaxFileSizeBytes     = int64(5 * 1024 * 1024)
	MaxImageDimension    = 4096
	UploadChecks         = "'file.mime' IN ['image/jpeg', 'image/png', 'image/webp'] AND 'file.size' <= 5242880 AND 'mediaMetadata.width' <= 4096 AND 'mediaMetadata.height' <= 4096"
	profileImageRoot     = "/profile-images"
	maxOriginalNameBytes = 255
	maxFileIDBytes       = 256
	maxAssetPathBytes    = 1024
)

var (
	ErrInvalidConfig        = errors.New("invalid ImageKit configuration")
	ErrNotInitialized       = errors.New("ImageKit service is not initialized")
	ErrInvalidUID           = errors.New("invalid Firebase UID")
	ErrInvalidFilename      = errors.New("invalid image filename")
	ErrUnsupportedExtension = errors.New("unsupported image extension")
	ErrInvalidFileID        = errors.New("invalid ImageKit file ID")
	ErrInvalidAsset         = errors.New("invalid ImageKit profile asset")
	ErrInvalidAssetPath     = errors.New("invalid ImageKit asset path")
	ErrAssetPathMismatch    = errors.New("ImageKit asset path mismatch")
)

var (
	fileIDPattern           = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	assetPathPattern        = regexp.MustCompile(`^/[A-Za-z0-9._/-]+$`)
	profileImageNamePattern = regexp.MustCompile(`^avatar-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\.(jpg|png|webp)$`)
)

// Config contains the server-only ImageKit credentials and public CDN endpoint.
type Config struct {
	PublicKey   string
	PrivateKey  string
	URLEndpoint string
}

// UploadPayload is the complete set of non-file multipart fields signed into a
// V2 client-upload JWT. Values are strings because ImageKit requires upload
// parameters to be stringified in V2 JWT claims and multipart form fields.
type UploadPayload struct {
	FileName          string `json:"fileName"`
	Folder            string `json:"folder"`
	UseUniqueFileName string `json:"useUniqueFileName"`
	OverwriteFile     string `json:"overwriteFile"`
	IsPrivateFile     string `json:"isPrivateFile"`
	IsPublished       string `json:"isPublished"`
	Checks            string `json:"checks"`
}

// UploadAuth contains the short-lived credentials and immutable fields needed
// for a direct browser upload to ImageKit Upload API V2.
type UploadAuth struct {
	Token         string        `json:"token"`
	ExpiresAt     int64         `json:"expiresAt"`
	UploadURL     string        `json:"uploadUrl"`
	UploadPayload UploadPayload `json:"uploadPayload"`
}

// ProfileAsset is the authoritative subset of ImageKit file metadata needed to
// decide whether an uploaded file may be associated with a user profile.
type ProfileAsset struct {
	FileID        string
	FilePath      string
	Type          string
	FileType      string
	Mime          string
	Size          float64
	Width         float64
	Height        float64
	IsPrivateFile bool
	IsPublished   bool
}

var (
	sdk         *imagekitsdk.Client
	publicKey   string
	privateKey  string
	urlEndpoint string
	now         = time.Now
	newID       = uuid.NewString
)

// Init validates cfg and initializes the ImageKit service.
func Init(cfg Config) error {
	normalized, err := validateConfig(cfg)
	if err != nil {
		return err
	}

	imagekit := imagekitsdk.NewClient(option.WithPrivateKey(normalized.PrivateKey))
	sdk = &imagekit
	publicKey = normalized.PublicKey
	privateKey = normalized.PrivateKey
	urlEndpoint = normalized.URLEndpoint
	return nil
}

func ensureInitialized() error {
	if sdk == nil {
		return ErrNotInitialized
	}
	return nil
}

// ProfileImageFolder returns the only ImageKit folder authorized for uid. Raw
// URL-safe base64 avoids collisions and prevents UID path traversal.
func ProfileImageFolder(uid string) (string, error) {
	if uid == "" || strings.TrimSpace(uid) == "" || !utf8.ValidString(uid) || containsControl(uid) {
		return "", ErrInvalidUID
	}
	return profileImageRoot + "/" + base64.RawURLEncoding.EncodeToString([]byte(uid)), nil
}

func GenerateUploadAuth(uid, originalFilename string) (UploadAuth, error) {
	if err := ensureInitialized(); err != nil {
		return UploadAuth{}, err
	}
	folder, err := ProfileImageFolder(uid)
	if err != nil {
		return UploadAuth{}, err
	}
	extension, err := normalizeExtension(originalFilename)
	if err != nil {
		return UploadAuth{}, err
	}

	fileName := "avatar-" + newID() + extension
	issuedAt := now().UTC().Unix()
	expiresAt := issuedAt + int64(UploadAuthTTL/time.Second)
	claims := jwt.MapClaims{
		"fileName":          fileName,
		"folder":            folder,
		"useUniqueFileName": "false",
		"overwriteFile":     "false",
		"isPrivateFile":     "false",
		"isPublished":       "true",
		"checks":            UploadChecks,
		"iat":               issuedAt,
		"exp":               expiresAt,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = publicKey
	signed, err := token.SignedString([]byte(privateKey))
	if err != nil {
		return UploadAuth{}, fmt.Errorf("sign ImageKit upload token: %w", err)
	}

	return UploadAuth{
		Token:     signed,
		ExpiresAt: expiresAt,
		UploadURL: UploadURL,
		UploadPayload: UploadPayload{
			FileName:          fileName,
			Folder:            folder,
			UseUniqueFileName: "false",
			OverwriteFile:     "false",
			IsPrivateFile:     "false",
			IsPublished:       "true",
			Checks:            UploadChecks,
		},
	}, nil
}

func GetFile(ctx context.Context, fileID string) (ProfileAsset, error) {
	if err := ensureInitialized(); err != nil {
		return ProfileAsset{}, err
	}
	if err := validateFileID(fileID); err != nil {
		return ProfileAsset{}, err
	}
	file, err := sdk.Files.Get(ctx, fileID)
	if err != nil {
		return ProfileAsset{}, fmt.Errorf("get ImageKit file: %w", err)
	}
	if file == nil {
		return ProfileAsset{}, fmt.Errorf("%w: ImageKit returned no file", ErrInvalidAsset)
	}
	if file.FileID != fileID {
		return ProfileAsset{}, fmt.Errorf("%w: returned file ID does not match requested file ID", ErrInvalidAsset)
	}

	return ProfileAsset{
		FileID:        file.FileID,
		FilePath:      file.FilePath,
		Type:          string(file.Type),
		FileType:      file.FileType,
		Mime:          file.Mime,
		Size:          file.Size,
		Width:         file.Width,
		Height:        file.Height,
		IsPrivateFile: file.IsPrivateFile,
		IsPublished:   file.IsPublished,
	}, nil
}

func DeleteFile(ctx context.Context, fileID string) error {
	if err := ensureInitialized(); err != nil {
		return err
	}
	if err := validateFileID(fileID); err != nil {
		return err
	}
	if err := sdk.Files.Delete(ctx, fileID); err != nil {
		return fmt.Errorf("delete ImageKit file: %w", err)
	}
	return nil
}

func BuildURL(filePath string) (string, error) {
	if err := ensureInitialized(); err != nil {
		return "", err
	}
	if err := validateAssetPath(filePath); err != nil {
		return "", err
	}
	built := sdk.Helper.BuildURL(shared.SrcOptionsParam{
		Src:         filePath,
		URLEndpoint: urlEndpoint,
	})
	if built == "" {
		return "", fmt.Errorf("%w: could not build URL", ErrInvalidAssetPath)
	}
	return built, nil
}

func ValidateAsset(asset ProfileAsset, uid, optionalPath string) error {
	if err := ensureInitialized(); err != nil {
		return err
	}
	if err := validateFileID(asset.FileID); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidAsset, err)
	}
	expectedFolder, err := ProfileImageFolder(uid)
	if err != nil {
		return err
	}
	if err := validateAssetPath(asset.FilePath); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidAsset, err)
	}
	if optionalPath != "" && optionalPath != asset.FilePath {
		return fmt.Errorf("%w: %w", ErrInvalidAsset, ErrAssetPathMismatch)
	}
	if path.Dir(asset.FilePath) != expectedFolder {
		return fmt.Errorf("%w: file is outside the user's profile-image folder", ErrInvalidAsset)
	}
	if !profileImageNamePattern.MatchString(path.Base(asset.FilePath)) {
		return fmt.Errorf("%w: filename was not generated by the profile-image upload flow", ErrInvalidAsset)
	}
	if asset.Type != "file" {
		return fmt.Errorf("%w: asset is not a regular file", ErrInvalidAsset)
	}
	if asset.FileType != "image" {
		return fmt.Errorf("%w: file is not an image", ErrInvalidAsset)
	}
	if asset.IsPrivateFile || !asset.IsPublished {
		return fmt.Errorf("%w: image must be public and published", ErrInvalidAsset)
	}

	wantExtension, ok := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}[asset.Mime]
	if !ok || strings.ToLower(path.Ext(asset.FilePath)) != wantExtension {
		return fmt.Errorf("%w: MIME type and filename extension do not match", ErrInvalidAsset)
	}
	if !isWholeNumberInRange(asset.Size, 1, float64(MaxFileSizeBytes)) {
		return fmt.Errorf("%w: file size is outside the allowed range", ErrInvalidAsset)
	}
	if !isWholeNumberInRange(asset.Width, 1, MaxImageDimension) ||
		!isWholeNumberInRange(asset.Height, 1, MaxImageDimension) {
		return fmt.Errorf("%w: image dimensions are outside the allowed range", ErrInvalidAsset)
	}
	return nil
}

func validateConfig(cfg Config) (Config, error) {
	if !validKey(cfg.PublicKey, "public_") {
		return Config{}, fmt.Errorf("%w: public key must start with public_", ErrInvalidConfig)
	}
	if !validKey(cfg.PrivateKey, "private_") {
		return Config{}, fmt.Errorf("%w: private key must start with private_", ErrInvalidConfig)
	}
	if cfg.URLEndpoint == "" || strings.TrimSpace(cfg.URLEndpoint) != cfg.URLEndpoint || containsControl(cfg.URLEndpoint) {
		return Config{}, fmt.Errorf("%w: URL endpoint is required", ErrInvalidConfig)
	}
	parsed, err := url.Parse(cfg.URLEndpoint)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil ||
		parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || parsed.Opaque != "" {
		return Config{}, fmt.Errorf("%w: URL endpoint must be an HTTPS origin or path without credentials, query, or fragment", ErrInvalidConfig)
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	cfg.URLEndpoint = strings.TrimRight(parsed.String(), "/")
	return cfg, nil
}

func validKey(value, prefix string) bool {
	return strings.HasPrefix(value, prefix) && len(value) > len(prefix) &&
		strings.TrimSpace(value) == value && !containsControl(value)
}

func normalizeExtension(originalFilename string) (string, error) {
	if originalFilename == "" || len(originalFilename) > maxOriginalNameBytes ||
		strings.TrimSpace(originalFilename) != originalFilename || !utf8.ValidString(originalFilename) ||
		containsControl(originalFilename) || strings.ContainsAny(originalFilename, `/\`) {
		return "", ErrInvalidFilename
	}
	extension := strings.ToLower(path.Ext(originalFilename))
	stem := strings.TrimSuffix(originalFilename, path.Ext(originalFilename))
	if stem == "" || stem == "." {
		return "", ErrInvalidFilename
	}
	switch extension {
	case ".jpg":
		return ".jpg", nil
	case ".jpeg":
		return ".jpg", nil
	case ".png":
		return ".png", nil
	case ".webp":
		return ".webp", nil
	default:
		return "", ErrUnsupportedExtension
	}
}

func validateFileID(fileID string) error {
	if fileID == "" || len(fileID) > maxFileIDBytes || !fileIDPattern.MatchString(fileID) {
		return ErrInvalidFileID
	}
	return nil
}

func validateAssetPath(filePath string) error {
	if filePath == "" || len(filePath) > maxAssetPathBytes || !assetPathPattern.MatchString(filePath) ||
		strings.HasPrefix(filePath, "//") || path.Clean(filePath) != filePath || path.Base(filePath) == "." {
		return ErrInvalidAssetPath
	}
	return nil
}

func isWholeNumberInRange(value float64, min, max float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && math.Trunc(value) == value && value >= min && value <= max
}

func containsControl(value string) bool {
	return strings.IndexFunc(value, unicode.IsControl) >= 0
}
