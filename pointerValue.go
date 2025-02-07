package unsafely

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"unsafe"
)

// The type used to represent pointer values in encoded structs.
var pointerValueType = reflect.TypeFor[pointerValue]()

// Returns true if the provided type is a pointerValue.
func isPointerValueType(t reflect.Type) bool {
	return t == pointerValueType
}

// Represents a pointer stored in a field.
type pointerValue struct {
	// An integer indicating a reference number for a pointer. All pointerValue
	// objects with the same pointer number represent the same pointer.
	Pointer int `json:"pointer"`

	// The JSON representation of the underlying value; null if the pointer is
	// nil.
	Value json.RawMessage `json:"value,omitempty"`
}

// Encodes the value to a pointerValue object.
//
// The underlying value is marshaled and stored as JSON.
//
// If the pointer has been seen before, a encodeTo of the existing pointerValue
// object may be returned, rather than re-marshaling the underlying value.
func (s *JSONEncoder) encodeToPointerValue(inV reflect.Value) (out reflect.Value, err error) {
	if inV.Kind() != reflect.Pointer {
		return reflect.Value{}, fmt.Errorf(
			"encodeToPointerValue: expected value to be a pointer; received %v", inV.Kind(),
		)
	}

	var (
		zeroPointer unsafe.Pointer
		inPtr       = inV.UnsafePointer()
		isNil       = inV.IsNil()
	)

	// The expectation is that we only see zero pointers iff the input is nil.
	// We don't cache zero pointers.
	if isNil != (inPtr == zeroPointer) {
		return reflect.Value{}, fmt.Errorf(
			"encodeToPointerValue: assertion failed; isNil: %v, inPtr: %v", inV.IsNil(), inPtr,
		)
	}

	// Marshal the underlying type; default to null.
	var value = json.RawMessage("null")

	if !isNil {
		// Store pointers that we've seen so we don't need to remarshal them later.
		if val, ok := s.pointerValues[inPtr]; ok {
			return val, nil
		}

		// Track the pointers that we're processing to ensure we don't have any data
		// cycles.
		if _, pending := s.pendingPointers[inPtr]; pending {
			return reflect.Value{}, fmt.Errorf(
				"encodeToPointerValue: cycle detected; structs with cyclic data are not supported",
			)
		}
		s.pendingPointers[inPtr] = struct{}{}

		// Encode the underlying value.
		outV, err := s.encode(inV.Elem())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("encodeToPointerValue: %w", err)
		}

		// We're done processing this pointer, so stop tracking it.
		delete(s.pendingPointers, inPtr)

		value, err = s.jsonMarshalInternal(outV.Interface())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("encodeToPointerValue: %w", err)
		}
	}

	s.pointerIndex++
	pv := reflect.ValueOf(pointerValue{
		Pointer: s.pointerIndex,
		Value:   value,
	})

	if !isNil {
		// Cache the encoded value for the pointer.
		s.pointerValues[inPtr] = pv
	}

	return pv, nil
}

// Decodes the pointerValue object and writes it to outPtrV.
//
// If the pointerValue is null, nothing is written to outPtrV.
func (s *JSONDecoder) decodeFromPointerValue(
	pvV reflect.Value, outPtrV reflect.Value,
) (err error) {
	pv, ok := pvV.Interface().(pointerValue)
	if !ok {
		return fmt.Errorf(
			"convertFromPointerValue: expected inV to be a *pointerValue; received %T", pvV.Interface(),
		)
	}

	if outPtrV.Kind() != reflect.Pointer {
		return fmt.Errorf(
			"convertFromPointerValue: expected outV to be a pointer; received %v", outPtrV.Kind(),
		)
	}

	// Short-circuit null, since outPtrV should already be a null pointer.
	if bytes.Equal(pv.Value, []byte("null")) {
		return nil
	}

	// If we've already decoded the pointerValue before, reuse the existing value.
	if val, ok := s.pointerValues[pv.Pointer]; ok {
		outPtrV.Set(val)
		return nil
	}

	// Unmarshal the underlying JSON value into the encoded type.
	encodedT, err := encodedTypeFor(outPtrV.Type().Elem())
	if err != nil {
		return fmt.Errorf("convertFromPointerValue: %w", err)
	}

	encodedPtrV := reflect.New(encodedT)
	if err := json.Unmarshal(pv.Value, encodedPtrV.Interface()); err != nil {
		return fmt.Errorf("convertFromPointerValue: %w", err)
	}

	// Instantiate the pointer, then decodeTo the value.
	setField(outPtrV, reflect.New(outPtrV.Type().Elem()))
	if err := s.decodeTo(encodedPtrV.Elem(), outPtrV.Elem()); err != nil {
		return fmt.Errorf("convertFromPointerValue: %w", err)
	}

	// Store the decoded value for the pointer so we can reuse it later.
	s.pointerValues[pv.Pointer] = outPtrV
	return nil
}
