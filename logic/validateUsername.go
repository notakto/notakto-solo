package logic

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	MinUsernameLength = 3
	MaxUsernameLength = 30
)

// ValidateUsername checks if a username meets the required criteria:
// - Non-empty
// - Length between 3 and 30 characters
// - No whitespace
// - Contains only uppercase/lowercase English letters, digits, and underscores ('_')
func ValidateUsername(username string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}

	if strings.ContainsAny(username, " \t\n\r\v\f") {
		return errors.New("username cannot contain whitespace")
	}

	charCount := utf8.RuneCountInString(username)
	if charCount < MinUsernameLength {
		return fmt.Errorf("username is too short: must be at least %d characters", MinUsernameLength)
	}
	if charCount > MaxUsernameLength {
		return fmt.Errorf("username is too long: must not exceed %d characters", MaxUsernameLength)
	}

	for _, r := range username {
		if unicode.IsSpace(r) {
			return errors.New("username cannot contain whitespace")
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return fmt.Errorf("username contains invalid character '%c': only letters (a-z, A-Z), numbers (0-9), and underscores (_) are allowed", r)
	}

	return nil
}
