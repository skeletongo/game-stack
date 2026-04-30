package room

type options struct {
	store      Store
	maxPlayers int32
}

func defaultOptions() *options {
	return &options{
		store:      newMemoryStore(),
		maxPlayers: 20,
	}
}

type Option func(o *options)

func WithStore(s Store) Option      { return func(o *options) { o.store = s } }
func WithMaxPlayers(n int32) Option { return func(o *options) { o.maxPlayers = n } }
