package playerlife

import "time"

type options struct {
	delay      time.Duration
	maxRetries int
}

func defaultOptions() *options {
	return &options{
		delay:      30 * time.Second,
		maxRetries: 3,
	}
}

// Option 配置玩家生命周期管理器。
type Option func(o *options)

// WithDelay 设置卸载玩家本地数据前的 Grace Period。
func WithDelay(delay time.Duration) Option {
	return func(o *options) { o.delay = delay }
}

// WithMaxRetries 设置清理重试次数：-1 表示不重试，0 表示一直重试。
func WithMaxRetries(n int) Option {
	return func(o *options) { o.maxRetries = n }
}
