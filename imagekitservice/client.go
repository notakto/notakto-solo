package imagekitservice

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	imagekitsdk "github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
)

const (
	UploadURL            = "https://upload.imagekit.io/api/v2/files/upload"
	UploadAuthTTL        = 5 * time.Minute
	MaxFileSizeBytes     = int64(5 * 1024 * 1024)
	MaxImageDimension    = 4096
	UploadChecks         = "'file.mime' IN ['image/jpeg', 'image/png', 'image/webp'] AND 'file.size' <= 5242880 AND 'mediaMetadata.width' <= 4096 AND 'mediaMetadata.height' <= 4096"
	profileImageRoot     = "/profile-images"
	maxOriginalNameBytes = 255
)

var (
	ErrInvalidConfig        = errors.New("invalid ImageKit configuration")
	ErrNotInitialized       = errors.New("ImageKit service is not initialized")
	ErrInvalidUID           = errors.New("invalid Firebase UID")
	ErrInvalidFilename      = errors.New("invalid image filename")
	ErrUnsupportedExtension = errors.New("unsupported image extension")
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

func containsControl(value string) bool {
	return strings.IndexFunc(value, unicode.IsControl) >= 0
}
