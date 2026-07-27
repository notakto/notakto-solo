package contextkey

import "context"

type keepaliveAuthorizedKey struct{}

var KeepaliveAuthorized keepaliveAuthorizedKey = struct{}{}

// KeepaliveAuthorizedFromContext returns the keepalive authorization result if set.
func KeepaliveAuthorizedFromContext(ctx context.Context) (bool, bool) {
	authorized, ok := ctx.Value(KeepaliveAuthorized).(bool)
	return authorized, ok
}
