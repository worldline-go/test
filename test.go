package test

import (
	"os"
	"testing"
)

// Main is a wrapper around testing.M.Run that allows for setup and teardown with function.
//   - function before to run and return a defer function and error. Defer for cleanup after the tests.
//   - return nil in the function if doesn't have any cleanup.
func Main(m *testing.M, fn func() func()) {
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
