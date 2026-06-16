package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func Sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
