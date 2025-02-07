package unsafely

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalJSON_Options_IndentAndPrefix(t *testing.T) {
	type testStruct struct {
		String  string
		Int     int
		Slice   []string
		Map     map[string]int
		Complex complex128
		StrPtr  *string
		Any     any
	}

	value := testStruct{
		String:  "hello",
		Int:     42,
		Slice:   []string{"a", "b", "c"},
		Map:     map[string]int{"x": 1, "y": 2},
		Complex: 1 + 2i,
		StrPtr:  ptrTo("pointer"),
		Any:     []int{1, 2, 3},
	}

	tests := map[string]struct {
		options  []MarshalJSONOption
		expected string
	}{
		"no options": {
			options:  nil,
			expected: `{"value":{"String":"hello","Int":42,"Slice":["a","b","c"],"Map":{"x":1,"y":2},"Complex":{"real":1,"imag":2},"StrPtr":{"pointer":1,"value":"pointer"},"Any":{"typeString":"[]int","value":[1,2,3]}}}`,
		},
		"with indent": {
			options: []MarshalJSONOption{WithIndent("  ")},
			expected: `{
  "value": {
    "String": "hello",
    "Int": 42,
    "Slice": [
      "a",
      "b",
      "c"
    ],
    "Map": {
      "x": 1,
      "y": 2
    },
    "Complex": {
      "real": 1,
      "imag": 2
    },
    "StrPtr": {
      "pointer": 1,
      "value": "pointer"
    },
    "Any": {
      "typeString": "[]int",
      "value": [
        1,
        2,
        3
      ]
    }
  }
}`,
		},
		"with prefix": {
			options: []MarshalJSONOption{WithPrefix(">>>")},
			expected: `{
>>>"value": {
>>>"String": "hello",
>>>"Int": 42,
>>>"Slice": [
>>>"a",
>>>"b",
>>>"c"
>>>],
>>>"Map": {
>>>"x": 1,
>>>"y": 2
>>>},
>>>"Complex": {
>>>"real": 1,
>>>"imag": 2
>>>},
>>>"StrPtr": {
>>>"pointer": 1,
>>>"value": "pointer"
>>>},
>>>"Any": {
>>>"typeString": "[]int",
>>>"value": [
>>>1,
>>>2,
>>>3
>>>]
>>>}
>>>}
>>>}`,
		},
		"with both": {
			options: []MarshalJSONOption{WithPrefix(">>>"), WithIndent("  ")},
			expected: `{
>>>  "value": {
>>>    "String": "hello",
>>>    "Int": 42,
>>>    "Slice": [
>>>      "a",
>>>      "b",
>>>      "c"
>>>    ],
>>>    "Map": {
>>>      "x": 1,
>>>      "y": 2
>>>    },
>>>    "Complex": {
>>>      "real": 1,
>>>      "imag": 2
>>>    },
>>>    "StrPtr": {
>>>      "pointer": 1,
>>>      "value": "pointer"
>>>    },
>>>    "Any": {
>>>      "typeString": "[]int",
>>>      "value": [
>>>        1,
>>>        2,
>>>        3
>>>      ]
>>>    }
>>>  }
>>>}`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := MarshalJSON(value, tt.options...)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(output))
		})
	}
}
