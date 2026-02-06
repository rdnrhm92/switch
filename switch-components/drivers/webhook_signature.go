package drivers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"gitee.com/fatzeng/switch-sdk-core/logger"
)

// GenerateSignature 生成HMAC签名
func GenerateSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("sha256=%s", signature)
}

// GetSignature 获取请求头中的Signature
func GetSignature(header map[string][]string) string {
	if header == nil {
		return ""
	}
	if len(header["X-Webhook-Signature"]) == 0 {
		return ""
	}
	return header["X-Webhook-Signature"][0]
}

// BuildSignature 构建Signature
func BuildSignature(secret string, payload []byte) (string, string) {
	return "X-Webhook-Signature", GenerateSignature(secret, payload)
}

// ValidateHMACSignature 使用HMAC校验签名
func ValidateHMACSignature(signature string, body []byte, secret string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		logger.Logger.Warn("Invalid HMAC signature format, expected 'sha256=' prefix")
		return false
	}
	//获取签名
	expectedSignature := strings.TrimPrefix(signature, "sha256=")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	computedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSignature), []byte(computedSignature))
}
