package unsafely

import (
	"reflect"
	"unsafe"
)

// A zero reflect.Value, typically the result of an invalid operation, or
// reflecting a nil value, e.g, reflect.ValueOf(nil).
var zeroValue reflect.Value

// Returns true if the kind is a simple primitive type that doesn't require any
// encoding or decoding.
func isSimplePrimitive(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.String:
		return true
	default:
		return false
	}
}

// Returns true if the kind is a complex type.
func isComplexNumber(kind reflect.Kind) bool {
	return kind == reflect.Complex64 || kind == reflect.Complex128
}

// The following functions allow us access and set the values of unexported
// fields. The following was inspired by https://stackoverflow.com/a/60598827.

// Gets the value of a (possibly unexported) field.
func getField(field reflect.Value) reflect.Value {
	if field.CanSet() {
		return field
	}

	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem()
}

// Sets the value of a (possibly unexported) field.
func setField(field reflect.Value, value reflect.Value) {
	getField(field).Set(value)
}

// Returns an addressable version of the value. The value may be returned as-is
// if it is already addressable.
//
// The value must not be a zero value.
//
// For more info, see https://stackoverflow.com/a/43918797.
func ensureAddressable(v reflect.Value) reflect.Value {
	if v.CanAddr() {
		return v
	}

	vCopy := reflect.New(v.Type()).Elem()
	vCopy.Set(v)
	return vCopy
}
