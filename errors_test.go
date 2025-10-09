package unsafely

import (
	"fmt"
	"testing"

	perr "github.com/outriggerlabs/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrProps(t *testing.T) {
	errProps := errProps()

	// Get the builder and extract the op property
	builder := errProps()
	err := builder.Annotate(fmt.Errorf("test error"))

	require.Error(t, err)

	// Extract the "op" property using the Property function
	op := perr.Property[string](err, "op")
	assert.Equal(t, "unsafely.TestErrProps", op)
}

func TestErrPropsInMethod(t *testing.T) {
	decoder := &JSONDecoder{}
	err := decoder.testHelperMethod()

	require.Error(t, err)

	// The op should be "unsafely.testHelperMethod"
	op := perr.Property[string](err, "op")
	assert.Equal(t, "unsafely.testHelperMethod", op)
}

// testHelperMethod is a helper method for testing errProps() in method context
func (s *JSONDecoder) testHelperMethod() error {
	errProps := errProps()
	return errProps().Annotate(fmt.Errorf("method error"))
}

func TestErrPropsMultipleCalls(t *testing.T) {
	// Test that errProps can be called multiple times and each returns
	// a fresh builder
	errProps := errProps()

	err1 := errProps().Annotate(fmt.Errorf("error 1"))
	err2 := errProps().Annotate(fmt.Errorf("error 2"))

	require.Error(t, err1)
	require.Error(t, err2)

	// Both should have the same op name
	op1 := perr.Property[string](err1, "op")
	op2 := perr.Property[string](err2, "op")
	assert.Equal(t, "unsafely.TestErrPropsMultipleCalls", op1)
	assert.Equal(t, "unsafely.TestErrPropsMultipleCalls", op2)

	// But the underlying error messages should be different
	assert.Contains(t, err1.Error(), "error 1")
	assert.Contains(t, err2.Error(), "error 2")
}

func TestFormatOpName(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"simple function": {
			input:    "github.com/outriggerlabs/unsafely.MarshalJSON",
			expected: "unsafely.MarshalJSON",
		},
		"method with pointer receiver": {
			input:    "github.com/outriggerlabs/unsafely.(*JSONDecoder).Decode",
			expected: "unsafely.Decode",
		},
		"method with value receiver": {
			input:    "github.com/outriggerlabs/unsafely.(JSONEncoder).Encode",
			expected: "unsafely.Encode",
		},
		"function with closure suffix": {
			input:    "github.com/outriggerlabs/unsafely.(*JSONDecoder).Decode-fm",
			expected: "unsafely.Decode",
		},
		"nested package": {
			input:    "github.com/outriggerlabs/unsafely/typeutil.(*StaticResolver).Resolve",
			expected: "typeutil.Resolve",
		},
		"no package path": {
			input:    "main.main",
			expected: "main.main",
		},
		"single element": {
			input:    "unsafely",
			expected: "unsafely",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := formatOpName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOpName(t *testing.T) {
	name := opName()
	assert.Equal(t, "unsafely.TestOpName", name)
}

func TestOpNameInHelper(t *testing.T) {
	name := helperForOpName()
	assert.Equal(t, "unsafely.helperForOpName", name)
}

func helperForOpName() string {
	return opName()
}

func TestErrPropsWithExtend(t *testing.T) {
	// Test that errProps works with Extend() as well as Annotate()
	errProps := errProps()

	innerErr := fmt.Errorf("inner error")
	outerErr := errProps().Extend(innerErr)

	require.Error(t, outerErr)

	op := perr.Property[string](outerErr, "op")
	assert.Equal(t, "unsafely.TestErrPropsWithExtend", op)
}
