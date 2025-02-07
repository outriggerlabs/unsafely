package unsafely

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Struct that implements fmt.Stringer. Note that this also has an interface
// inside.
type stringer struct {
	value any
}

func (s stringer) String() string {
	return fmt.Sprint(s.value)
}

func makeStringer(in any) fmt.Stringer {
	if in == nil {
		return nil
	}

	return stringer{value: in}
}

// Alias type for string that implementation fmt.Stringer.
type stringAlias string

func (s stringAlias) String() string {
	return string(s)
}

func TestMarshalJSON_Interface(t *testing.T) {
	tests := map[string]fmt.Stringer{
		"string":       makeStringer("hello"),
		"int":          makeStringer(42),
		"string alias": makeStringer(stringAlias("goodbye")),
	}

	testMarshalJSON(t, tests)
}

func TestMarshalJSON_Interface_Null(t *testing.T) {
	var null fmt.Stringer
	b, err := MarshalJSON(null)
	require.NoError(t, err)
	assert.JSONEq(t, `{"value":null}`, string(b))

	var decoded fmt.Stringer
	require.NoError(t, UnmarshalJSON(b, &decoded))
	assert.Equal(t, null, decoded)
}

func TestMarshalJSON_GenericInterface(t *testing.T) {
	type generic2[T any] struct {
		v T
	}
	type wrapper struct {
		internal any
	}

	tests := map[string]any{
		"generic2[int]":          generic2[int]{v: 42},
		"wrapper{generic2[int]}": wrapper{internal: generic2[int]{v: 42}},
	}

	testMarshalJSON(t, tests)
}
