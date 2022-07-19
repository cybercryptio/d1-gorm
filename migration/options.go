package migration

const defaultBatchSize = 10

type options struct {
	batchSize int
}

type Option func(*options)

func defaultOptions() options {
	return options{
		batchSize: defaultBatchSize,
	}
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithBatchSize(batchSize int) Option {
	return func(o *options) {
		o.batchSize = batchSize
	}
}
