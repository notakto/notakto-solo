package coinbase

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func VerifyWebhookSignature(secret string, signature string, payload []byte) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
