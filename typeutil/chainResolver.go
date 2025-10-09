package typeutil

import (
	"fmt"
	"reflect"
)

type chainResolver struct {
	resolvers []Resolver
}

// NewChainResolver returns a Resolver that returns the first resolved type from
// a chain of Resolvers.
func NewChainResolver(resolvers ...Resolver) Resolver {
	return chainResolver{resolvers: resolvers}
}

// ResolveType (see Resolver.ResolveType).
func (s chainResolver) ResolveType(
	pkgPath string,
	typeName string,
	typeString string,
) (reflect.Type, error) {
	// Return the first resolved type.
	for _, resolver := range s.resolvers {
		typ, _ := resolver.ResolveType(pkgPath, typeName, typeString)
		if typ != nil {
			return typ, nil
		}
	}

	return nil, errProps().Annotate(fmt.Errorf("could not find type for "+
		"pkgPath: %s, typeName: %s, typeString: %s", pkgPath, typeName, typeString))
}
