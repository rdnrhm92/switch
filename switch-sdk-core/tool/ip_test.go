package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocalIP(t *testing.T) {
	// 测试获取本地IP
	ip, err := GetLocalIP()
	assert.NoError(t, err)
	assert.NotEmpty(t, ip)

	// 验证返回的IP格式是否正确
	assert.Regexp(t, `^(\d{1,3}\.){3}\d{1,3}$`, ip)
}
