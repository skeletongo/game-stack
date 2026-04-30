package chat

type options struct {
	store       Store
	maxMsgLen   int32
	rateLimit   int32 // 每分钟最大发言数
	historySize int32
}

func defaultOptions() *options {
	return &options{
		store:       newMemoryStore(),
		maxMsgLen:   200,
		rateLimit:   10,
		historySize: 100,
	}
}

type Option func(o *options)

func WithStore(s Store) Option     { return func(o *options) { o.store = s } }
func WithMaxMsgLen(n int32) Option { return func(o *options) { o.maxMsgLen = n } }
func WithRateLimit(n int32) Option { return func(o *options) { o.rateLimit = n } }
