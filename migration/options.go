// Copyright 2022 CYBERCRYPT
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License

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
