package contextkey

import "context"

type uidKey struct{}

var UID uidKey = struct{}{}

// UIDFromContext returns the Firebase UID from ctx if set.
func UIDFromContext(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(UID).(string)
	return uid, ok
}
