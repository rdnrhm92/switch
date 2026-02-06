package tool

import (
	"math/rand"
	"time"
)

// GetRandomBetween 取随机数 设置缓存过期时间
func GetRandomBetween(min, max int) int {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	if min > max {
		min, max = max, min
	}
	return min + r.Intn(max-min+1)
}
