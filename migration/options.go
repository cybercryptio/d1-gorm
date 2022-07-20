package migration

const defaultBatchSize = 10

type options struct {
	batchSize int
}

// Option is used to configure optional settings for the migration.
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

// WithBatchSize sets the batch size to be used when migrating data. The default batch size is 10.
func WithBatchSize(batchSize int) Option {
	return func(o *options) {
		o.batchSize = batchSize
	}
}
