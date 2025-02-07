package unsafely

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// The type used to represent interface values in encoded structs.
var interfaceValueType = reflect.TypeFor[*interfaceValue]()

// Returns true if the provided type is an *interfaceValue.
func isInterfaceValueType(t reflect.Type) bool {
	return t == interfaceValueType
}

// Represents a value that was stored in an interface field. If the
// interfaceValue is nil, the underlying value was a nil interface.
type interfaceValue struct {
	// PtrDepth is the number of pointer indirections to the type for the value.
	PtrDepth int `json:"ptrDepth,omitempty"`

	// PkgPath is the package path of the type; empty for built-in types.
	PkgPath string `json:"pkgPath,omitempty"`

	// TypeName is the name of the type, namespaced by PkgPath.
	TypeName string `json:"typeName,omitempty"`

	// TypeString is the string representation of the type, only included if the
	// PkgPath and TypeName are empty.
	TypeString string `json:"typeString,omitempty"`

	// Value is a JSON string representing the underlying value.
	Value json.RawMessage `json:"value"`
}

// Encodes the value to an interfaceValue object.
//
// The underlying value is marshaled and stored as JSON.
//
// If the input value is the nil interface, returns the zero value.
func (s *JSONEncoder) encodeToInterfaceValue(inV reflect.Value) (out reflect.Value, err error) {
	if inV.Kind() != reflect.Interface {
		return zeroValue, fmt.Errorf(
			"encodeToInterfaceValue: expected value to be an interface; received %v (%T)",
			inV.Kind(), inV,
		)
	}

	// nil interface
	if inV.IsZero() {
		return zeroValue, nil
	}

	// inV is the interface type; decodedV is the underlying type.
	decodedV := ensureAddressable(inV.Elem())

	// Encode the underlying value to JSON.
	encodedV, err := s.encode(decodedV)
	if err != nil {
		return zeroValue, fmt.Errorf("encodeToInterfaceValue: %w", err)
	}

	encodedBytes, err := s.jsonMarshalInternal(encodedV.Interface())
	if err != nil {
		return zeroValue, fmt.Errorf("encodeToInterfaceValue: %w", err)
	}

	// Pointers seem to have no package path or name, and the string
	// representation isn't canonical.
	//
	// To capture a concrete type, we extract the underlying type of the pointer
	// and record that and the pointer depth separately.
	var (
		pointerDepth int
		recordedT    = decodedV.Type()
	)
	for recordedT.Kind() == reflect.Pointer {
		pointerDepth++
		recordedT = recordedT.Elem()
	}

	var (
		pkgPath    = recordedT.PkgPath()
		typeName   = recordedT.Name()
		typeString string
	)

	if pkgPath == "" && typeName == "" {
		typeString = recordedT.String()
	}

	iv := &interfaceValue{
		PtrDepth:   pointerDepth,
		PkgPath:    pkgPath,
		TypeName:   typeName,
		TypeString: typeString,
		Value:      encodedBytes,
	}

	return reflect.ValueOf(iv), nil
}

// Decodes the interfaceValue object back to the original interface.
//
// Returns a zero reflect.Value if the interfaceValue is nil.
func (s *JSONDecoder) decodeFromInterfaceValue(inV reflect.Value) (reflect.Value, error) {
	iv, ok := inV.Interface().(*interfaceValue)
	if !ok {
		return zeroValue, fmt.Errorf(
			"decodeFromInterfaceValue(): expected value to be an *interfaceValue; received %T",
			inV.Interface(),
		)
	}

	// Nil interface
	if iv == nil {
		return zeroValue, nil
	}

	if s.config.typeResolver == nil {
		return zeroValue, errors.New(
			"decodeFromInterfaceValue(): a type resolver must be configured using WithTypeResolver() " +
				"to resolve types to interface values")
	}

	resolvedT, err := s.config.typeResolver.ResolveType(iv.PkgPath, iv.TypeName, iv.TypeString)
	if err != nil {
		return zeroValue, fmt.Errorf("decodeFromInterfaceValue(): %w", err)
	}

	// Add the pointer indirections recorded in the pointer depth.
	decodedT := resolvedT
	for range iv.PtrDepth {
		decodedT = reflect.PointerTo(decodedT)
	}

	encodedT, err := encodedTypeFor(decodedT)
	if err != nil {
		return zeroValue, fmt.Errorf("decodeFromInterfaceValue(): %w", err)
	}

	encodedPtrV := reflect.New(encodedT)
	if err := json.Unmarshal(iv.Value, encodedPtrV.Interface()); err != nil {
		return zeroValue, fmt.Errorf("decodeFromInterfaceValue(): %w", err)
	}

	var (
		encodedV = encodedPtrV.Elem()
		decodedV = reflect.New(decodedT).Elem()
	)
	if err := s.decodeTo(encodedV, decodedV); err != nil {
		return zeroValue, fmt.Errorf("decodeFromInterfaceValue(): %w", err)
	}

	return decodedV, nil
}
