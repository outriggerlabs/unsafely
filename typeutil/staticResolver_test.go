package typeutil_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/outriggerlabs/unsafely/typeutil"
)

func TestStaticResolver(t *testing.T) {
	type testStruct struct {
		Field string
	}

	tests := map[string]struct {
		types        []reflect.Type
		pkgPath      string
		typeName     string
		typeString   string
		expectedType reflect.Type
		expectedErr  string
	}{
		"resolve by type string": {
			types:        []reflect.Type{reflect.TypeOf(map[string]int{})},
			typeString:   "map[string]int",
			expectedType: reflect.TypeOf(map[string]int{}),
		},
		"resolve by package path and name": {
			types:        []reflect.Type{reflect.TypeOf(testStruct{})},
			pkgPath:      "github.com/outriggerlabs/unsafely/typeutil_test",
			typeName:     "testStruct",
			expectedType: reflect.TypeOf(testStruct{}),
		},
		"type string not found": {
			types:       []reflect.Type{},
			typeString:  "nonexistentType",
			expectedErr: "could not find type for pkgPath: , typeName: , typeString: nonexistentType",
		},
		"package path not found": {
			types:       []reflect.Type{},
			pkgPath:     "nonexistent/pkg",
			typeName:    "testStruct",
			expectedErr: fmt.Sprintf("could not find type for pkgPath: nonexistent/pkg, typeName: testStruct, typeString: "),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resolver := typeutil.NewStaticResolver().AddTypes(tt.types...)

			result, err := resolver.ResolveType(tt.pkgPath, tt.typeName, tt.typeString)
			if tt.expectedErr != "" {
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedType, result)
			}
		})
	}
}
