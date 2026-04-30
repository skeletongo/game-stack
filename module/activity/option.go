package activity

type options struct{ store Store }

func defaultOptions() *options { return &options{store: newMemoryStore()} }

type Option func(o *options)

func WithStore(s Store) Option { return func(o *options) { o.store = s } }
