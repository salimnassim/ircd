package ircd

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func cloakHost(secret []byte, address string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(address))
	return hex.EncodeToString(mac.Sum(nil))[:12]
}
