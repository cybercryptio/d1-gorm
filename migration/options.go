package migration

const defaultBatchSize = 10

type options struct {
	batchSize int
}

type option func(*options)

func defaultOptions() options {
	return options{
		batchSize: defaultBatchSize,
	}
}

func (o *options) apply(opts ...option) {
	for _, opt := range opts {
		opt(o)
	}
}

func BatchSize(batchSize int) option {
	return func(o *options) {
		o.batchSize = batchSize
	}
}
