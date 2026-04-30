package inventory

type options struct {
	store    Store
	bagSlots int32
}

func defaultOptions() *options {
	return &options{
		store:    newMemoryStore(),
		bagSlots: 100,
	}
}

type Option func(o *options)

func WithStore(s Store) Option    { return func(o *options) { o.store = s } }
func WithBagSlots(n int32) Option { return func(o *options) { o.bagSlots = n } }
