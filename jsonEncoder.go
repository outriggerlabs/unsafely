package unsafely

import (
	"encoding/json"
	"fmt"
	"reflect"
	"unsafe"
)

// JSONEncoder is a helper struct for encoding values to JSON.
//
// See MarshalJSON and JSONEncoder.Encode.
type JSONEncoder struct {
	config marshalJSONConfig

	// Used to calculate the pointer reference numbers for pointerValues.
	pointerIndex int

	// Map from pointers to previously encoded values.
	pointerValues map[unsafe.Pointer]reflect.Value

	// Pointers that are in processing, to avoid cycles.
	pendingPointers map[unsafe.Pointer]struct{}
}

// NewJSONEncoder creates a JSONEncoder with the given options.
//
// If the same JSONEncoder is used to marshal multiple values that share the
// same pointer values, the output will also contain the same pointer
// references. Unmarshaling those JSON strings with the same JSONDecoder will
// create objects that share the recreated pointer values.
func NewJSONEncoder(options ...MarshalJSONOption) *JSONEncoder {
	var config marshalJSONConfig
	for _, opt := range options {
		opt(&config)
	}

	return &JSONEncoder{
		config:          config,
		pointerValues:   make(map[unsafe.Pointer]reflect.Value),
		pendingPointers: make(map[unsafe.Pointer]struct{}),
	}
}

// Encode serializes the value to a JSON string, including unexported
// fields.
//
// See the package notes for restrictions, limitations and options.
func (s *JSONEncoder) Encode(in any) ([]byte, error) {
	var (
		inV     = reflect.ValueOf(in)
		encoded any
	)

	if inV.IsValid() /* non-nil */ {
		inV = ensureAddressable(inV)

		encodedV, err := s.encode(inV)
		if err != nil {
			return nil, fmt.Errorf("MarshalJSON: %w", err)
		}
		encoded = encodedV.Interface()
	}

	encodedBytes, err := s.jsonMarshalInternal(encoded)
	if err != nil {
		return nil, fmt.Errorf("MarshalJSON: %w", err)
	}

	// Wrap the encoded value.
	//
	// Future: Add any configuration options required for unmarshaling.
	out := encodedJSONWrapper{
		Value: encodedBytes,
	}

	// Output is a single line if prefix and indent are empty; multiline otherwise,
	if s.config.prefix == "" && s.config.indent == "" {
		return json.Marshal(out)
	}
	return json.MarshalIndent(out, s.config.prefix, s.config.indent)
}

// Marshals an internal value to json. The prefix should only be applied once
// at the end, but the indent should be applied to internal values.
func (s *JSONEncoder) jsonMarshalInternal(v any) ([]byte, error) {
	if s.config.prefix == "" && s.config.indent == "" {
		return json.Marshal(v)
	}

	// We want to split these into multiple lines, even if indent is empty, so
	// the prefix is consistently applied to all lines.
	return json.MarshalIndent(v, "", s.config.indent)
}

// Encodes the fromV and writes it to the encodedV.
func (s *JSONEncoder) encodeTo(originalV, encodedV reflect.Value) error {
	var (
		originalT = originalV.Type()
		encodedT  = encodedV.Type()

		originalKind = originalT.Kind()
		encodedKind  = encodedT.Kind()
	)

	// If the value implements json.Marshaler, we defer to the existing marshaling
	// mechanism and simply store the output.
	if originalT.Implements(jsonMarshalerType) {
		b, err := s.jsonMarshalInternal(originalV.Interface())
		if err != nil {
			return fmt.Errorf("encodeTo(): custom json.Marshal for %s failed: %w",
				originalT.String(), err)
		}
		setField(encodedV, reflect.ValueOf(json.RawMessage(b)))
		return nil
	}

	// We're encoding a pointer value.
	if isPointerValueType(encodedT) {
		pv, err := s.encodeToPointerValue(originalV)
		if err != nil {
			return err
		}

		setField(encodedV, pv)
		return nil
	}

	// We're encoding an interface value.
	if isInterfaceValueType(encodedT) {
		iv, err := s.encodeToInterfaceValue(originalV)
		if err != nil {
			return err
		}

		// The zero interface value represents nil, which is the default value.
		if iv != zeroValue {
			setField(encodedV, iv)
		}

		return nil
	}

	// We're encoding a complex value.
	if isComplexValueType(encodedT) {
		cv, err := encodeToComplexValue(originalV)
		if err != nil {
			return err
		}

		setField(encodedV, cv)
		return nil
	}

	if originalKind != encodedKind {
		return fmt.Errorf(
			"encodeTo(): expected values to be the same kind; received %v and %v",
			originalV.Type(), encodedV.Type(),
		)
	}

	if originalKind == reflect.Map {
		if originalV.IsNil() {
			return nil // The map in encodedV is nil by default.
		}

		encodedMap := reflect.MakeMapWithSize(encodedV.Type(), originalV.Len())
		setField(encodedV, encodedMap)

		// Copy the map keys and values.
		originalMapIter := originalV.MapRange()
		for originalMapIter.Next() {
			var (
				originalKey = originalMapIter.Key()
				originalVal = originalMapIter.Value()

				encodedKey = originalKey
			)

			// JSON does not support non-primitive keys (e.g, structs, pointers), so
			// we convert these map keys to JSON strings.
			if !isSimplePrimitive(originalKey.Kind()) {
				// The key may not be addressable, e.g, if it is a struct.
				originalKey = ensureAddressable(originalKey)

				// Encode and marshal the key to a JSON string.
				var err error
				encodedKey, err = s.encode(originalKey)
				if err != nil {
					return fmt.Errorf("encodeTo(): %w", err)
				}

				// Note: JSON doesn't support multiline strings, so just encode to a
				// single line rather than adding prefixes and indents.
				encodedKeyBytes, err := json.Marshal(encodedKey.Interface())
				if err != nil {
					return fmt.Errorf("encodeTo(): %w", err)
				}

				encodedKey = reflect.ValueOf(string(encodedKeyBytes))
			}

			// Encode the map value. The map value may not be addressable, e.g, if
			// it is a struct.
			encodedVal, err := s.encode(ensureAddressable(originalVal))
			if err != nil {
				return err
			}

			// Set the key and value on the encoded map.
			encodedMap.SetMapIndex(encodedKey, encodedVal)
		}

		return nil
	}

	if originalKind == reflect.Struct {
		return copyStruct(s.encodeTo, originalV, encodedV, true /* isEncode */)
	}

	return copyCommon(s.encodeTo, originalV, encodedV)
}

// Encodes the original value for marhsaling to JSON.
func (s *JSONEncoder) encode(fromV reflect.Value) (reflect.Value, error) {
	if fromV == zeroValue {
		return zeroValue, nil
	}

	encodedT, err := encodedTypeFor(fromV.Type())
	if err != nil {
		return zeroValue, fmt.Errorf("encode(): %w", err)
	}

	encodedV := reflect.New(encodedT).Elem()
	if err := s.encodeTo(fromV, encodedV); err != nil {
		return zeroValue, fmt.Errorf("encode(): %w", err)
	}

	return encodedV, nil
}
