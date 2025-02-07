package typeutil_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/outriggerlabs/unsafely/typeutil"
)

func TestChainResolver(t *testing.T) {
	type testStruct struct {
		Field string
	}

	tests := map[string]struct {
		resolvers    []typeutil.Resolver
		pkgPath      string
		typeName     string
		typeString   string
		expectedType reflect.Type
		expectedErr  string
	}{
		"resolve from first resolver": {
			resolvers: []typeutil.Resolver{
				typeutil.NewStaticResolver().AddTypes(reflect.TypeOf(map[string]int{})),
				typeutil.NewStaticResolver().AddTypes(reflect.TypeOf(testStruct{})),
			},
			typeString:   "map[string]int",
			expectedType: reflect.TypeOf(map[string]int{}),
		},
		"resolve from second resolver": {
			resolvers: []typeutil.Resolver{
				typeutil.NewStaticResolver().AddTypes(reflect.TypeOf(map[string]int{})),
				typeutil.NewStaticResolver().AddTypes(reflect.TypeOf(testStruct{})),
			},
			pkgPath:      "github.com/outriggerlabs/unsafely/typeutil_test",
			typeName:     "testStruct",
			expectedType: reflect.TypeOf(testStruct{}),
		},
		"type not found in any resolver": {
			resolvers: []typeutil.Resolver{
				typeutil.NewStaticResolver().AddTypes(reflect.TypeOf(map[string]int{})),
				typeutil.NewStaticResolver().AddTypes(reflect.TypeOf(testStruct{})),
			},
			typeString:  "nonexistentType",
			expectedErr: "could not find type for pkgPath: , typeName: , typeString: nonexistentType",
		},
		"package path and type name not found": {
			resolvers: []typeutil.Resolver{
				typeutil.NewStaticResolver().AddTypes(reflect.TypeOf(map[string]int{})),
				typeutil.NewStaticResolver().AddTypes(reflect.TypeOf(testStruct{})),
			},
			pkgPath:     "github.com/outriggerlabs/unsafely/typeutil_test",
			typeName:    "nonExistentType",
			expectedErr: "could not find type for pkgPath: github.com/outriggerlabs/unsafely/typeutil_test, typeName: nonExistentType, typeString: ",
		},
		"empty resolver chain": {
			resolvers:   []typeutil.Resolver{},
			typeString:  "map[string]int",
			expectedErr: "could not find type for pkgPath: , typeName: , typeString: map[string]int",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resolver := typeutil.NewChainResolver(tt.resolvers...)

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
