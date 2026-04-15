package nowpayments

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
)

// VerifyIPNSignature checks the x-nowpayments-sig header for an IPN callback.
//
// NOWPayments signs the payload by: (1) recursively sorting JSON object keys
// alphabetically, (2) re-serializing with no whitespace, (3) HMAC-SHA512 with
// the IPN secret, (4) hex-encoding the digest. We reproduce that canonical
// form here. Go's encoding/json marshals map[string]interface{} with sorted
// keys by default, so decoding into interface{} and re-encoding yields the
// same canonical bytes that NOWPayments signed. HTML escaping is disabled so
// characters like < > & round-trip unchanged (matching JSON.stringify).
func VerifyIPNSignature(secret string, signature string, payload []byte) bool {
	if signature == "" || secret == "" {
		return false
	}

	canonical, err := canonicalJSON(payload)
	if err != nil {
		return false
	}

	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(canonical)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}

func canonicalJSON(payload []byte) ([]byte, error) {
	var parsed interface{}
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.UseNumber()
	if err := dec.Decode(&parsed); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(parsed); err != nil {
		return nil, err
	}

	out := buf.Bytes()
	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	return out, nil
}
