package sol

import (
	"crypto/sha1"
	"encoding/base64"
)

// WSAccept computes Sec-WebSocket-Accept from a client key (mission 047).
func WSAccept(key string) string {
	h := sha1.Sum([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(h[:])
}
