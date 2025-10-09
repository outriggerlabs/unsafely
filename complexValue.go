package unsafely

import (
	"fmt"
	"reflect"
)

var complexValueType = reflect.TypeFor[complexValue]()

type complexValue struct {
	Real float64 `json:"real"`
	Imag float64 `json:"imag"`
}

// Returns true if the provided type is a complexValue.
func isComplexValueType(t reflect.Type) bool {
	return t == complexValueType
}

// Encodes the value to a complexValue object.
func encodeToComplexValue(inputV reflect.Value) (reflect.Value, error) {
	errProps := errProps()

	if inputV.Kind() != reflect.Complex128 && inputV.Kind() != reflect.Complex64 {
		return reflect.Value{}, errProps.Annotate(fmt.Errorf(
			"expected value to be a complex number; received %v", inputV.Kind(),
		))
	}
	var (
		input = inputV.Complex()
		cv    = complexValue{
			Real: real(input),
			Imag: imag(input),
		}
	)

	return reflect.ValueOf(cv), nil
}

// Decodes the complexValue object and writes the output value.
func decodeFromComplexValue(encodedV reflect.Value, outV reflect.Value) error {
	errProps := errProps()

	cv, ok := encodedV.Interface().(complexValue)
	if !ok {
		return errProps.Annotate(fmt.Errorf(
			"expected encodedV to be a complexValue; received %T",
			encodedV.Interface(),
		))
	}

	decoded := complex(cv.Real, cv.Imag)

	switch outV.Kind() {
	case reflect.Complex128:
		setField(outV, reflect.ValueOf(decoded))
		return nil

	case reflect.Complex64:
		setField(outV, reflect.ValueOf(complex64(decoded)))
		return nil

	default:
		return errProps.Annotate(fmt.Errorf("unsupported kind for outV: %v", outV.Kind()))
	}
}
