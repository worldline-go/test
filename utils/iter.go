package utils

import (
	"iter"
	"reflect"
	"testing"
)

type IterKV[K, V any] struct {
	K K
	V V
}

func Iter2[K, V any](values ...IterKV[K, V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, value := range values {
			if !yield(value.K, value.V) {
				return
			}
		}
	}
}

func Iter2Check[K, V any](t *testing.T, its iter.Seq2[K, V], values []IterKV[K, V]) {
	t.Helper()

	if its == nil {
		t.Fatal("iter.Seq2 is nil")
	}

	for k, v := range its {
		for _, value := range values {
			if reflect.DeepEqual(value.K, k) && reflect.DeepEqual(value.V, v) {
				return
			}
		}
		t.Errorf("unexpected value: %#v, %#v", k, v)
	}
}
