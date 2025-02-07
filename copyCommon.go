package unsafely

import (
	"fmt"
	"reflect"
)

// Handles cases that are the same or similar for encoding and decoding.
//
// The copyFn is either encodeTo or decodeTo.
func copyCommon(copyFn func(fromV, toV reflect.Value) error, fromV, toV reflect.Value) error {
	var (
		fromT    = fromV.Type()
		toT      = toV.Type()
		fromKind = fromT.Kind()
	)

	// No conversion is needed for int, string, etc.
	if isSimplePrimitive(fromKind) {
		if fromT != toT {
			return fmt.Errorf(
				"copyCommon(): primitive values must be the same type; received %v and %v",
				fromT, toT)
		}

		setField(toV, fromV)
		return nil
	}

	// Copying a slice or array works the same in both directions, with the
	// underlying encoding/decoding deferred to the copyFn.
	if fromKind == reflect.Slice || fromKind == reflect.Array {
		return copyArrayLike(copyFn, fromV, toV)
	}

	// All other cases must be handled by the caller.
	return fmt.Errorf("copyCommon(): unexpected kind: %v", fromKind)
}

// Copies an array or slice of values. Encoding/decoding is delegated to the copyFn.
func copyArrayLike(copyFn func(fromV, toV reflect.Value) error, fromV, toV reflect.Value) error {
	var (
		fromT    = fromV.Type()
		toT      = toV.Type()
		fromKind = fromT.Kind()
		toKind   = toT.Kind()
	)

	if fromKind != toKind {
		return fmt.Errorf(
			"copyArrayLike(): values must be the same kind; received %v and %v",
			fromKind, toKind)
	}

	if toKind != reflect.Array && toKind != reflect.Slice {
		return fmt.Errorf("copyArrayLike(): must be array or slice, received %v", toKind)
	}

	// Allocate slice, if necessary. This preserves nil slices.
	if fromKind == reflect.Slice {
		if fromV.IsNil() {
			return nil // slice is already nil
		}

		setField(toV, reflect.MakeSlice(toT, fromV.Len(), fromV.Len()))
	}

	// Copy each element using the encoding/decoding function.
	for i := 0; i < fromV.Len(); i++ {
		if err := copyFn(fromV.Index(i), toV.Index(i)); err != nil {
			return err
		}
	}

	return nil
}

// Copies a struct, either for encoding or decoding.
func copyStruct(
	copyFn func(fromV, toV reflect.Value) error,
	fromV, toV reflect.Value,
	isEncode bool,
) error {
	var (
		fromT    = fromV.Type()
		toT      = toV.Type()
		fromKind = fromT.Kind()
		toKind   = toT.Kind()
	)

	if fromKind != reflect.Struct || toKind != reflect.Struct {
		return fmt.Errorf(
			"copyStruct: both values must be structs; received %v and %v",
			fromV.Type(), toV.Type())
	}

	var originalV, encodedV reflect.Value
	if isEncode {
		originalV, encodedV = fromV, toV
	} else {
		originalV, encodedV = toV, fromV
	}

	encodedT := encodedV.Type()
	for i := 0; i < encodedT.NumField(); i++ {
		var (
			encodedFieldV = encodedV.Field(i)
			encodedFieldT = encodedT.Field(i)
		)

		// Get the original field name from the original tag
		originalFieldName := encodedFieldT.Tag.Get("original")
		if originalFieldName == "" {
			return fmt.Errorf("copyStruct(): could not find original field name for field %s", encodedFieldT.Name)
		}

		// Look up the field in the original struct
		originalFieldV := originalV.FieldByName(originalFieldName)
		if !originalFieldV.IsValid() {
			return fmt.Errorf(
				"copyStruct(): could not find original field %s for field %s",
				originalFieldName, encodedFieldT.Name)
		}

		// The original field may be unexported, so we may need to get an exported copy.
		originalFieldV = getField(originalFieldV)

		var fromFieldV, toFieldV reflect.Value
		if isEncode {
			fromFieldV, toFieldV = originalFieldV, encodedFieldV
		} else {
			fromFieldV, toFieldV = encodedFieldV, originalFieldV
		}

		if err := copyFn(fromFieldV, toFieldV); err != nil {
			return fmt.Errorf("copyStruct(): %w", err)
		}
	}

	return nil
}
