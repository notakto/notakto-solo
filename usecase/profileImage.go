package usecase

import (
	"errors"
)

var (
	ErrProfileImagePlayerNotFound = errors.New("player profile not found")
	ErrInvalidProfileImageRequest = errors.New("invalid profile-image request")
)
