package auth

// options 认证模块配置。
type options struct {
	store Store
}

func defaultOptions() *options {
	return &options{
		store: newMemoryStore(),
	}
}

// Option 函数式选项。
type Option func(o *options)

// WithStore 设置数据存储实现。
func WithStore(s Store) Option {
	return func(o *options) { o.store = s }
}
