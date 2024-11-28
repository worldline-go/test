package test

import (
	"context"
	"log"
	"os"
	"testing"
)

// Main is a wrapper around testing.M.Run that allows for setup and teardown with function.
//   - function before to run and return a defer function and error. Defer for cleanup after the tests.
//   - return nil in the function if doesn't have any cleanup.
func Main(m *testing.M, fn func(ctx context.Context) (func(), error), opts ...Option) {
	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()

	opt := apply(opts)

	deferFn, err := fn(opt.ctx)
	if err != nil {
		log.Println(err.Error())
		exitCode = 1
		return
	}

	if deferFn != nil {
		defer deferFn()
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
//   - default is context.Background()
func WithContext(ctx context.Context) Option {
	return func(o *option) {
		o.ctx = ctx
	}
}
