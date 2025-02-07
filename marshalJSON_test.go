package unsafely

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/outriggerlabs/unsafely/typeutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptrTo[T any](v T) *T {
	return &v
}

func TestMarshalJSON_Primitives(t *testing.T) {
	var tests = map[string]any{
		"int":         int(1),
		"int8":        int8(2),
		"int16":       int16(3),
		"int32":       int32(4),
		"int64":       int64(5),
		"uint":        uint(6),
		"uint8":       uint8(7),
		"uint16":      uint16(8),
		"uint32":      uint32(9),
		"uint64":      uint64(10),
		"uintptr":     uintptr(11),
		"float32":     float32(12.0),
		"float64":     float64(13.0),
		"json.Number": json.Number("14"),
		"string":      "fifteen",
		"complex64":   complex64(16 + 17i),
		"complex128":  complex128(18 + 19i),
	}

	testMarshalJSON(t, tests)
}

func TestMarshalJSON_Primitive_Pointers(t *testing.T) {
	var tests = map[string]any{
		"*int":         ptrTo(int(1)),
		"*int8":        ptrTo(int8(2)),
		"*int16":       ptrTo(int16(3)),
		"*int32":       ptrTo(int32(4)),
		"*int64":       ptrTo(int64(5)),
		"*uint":        ptrTo(uint(6)),
		"*uint8":       ptrTo(uint8(7)),
		"*uint16":      ptrTo(uint16(8)),
		"*uint32":      ptrTo(uint32(9)),
		"*uint64":      ptrTo(uint64(10)),
		"*uintptr":     ptrTo(uintptr(11)),
		"*float32":     ptrTo(float32(12.0)),
		"*float64":     ptrTo(float64(13.0)),
		"*json.Number": ptrTo(json.Number("14")),
		"*string":      ptrTo("fifteen"),
	}

	testMarshalJSON(t, tests)
}

func TestMarshalJSON_MorePointers(t *testing.T) {
	tests := map[string]any{
		"**int":     ptrTo(ptrTo(int(1))),
		"*int(nil)": (*int)(nil),
	}

	testMarshalJSON(t, tests)
}

func TestMarshalJSON_PointerSharing(t *testing.T) {
	var (
		a = ptrTo("one")
		b = []*string{a, a}
	)

	out, err := MarshalJSON(b)
	require.NoError(t, err)
	assert.JSONEq(t, `{"value":[{"pointer":1,"value":"one"},{"pointer":1,"value":"one"}]}`, string(out))

	var decoded []*string
	require.NoError(t, UnmarshalJSON(out, &decoded))
	require.Len(t, decoded, 2)
	assert.Same(t, decoded[0], decoded[1])
}

func TestMarshalJSON_ArrayLike(t *testing.T) {
	tests := map[string]any{
		"[]int64":      []int64{1, 2, 3},
		"[]int64{}":    []int64{},
		"[]int64(nil)": []int64(nil),
		"[3]int64":     [3]int64{1, 2, 3},
		"[3]int64{}":   [3]int64{},

		"[]*int64":  []*int64{ptrTo(int64(1)), ptrTo(int64(2)), ptrTo(int64(3))},
		"[3]*int64": [3]*int64{ptrTo(int64(1)), ptrTo(int64(2)), ptrTo(int64(3))},

		"*[]*int64":  ptrTo([]*int64{ptrTo(int64(1)), ptrTo(int64(2)), ptrTo(int64(3))}),
		"*[3]*int64": ptrTo([3]*int64{ptrTo(int64(1)), ptrTo(int64(2)), ptrTo(int64(3))}),

		"[]*int64(nil)": []*int64{nil, ptrTo(int64(2)), nil},
	}

	testMarshalJSON(t, tests)
}

func TestMarshalJSON_Unsupported(t *testing.T) {
	tests := map[string]any{
		"chan":           make(chan struct{}),
		"func":           func() {},
		"unsafe.Pointer": reflect.New(reflect.TypeOf(1)).UnsafePointer(),
	}

	for desc, value := range tests {
		t.Run(desc, func(t *testing.T) {
			_, err := MarshalJSON(value)
			require.Error(t, err)
			require.Contains(t, err.Error(), "unsupported kind")
		})
	}
}

func testMarshalJSON[T any](t *testing.T, tests map[string]T) {
	for desc, value := range tests {
		t.Run(desc, func(t *testing.T) {
			b, err := MarshalJSON(value, WithIndent("  "))
			require.NoError(t, err)
			// fmt.Println("MarshalJSON():", string(b))

			decodedPtrV := reflect.New(reflect.TypeOf(value))
			require.NoError(t,
				UnmarshalJSON(b, decodedPtrV.Interface(), WithTypeResolver(typeutil.UnsafeResolver())),
			)
			require.Equal(t, value, decodedPtrV.Elem().Interface())
		})
	}
}
