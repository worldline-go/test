package test

import (
	"os"
	"testing"
)

// Main is a wrapper around testing.M.Run that allows for setup and teardown with function.
//   - function before to run and return a defer function and error. Defer for cleanup after the tests.
//   - return nil in the function if doesn't have any cleanup.
func Main(m *testing.M, fn func() func(), opts ...Option) {
	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()

	deferFn := fn()

	if deferFn != nil {
		defer deferFn()
	}

	exitCode = m.Run()
}

// ///////////////////////////////////////////////////////////////////////////

type Option func(*option)

type option struct{}

func (o *option) Default() {}

func apply(opts []Option) *option {
	opt := new(option)
	for _, o := range opts {
		o(opt)
	}

	opt.Default()

	return opt
}
