package migration

type options struct {
	debug       bool
	batchSize   int
	readFields  []string
	writeFields []string
}

type option func(*options)

func Debug(debug bool) option {
	return func(o *options) {
		o.debug = debug
	}
}

func BatchSize(batchSize int) option {
	return func(o *options) {
		o.batchSize = batchSize
	}
}

func Read(fields ...string) option {
	return func(o *options) {
		o.readFields = fields
	}
}

func Write(fields ...string) option {
	return func(o *options) {
		o.writeFields = fields
	}
}
