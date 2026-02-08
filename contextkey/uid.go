package contextkey

import "context"

// uidKey is a dedicated type for the Firebase UID context key to avoid collisions.
type uidKey struct{}

// UID is the context key under which the Firebase UID is stored by the auth middleware.
var UID uidKey = struct{}{}

// UIDFromContext returns the Firebase UID from ctx if set. Otherwise returns "", false.
func UIDFromContext(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(UID).(string)
	return uid, ok
}
