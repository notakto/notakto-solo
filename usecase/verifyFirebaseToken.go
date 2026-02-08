package usecase

import (
	"context"
	"errors"

	"firebase.google.com/go/v4/auth"

	"github.com/rakshitg600/notakto-solo/contextkey"
)

func VerifyFirebaseToken(ctx context.Context, authClient *auth.Client, idToken string) (string, error) {
	token, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", err
	}
	return token.UID, nil
}

func GetFirebaseUserProfile(ctx context.Context, authClient *auth.Client) (name string, email string, photo string, err error) {
	uid, ok := contextkey.UIDFromContext(ctx)
	if !ok || uid == "" {
		return "", "", "", errors.New("missing or invalid uid in context")
	}
	u, err := authClient.GetUser(ctx, uid)
	if err != nil {
		return "", "", "", err
	}
	return u.DisplayName, u.Email, u.PhotoURL, nil
}
