package typeutil

import (
	"fmt"
	"reflect"
)

// StaticResolver is a type resolver that uses a static map of types.
type StaticResolver struct {
	// Map from package path -> type name -> type.
	packageNames map[string]map[string]reflect.Type

	// Map from type string -> type.
	typeStrings map[string]reflect.Type
}

// NewStaticResolver returns a new StaticResolver.
func NewStaticResolver() StaticResolver {
	return StaticResolver{
		packageNames: make(map[string]map[string]reflect.Type),
		typeStrings:  make(map[string]reflect.Type),
	}
}

// AddTypes adds the types to the StaticResolver.
func (s StaticResolver) AddTypes(types ...reflect.Type) StaticResolver {
	for _, typ := range types {
		var (
			pkgPath = typ.PkgPath()
			name    = typ.Name()
		)

		pkgTypes := s.packageNames[pkgPath]
		if pkgTypes == nil {
			pkgTypes = make(map[string]reflect.Type)
			s.packageNames[pkgPath] = pkgTypes
		}

		// Only store the type by string if we don't have a name and package path.
		if pkgPath == "" && name == "" {
			s.typeStrings[typ.String()] = typ
		} else {
			pkgTypes[name] = typ
		}
	}

	return s
}

// ResolveType (see Resolver.ResolveType).
func (s StaticResolver) ResolveType(pkgPath, typeName, typeString string) (reflect.Type, error) {
	var typ reflect.Type

	if pkgPath == "" && typeName == "" {
		typ = s.typeStrings[typeString]
	} else {
		typ = s.packageNames[pkgPath][typeName]
	}

	if typ == nil {
		return nil, fmt.Errorf("StaticResolver.ResolveType(): could not find type for "+
			"pkgPath: %s, typeName: %s, typeString: %s", pkgPath, typeName, typeString)
	}

	return typ, nil
}
