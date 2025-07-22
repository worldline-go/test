package utils

import "testing"

func ErrCheck[T any](v T, err error) func(t *testing.T) T {
	return func(t *testing.T) T {
		t.Helper()

		if err != nil {
			t.Fatal(err)
		}

		return v
	}
}
