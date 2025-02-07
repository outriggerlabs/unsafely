package unsafely

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONEncoder_SharedPointers(t *testing.T) {
	var (
		int1 = 42
		int2 = 53
	)

	type object1 struct {
		a *int
		b *int
	}
	obj1 := object1{a: &int1, b: &int2}

	// Second object just contains the second pointer to demonstrate sharing
	// of the pointer reference.
	type object2 struct {
		c *int
	}
	obj2 := object2{c: &int2}

	// Encode both objects using the same encoder to maintain pointer references.
	encoder := NewJSONEncoder()

	encoded1, err := encoder.Encode(obj1)
	require.NoError(t, err)

	encoded2, err := encoder.Encode(obj2)
	require.NoError(t, err)

	// The first object has pointer values 1 and 2.
	expectedJSON1 := `{"value":{"a":{"pointer":1,"value":42},"b":{"pointer":2,"value":53}}}`
	assert.JSONEq(t, expectedJSON1, string(encoded1))

	// The second object should reuse pointer value 2.
	expectedJSON2 := `{"value":{"c":{"pointer":2,"value":53}}}`
	assert.JSONEq(t, expectedJSON2, string(encoded2))

	// Decode both objects using the same decoder and verify they share the same
	// second pointer.
	decoder := NewJSONDecoder()

	var decoded1 object1
	err = decoder.Decode(encoded1, &decoded1)
	require.NoError(t, err)

	var decoded2 object2
	err = decoder.Decode(encoded2, &decoded2)
	require.NoError(t, err)

	assert.Same(t, decoded1.b, decoded2.c)

	// Note that even though the decoded objects share the same second pointer,
	// the underlying objects are different from the original values.
	assert.NotSame(t, obj1.b, decoded1.b)
}
