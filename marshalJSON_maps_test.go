package unsafely

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type point struct {
	x int
	y int
}

// Note: Non-primitive keys are not supported by JSON and are marshalled and
// converted to strings.
func TestMarshalJSON_Maps(t *testing.T) {
	tests := map[string]any{
		"nil map": map[string]string(nil),

		"map[string]struct{}": map[string]struct{}{"a": {}, "b": {}, "c": {}},

		"map[string]int":  map[string]int{"a": 1, "b": 2, "c": 3},
		"map[string]*int": map[string]*int{"a": ptrTo(int(1)), "b": ptrTo(int(2)), "c": ptrTo(int(3))},
		"*map[string]int": ptrTo(map[string]int{"a": 1, "b": 2, "c": 3}),

		"map[point]int":     map[point]int{{1, 2}: 1, {2, 3}: 2, {3, 4}: 3},
		"map[[2]int]string": map[[2]int]string{{1, 2}: "a", {2, 3}: "b", {3, 4}: "c"},

		"map[int]point":  map[int]point{1: {1, 2}, 2: {2, 3}, 3: {3, 4}},
		"map[int]*point": map[int]*point{1: ptrTo(point{1, 2}), 2: ptrTo(point{2, 3}), 3: ptrTo(point{3, 4})},

		"map[int][]point": map[int][]point{
			1: {{1, 2}, {1, 3}},
			2: {{2, 3}, {2, 4}},
			3: {{3, 4}, {3, 5}},
		},
		"map[int][]*point": map[int][]*point{
			1: {ptrTo(point{1, 2}), ptrTo(point{1, 3})},
			2: {ptrTo(point{2, 3}), ptrTo(point{2, 4})},
			3: {ptrTo(point{3, 4}), ptrTo(point{3, 5})},
		},

		"map[int][2]point": map[int][2]point{
			1: {{1, 2}, {1, 3}},
			2: {{2, 3}, {2, 4}},
			3: {{3, 4}, {3, 5}},
		},
		"map[int][2]*point": map[int][2]*point{
			1: {ptrTo(point{1, 2}), ptrTo(point{1, 3})},
			2: {ptrTo(point{2, 3}), ptrTo(point{2, 4})},
			3: {ptrTo(point{3, 4}), ptrTo(point{3, 5})},
		},
	}

	testMarshalJSON(t, tests)
}

func TestMarshalJSON_MapsWithPointerKeys(t *testing.T) {
	t.Run("map[*int]string", func(t *testing.T) {
		testMapWithPointerKey(t, map[*int]string{
			ptrTo(int(1)): "a",
			ptrTo(int(2)): "b",
			ptrTo(int(3)): "c",
		})
	})

	t.Run("map[*point]string", func(t *testing.T) {
		testMapWithPointerKey(t, map[*point]string{
			ptrTo(point{1, 2}): "a",
			ptrTo(point{2, 3}): "b",
			ptrTo(point{3, 4}): "c",
		})
	})
}

// Map with pointer keys won't be equal after unmarshaling, so this includes a
// conversion function that dereferences the key pointers before testing
// equality.
func testMapWithPointerKey[K *T, T comparable, V any](t *testing.T, input map[K]V) {
	b, err := MarshalJSON(input)
	require.NoError(t, err)

	inputCopy := make(map[K]V)
	require.NoError(t, UnmarshalJSON(b, &inputCopy))

	convert := func(in map[K]V) map[T]V {
		out := make(map[T]V)
		for k, v := range in {
			out[*k] = v
		}
		return out
	}

	assert.Equal(t, convert(input), convert(inputCopy))
}
