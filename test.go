package test

import (
	"context"
	"log"
	"os"
	"testing"
)

func Main(m *testing.M, fn func(ctx context.Context) error, opts ...Option) {
	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()

	opt := apply(opts)

	if err := fn(opt.ctx); err != nil {
		log.Println(err.Error())
		exitCode = 1
		return
	}

	exitCode = m.Run()
}

// ///////////////////////////////////////////////////////////////////////////

type Option func(*option)

type option struct {
	ctx context.Context
}

func (o *option) Default() {
	if o.ctx == nil {
		o.ctx = context.Background()
	}
}

func apply(opts []Option) *option {
	opt := new(option)
	for _, o := range opts {
		o(opt)
	}

	opt.Default()

	return opt
}

// WithContext sets the context for the main function.
func WithContext(ctx context.Context) Option {
	return func(o *option) {
		o.ctx = ctx
	}
}
