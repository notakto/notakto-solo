package imagekitservice

import (
	"encoding/base64"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"
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

type Client struct {
	sdk         *imagekitsdk.Client
	publicKey   string
	privateKey  string
	urlEndpoint string
}

// NewClient creates an initialized ImageKit client.
func NewClient(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.PublicKey) == "" {
		return nil, errors.New("ImageKit public key is required")
	}
	if strings.TrimSpace(cfg.PrivateKey) == "" {
		return nil, errors.New("ImageKit private key is required")
	}

	imagekit := imagekitsdk.NewClient(option.WithPrivateKey(cfg.PrivateKey))
	return &Client{
		sdk:         &imagekit,
		publicKey:   cfg.PublicKey,
		privateKey:  cfg.PrivateKey,
		urlEndpoint: strings.TrimRight(cfg.URLEndpoint, "/"),
	}, nil
}

func (c *Client) GenerateUploadAuth(uid, originalFilename string) (UploadAuth, error) {
	if uid == "" || strings.TrimSpace(uid) == "" || !utf8.ValidString(uid) {
		return UploadAuth{}, ErrInvalidUID
	}
	folder := profileImageRoot + "/" + base64.RawURLEncoding.EncodeToString([]byte(uid))
	if originalFilename == "" || len(originalFilename) > maxOriginalNameBytes ||
		strings.TrimSpace(originalFilename) != originalFilename || !utf8.ValidString(originalFilename) ||
		strings.ContainsAny(originalFilename, `/\`) {
		return UploadAuth{}, ErrInvalidFilename
	}
	extension := strings.ToLower(path.Ext(originalFilename))
	stem := strings.TrimSuffix(originalFilename, path.Ext(originalFilename))
	if stem == "" || stem == "." {
		return UploadAuth{}, ErrInvalidFilename
	}
	switch extension {
	case ".jpg":
		extension = ".jpg"
	case ".jpeg":
		extension = ".jpg"
	case ".png":
		extension = ".png"
	case ".webp":
		extension = ".webp"
	default:
		return UploadAuth{}, ErrUnsupportedExtension
	}

	fileName := "avatar-" + uuid.NewString() + extension
	issuedAt := time.Now().UTC().Unix()
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
	token.Header["kid"] = c.publicKey
	signed, err := token.SignedString([]byte(c.privateKey))
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
