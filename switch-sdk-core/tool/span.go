package tool

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateSpanId 生成SpanID
// 生成一个16位的数
func GenerateSpanId() string {
	spanID := make([]byte, 8)
	if _, err := rand.Read(spanID); err != nil {
		return "0000000000000000"
	}
	return hex.EncodeToString(spanID)
}
