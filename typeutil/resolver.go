package typeutil

import "reflect"

// Resolver is an interface for resolving reflect.Type objects.
type Resolver interface {
	// ResolveType resolves a type or returns an error.
	//
	// Since the reflect.Type.String() is not guaranteed unique, it is recommended
	// to resolve by package path and name when possible.
	//
	// However, there are some cases when package path and name will not be
	// present, e.g, for dynamically defined structs, and the string value will be
	// necessary.
	ResolveType(pkgPath string, typeName string, typeString string) (reflect.Type, error)
}
