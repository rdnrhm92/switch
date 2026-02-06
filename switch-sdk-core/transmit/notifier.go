package transmit

import (
	"context"
)

// Notifier 推送消息
type Notifier interface {
	Notify(ctx context.Context, data interface{}) error
}
