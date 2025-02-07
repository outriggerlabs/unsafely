package unsafely

import (
	"testing"
)

// Generic container with a single type parameter.
type container[T any] struct {
	value T
}

// Generic container with multiple type parameters.
type pair[K comparable, V any] struct {
	key   K
	value V
}

// Generic container with type constraints.
type numberContainer[T ~int | ~float64] struct {
	value T
	max   T
}

// Generic interface implementation.
type stack[T any] struct {
	items []T
}

func TestMarshalJSON_GenericTypes(t *testing.T) {
	type pointerContainer[T any] struct {
		value *T
		next  *pointerContainer[T]
	}

	tests := map[string]any{
		"basic generic container": container[string]{
			value: "test",
		},

		"multiple type parameters": pair[string, int]{
			key:   "answer",
			value: 42,
		},

		"constrained type parameters": numberContainer[float64]{
			value: 3.14,
			max:   10.0,
		},

		"generic interface implementation": &stack[int]{items: []int{1, 2, 3}},

		"nested generic types": container[pair[string, container[int]]]{
			value: pair[string, container[int]]{
				key: "nested",
				value: container[int]{
					value: 42,
				},
			},
		},

		"generic type with pointer fields": func() any {
			value := 42
			return pointerContainer[int]{
				value: &value,
				next: &pointerContainer[int]{
					value: &value,
				},
			}
		}(),
	}

	testMarshalJSON(t, tests)
}
