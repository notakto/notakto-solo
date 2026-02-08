package usecase

import (
	"context"

	"firebase.google.com/go/v4/auth"
)

// VerifyFirebaseToken validates a Firebase ID token using the Admin SDK.
// It returns the user's UID on success, or a non-nil error on failure.
func VerifyFirebaseToken(ctx context.Context, authClient *auth.Client, idToken string) (string, error) {
	token, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", err
	}
	return token.UID, nil
}

// GetFirebaseUserProfile fetches the profile data for a Firebase user by UID.
// It returns the display name, email, and photo URL from the Firebase user record.
func GetFirebaseUserProfile(ctx context.Context, authClient *auth.Client, uid string) (name string, email string, photo string, err error) {
	u, err := authClient.GetUser(ctx, uid)
	if err != nil {
		return "", "", "", err
	}
	return u.DisplayName, u.Email, u.PhotoURL, nil
}
