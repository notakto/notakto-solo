package functions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rakshitg600/notakto-solo/config"
)

type FirebaseTokenInfo struct {
	LocalID string `json:"localId"`
	Email   string `json:"email,omitempty"`
	Name    string `json:"displayName,omitempty"`
	Photo   string `json:"photoUrl,omitempty"`
}

var firebaseHTTPClient = &http.Client{
	Timeout: 5 * time.Second,
}

// VerifyFirebaseToken validates a Firebase ID token and looks up the associated user account.
// On success it returns the user's LocalID, display name, email, and photo URL.
// On failure it returns empty strings and a non-nil error describing the problem.
func VerifyFirebaseToken(ctx context.Context, idToken string) (string, string, string, string, error) {
	url := fmt.Sprintf(
		"https://identitytoolkit.googleapis.com/v1/accounts:lookup?key=%s",
		config.MustGetEnv("FIREBASE_API_KEY"),
	)

	payload := map[string]interface{}{
		"idToken": idToken,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", "", "", "", err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", "", "", "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := firebaseHTTPClient.Do(req)
	if err != nil {
		return "", "", "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", "", fmt.Errorf("firebase API returned status %d", resp.StatusCode)
	}

	var result struct {
		Users []FirebaseTokenInfo `json:"users"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", "", "", err
	}

	if len(result.Users) == 0 {
		return "", "", "", "", fmt.Errorf("no user found")
	}

	user := result.Users[0]
	return user.LocalID, user.Name, user.Email, user.Photo, nil
}