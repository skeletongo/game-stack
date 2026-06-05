package clean

import (
	"time"
)

// options 模块配置。
type options struct {
	delay      time.Duration
	maxRetries int
}

// defaultOptions 返回默认配置（内存仓储）。
func defaultOptions() *options {
	return &options{
		delay:      time.Second * 30,
		maxRetries: 3,
	}
}

// Option 函数式选项。
type Option func(o *options)

// WithDelay 延迟清理时长
func WithDelay(delay time.Duration) Option {
	return func(o *options) { o.delay = delay }
}

// WithMaxRetries 重试次数
func WithMaxRetries(n int) Option {
	return func(o *options) { o.maxRetries = n }
}
