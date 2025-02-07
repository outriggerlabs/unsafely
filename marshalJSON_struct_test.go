package unsafely

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/outriggerlabs/unsafely/typeutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Struct to verify how we handle various types.
type structType struct {
	int        int
	int8       int8
	int16      int16
	int32      int32
	int64      int64
	uint       uint
	uint8      uint8
	uint16     uint16
	uint32     uint32
	uint64     uint64
	uintptr    uintptr
	float32    float32
	float64    float64
	jsonNumber json.Number
	string     string

	complex64  complex64
	complex128 complex128

	intPtr    *int
	intPtrPtr **int

	intPtrNull *int

	intSlice    []int
	intPtrSlice []*int

	intArray    [3]int
	intPtrArray [3]*int

	intSlicePtr    *[]int
	intPtrSlicePtr *[]*int

	nilMap map[string]int

	stringSetMap map[string]struct{}

	stringIntMap    map[string]int
	stringIntPtrMap map[string]*int
	stringIntMapPtr *map[string]int

	structKeyIntMap   map[point]int
	intArrayStringMap map[[2]int]string

	intStructMap    map[int]point
	intStructPtrMap map[int]*point

	intStructSliceMap    map[int][]point
	intStructPtrSliceMap map[int][]*point

	intStructArrayMap    map[int][2]point
	intStructPtrArrayMap map[int][2]*point

	stringer     fmt.Stringer
	nullStringer fmt.Stringer
}

func TestMarshalJSON_Struct(t *testing.T) {
	makeValue := func() structType {
		return structType{
			int:        1,
			int8:       2,
			int16:      3,
			int32:      4,
			int64:      5,
			uint:       6,
			uint8:      7,
			uint16:     8,
			uint32:     9,
			uint64:     10,
			uintptr:    11,
			float32:    12.0,
			float64:    13.0,
			jsonNumber: "14",
			string:     "fifteen",

			complex64:  16 + 17i,
			complex128: 18 + 19i,

			intPtr:    ptrTo(int(1)),
			intPtrPtr: ptrTo(ptrTo(int(1))),

			intPtrNull: nil,

			intSlice:    []int{1, 2, 3},
			intPtrSlice: []*int{ptrTo(int(1)), nil, ptrTo(int(3))},

			intArray:    [3]int{1, 2, 3},
			intPtrArray: [3]*int{ptrTo(int(1)), nil, ptrTo(int(3))},

			intSlicePtr:    ptrTo([]int{1, 2, 3}),
			intPtrSlicePtr: ptrTo([]*int{ptrTo(int(1)), nil, ptrTo(int(3))}),

			stringSetMap: map[string]struct{}{"a": {}, "b": {}, "c": {}},

			stringIntMap:    map[string]int{"a": 1, "b": 2, "c": 3},
			stringIntPtrMap: map[string]*int{"a": ptrTo(int(1)), "b": ptrTo(int(2)), "c": ptrTo(int(3))},
			stringIntMapPtr: ptrTo(map[string]int{"a": 1, "b": 2, "c": 3}),

			structKeyIntMap:   map[point]int{{1, 2}: 1, {2, 3}: 2, {3, 4}: 3},
			intArrayStringMap: map[[2]int]string{{1, 2}: "a", {2, 3}: "b", {3, 4}: "c"},

			intStructMap:    map[int]point{1: {1, 2}, 2: {2, 3}, 3: {3, 4}},
			intStructPtrMap: map[int]*point{1: ptrTo(point{1, 2}), 2: ptrTo(point{2, 3}), 3: ptrTo(point{3, 4})},

			intStructSliceMap: map[int][]point{
				1: {{1, 2}, {1, 3}},
				2: {{2, 3}, {2, 4}},
				3: {{3, 4}, {3, 5}},
			},
			intStructPtrSliceMap: map[int][]*point{
				1: {ptrTo(point{1, 2}), ptrTo(point{1, 3})},
				2: {ptrTo(point{2, 3}), ptrTo(point{2, 4})},
				3: {ptrTo(point{3, 4}), ptrTo(point{3, 5})},
			},

			intStructArrayMap: map[int][2]point{
				1: {{1, 2}, {1, 3}},
				2: {{2, 3}, {2, 4}},
				3: {{3, 4}, {3, 5}},
			},
			intStructPtrArrayMap: map[int][2]*point{
				1: {ptrTo(point{1, 2}), ptrTo(point{1, 3})},
				2: {ptrTo(point{2, 3}), ptrTo(point{2, 4})},
				3: {ptrTo(point{3, 4}), ptrTo(point{3, 5})},
			},

			stringer: makeStringer("hello"),
		}
	}

	tests := map[string]any{
		"structType": makeValue(),
	}

	testMarshalJSON(t, tests)
}

// Tests whether we can marshal and unmarshal structs that are dynamically
// defined via reflection. Note that we need to explicitly register the
// type when unmarshaling because it was dynamically created.
func TestMarshalJSON_Struct_DynamicDefinition(t *testing.T) {
	pointT := reflect.StructOf([]reflect.StructField{
		{
			PkgPath: "unexported",
			Name:    "x",
			Type:    reflect.TypeFor[int](),
		},
		{
			Name: "Y", // "exported"
			Type: reflect.TypeFor[int](),
		},
	})

	newPoint := func(x, y int) any {
		pointV := reflect.New(pointT).Elem()
		setField(pointV.FieldByName("x"), reflect.ValueOf(x))
		setField(pointV.FieldByName("Y"), reflect.ValueOf(y))
		return pointV.Interface()
	}

	type structWithDynamic struct {
		values []any
	}

	in := structWithDynamic{
		values: []any{
			newPoint(1, 2),
			newPoint(3, 4),
		},
	}

	b, err := MarshalJSON(in, WithIndent("  "))
	require.NoError(t, err)
	assert.JSONEq(t, `
{
  "value": {
    "values": [
      {
        "typeString": "struct { x int; Y int }",
        "value": {
          "x": 1,
          "Y": 2
        }
      },
      {
        "typeString": "struct { x int; Y int }",
        "value": {
          "x": 3,
          "Y": 4
        }
      }
    ]
  }
}`, string(b))

	var out structWithDynamic
	typeResolver := typeutil.NewStaticResolver().AddTypes(pointT)
	require.NoError(t, UnmarshalJSON(b, &out, WithTypeResolver(typeResolver)))
	assert.Equal(t, in, out)
}

// Tests whether we can marshal and unmarshal structs with inline definitions.
func TestMarshalJSON_Struct_WithInlineDefinitions(t *testing.T) {
	type structWithInline struct {
		inline struct {
			value string
		}

		inlineSlice []struct {
			x int
			y int
		}
	}

	in := structWithInline{
		inline: struct {
			value string
		}{
			value: "hello",
		},

		inlineSlice: []struct {
			x int
			y int
		}{
			{1, 2}, {3, 4},
		},
	}

	b, err := MarshalJSON(in, WithIndent("  "))
	require.NoError(t, err)
	assert.JSONEq(t, `
{
  "value": {
		"inline": {
			"value": "hello"
		},
		"inlineSlice": [
			{
				"x": 1,
				"y": 2
			},
			{
				"x": 3,
				"y": 4
			}
		]
  }
}`, string(b))

	var out structWithInline
	require.NoError(t, UnmarshalJSON(b, &out, WithTypeResolver(typeutil.UnsafeResolver())))
	assert.Equal(t, in, out)
}

// Various tests of nil/empty interfaces and structs.
func TestMarshalJSON_Struct_WithNilInterface(t *testing.T) {
	type structWithInterfaces struct {
		nilInterface fmt.Stringer
		nilStruct    fmt.Stringer
		emptyString  fmt.Stringer
		emptyStruct  fmt.Stringer
	}

	var stringerPtr *stringer

	in := structWithInterfaces{
		nilInterface: nil,
		nilStruct:    stringerPtr,
		emptyString:  stringAlias(""),
		emptyStruct:  stringer{},
	}

	b, err := MarshalJSON(in, WithIndent("  "))
	require.NoError(t, err)
	assert.JSONEq(t, `
{
  "value": {
		"nilInterface": null,
		"nilStruct": {
			"ptrDepth": 1,
			"pkgPath": "github.com/outriggerlabs/unsafely",
			"typeName": "stringer",
			"value": {
				"pointer": 1,
				"value": null
			}
		},
		"emptyString": {
			"pkgPath": "github.com/outriggerlabs/unsafely",
			"typeName": "stringAlias",
			"value": ""
		},
		"emptyStruct": {
			"pkgPath": "github.com/outriggerlabs/unsafely",
			"typeName": "stringer",
			"value": {
				"value": null
			}
		}
  }
}`, string(b))

	var out structWithInterfaces
	require.NoError(t, UnmarshalJSON(b, &out, WithTypeResolver(typeutil.UnsafeResolver())))
	assert.Equal(t, in, out)
}

func TestMarshalJSON_Struct_EmbeddedTypes(t *testing.T) {
	type inner struct {
		value int
	}

	type outer struct {
		inner
	}

	tests := map[string]any{
		"outer": outer{inner{value: 42}},
	}

	testMarshalJSON(t, tests)
}

// Test that we support dash as a JSON name.
func TestMarshalJSON_Dash(t *testing.T) {
	type jsonDash struct {
		Val int `json:"-,"`
	}

	type unsafelyJSONDash struct {
		Val int `json:"-,"`
	}

	var (
		jsonTag         = jsonDash{Val: 1}
		unsafelyJSONTag = unsafelyJSONDash{Val: 1}
	)

	b, err := MarshalJSON(jsonTag)
	require.NoError(t, err)
	fmt.Println(string(b))
	assert.JSONEq(t, `{"value":{"-":1}}`, string(b))

	b2, err := MarshalJSON(unsafelyJSONTag)
	require.NoError(t, err)
	assert.JSONEq(t, `{"value":{"-":1}}`, string(b2))

	tests := map[string]any{
		"json tag":          jsonTag,
		"unsafely json tag": unsafelyJSONTag,
	}

	testMarshalJSON(t, tests)
}
