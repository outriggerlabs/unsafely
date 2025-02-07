package unsafely

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	jsonMarshalerType  = reflect.TypeFor[json.Marshaler]()
	jsonRawMessageType = reflect.TypeFor[json.RawMessage]()

	stringType = reflect.TypeFor[string]()

	// Map from decoded types to encoded types.
	typeCache = make(map[reflect.Type]reflect.Type)
)

// Returns an encoded type for the input type, possibly from the cache.
func encodedTypeFor(inputT reflect.Type) (reflect.Type, error) {
	if encodedT, ok := typeCache[inputT]; ok {
		return encodedT, nil
	}

	encodedT, err := createEncodedTypeFor(inputT)
	if err != nil {
		return nil, err
	}

	typeCache[inputT] = encodedT
	return encodedT, nil
}

// Dynamically constructs an encoded type with exported fields that mirrors the
// input type.
func createEncodedTypeFor(inputT reflect.Type) (reflect.Type, error) {
	// If the input type is a json.Marshaler, we store the raw JSON output of the
	// existing marshaling behavior.
	if inputT.Implements(jsonMarshalerType) {
		return jsonRawMessageType, nil
	}

	var kind = inputT.Kind()

	// Currently unsupported types.
	if kind == reflect.Chan ||
		kind == reflect.Func ||
		kind == reflect.UnsafePointer {
		return nil, fmt.Errorf("createEncodedTypeFor: unsupported kind %v for %v", inputT.Kind(), inputT)
	}

	// Interfaces are represented using a struct to track the underlying type.
	if kind == reflect.Interface {
		return interfaceValueType, nil
	}

	// Pointers are represented using a struct to track the pointer number and
	// underlying value.
	if kind == reflect.Pointer {
		return pointerValueType, nil
	}

	// Resolve the element type for slices and arrays.
	if kind == reflect.Slice {
		elemType, err := encodedTypeFor(inputT.Elem())
		if err != nil {
			return nil, fmt.Errorf("createEncodedTypeFor: %w", err)
		}

		return reflect.SliceOf(elemType), nil
	}

	if kind == reflect.Array {
		elemType, err := encodedTypeFor(inputT.Elem())
		if err != nil {
			return nil, fmt.Errorf("createEncodedTypeFor: %w", err)
		}

		return reflect.ArrayOf(inputT.Len(), elemType), nil
	}

	// Resolve the key and value element types for maps.
	if kind == reflect.Map {
		keyType, err := encodedTypeFor(inputT.Key())
		if err != nil {
			return nil, fmt.Errorf("createEncodedTypeFor: %w", err)
		}

		valueType, err := encodedTypeFor(inputT.Elem())
		if err != nil {
			return nil, fmt.Errorf("createEncodedTypeFor: %w", err)
		}

		// JSON maps cannot have struct keys, so we use string keys instead.
		if !isSimplePrimitive(keyType.Kind()) {
			keyType = stringType
		}

		return reflect.MapOf(keyType, valueType), nil
	}

	// Complex value types are encoded using complexValue.
	if isComplexNumber(kind) {
		return complexValueType, nil
	}

	// Simple primitive types are returned as-is.
	if isSimplePrimitive(kind) {
		return inputT, nil
	}

	// At this point, we should have handled everything except structs.
	if kind != reflect.Struct {
		return nil, fmt.Errorf(
			"createEncodedTypeFor: unhandled non-struct kind %v of %v",
			kind, inputT,
		)
	}

	// Dynamically create a struct to mirror the input type.
	//
	// Fields are named "F0", "F1", etc., and given a json tag based on:
	// 1. The unsafely.json tag if present
	// 2. The json tag if present
	// 3. The original field name if no tags are present
	var (
		fields        = make([]reflect.StructField, 0, inputT.NumField())
		usedJsonNames = make(map[string]string) // maps json name to field name
	)

	for i := 0; i < inputT.NumField(); i++ {
		var (
			field     = inputT.Field(i)
			fieldName = field.Name
			jsonTag   string
		)

		// Default to the "json" tag, but override with the "unsafely.json" tag.
		if tag := field.Tag.Get("unsafely.json"); tag != "" {
			jsonTag = tag
		} else if tag = field.Tag.Get("json"); tag != "" {
			jsonTag = tag
		}

		if jsonTag == "-" {
			continue
		}

		// Extract the JSON field name from the tag, e.g, json:"key,omitempty"
		//
		// If the JSON field name is empty, we rewrite the tag to add the struct
		// field name; otherwise, we overwrite the JSON name with that
		jsonName, jsonOptions, hasOptions := strings.Cut(jsonTag, ",")
		if jsonName == "" {
			jsonName = fieldName
		}

		if hasOptions {
			jsonTag = jsonName + "," + jsonOptions
		} else {
			jsonTag = jsonName
		}

		// Check for duplicate JSON field names.
		if existingField, exists := usedJsonNames[jsonName]; exists {
			return nil, fmt.Errorf("createEncodedTypeFor(): duplicate JSON field name %q (struct fields %q and %q)",
				jsonName, existingField, field.Name)
		}
		usedJsonNames[jsonName] = field.Name

		newType, err := encodedTypeFor(field.Type)
		if err != nil {
			return nil, fmt.Errorf("createEncodedTypeFor: %w", err)
		}

		field.Type = newType
		field.Name = "F" + strconv.Itoa(i)
		field.PkgPath = "" // mark as exported
		field.Tag = reflect.StructTag(fmt.Sprintf(`json:"%s" original:"%s"`, jsonTag, fieldName))

		fields = append(fields, field)
	}

	return reflect.StructOf(fields), nil
}
