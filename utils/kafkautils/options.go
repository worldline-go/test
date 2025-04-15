package kafkautils

import (
	"github.com/worldline-go/wkafka"
)

type option struct {
	WkafkaOpts []wkafka.Option
}

func (o *option) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

type Option func(*option)

func WithWafkaOptions(opts ...wkafka.Option) Option {
	return func(o *option) {
		o.WkafkaOpts = append(o.WkafkaOpts, opts...)
	}
}
