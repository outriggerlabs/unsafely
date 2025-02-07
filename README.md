# unsafely

⚠️ This package is only intended for testing, not for production use. ⚠️

This package provides a set of functions that can be used for testing and
analyzing Go types.

## MarshalJSON / UnmarshalJSON

These functions can be used to marshal structs with unexported fields to JSON
and, in many cases, reconstruct the structs from JSON.

For example:

```go
type example struct {
	Int   int
	valid bool
}

b, err := unsafely.MarshalJSON(example{Int: 42, valid: true})
// {"value:{"Int":42,"valid":true}}

var copy example
err := unsafely.UnmarshalJSON(b, &copy)
// copy == example{Int: 42, valid: true}
```

The JSON produced by these functions is *not* intended to be a drop-in
replacement for any other JSON package. The JSON may contain additional
information necessary for reconstructing the structs, and is subject to change.

See the examples in [jsonExamples_test.go](jsonExamples_test.go) for how special
cases are handled.

Intended use cases:

- Analyzing the internal representation of a struct.
- Generating a "golden version" of a struct state, and comparing it to future
versions.

Features:

- Unexported fields are generally supported.
- Most types are supported.
- Values in interfaces can be reconstructed (with some limitations).
  - This requires a `typeutil.Resolver`. See the [typeutil](typeutil) package.
- Non-primitive map keys are supported (by marshaling these to JSON).
- Complex values are supported.
- Pointer semantics are generally preserved, e.g, if the struct contains
  two copies of a pointer, the unmarshaled copy will also be a struct with
  two copies of a pointer.
- `json` tag behavior can be overridden with `unsafely.json`.
  - Tip: To cancel out the `json` tag, use `unsafely.json:","`. 
- Supports adding prefixes and indents to the JSON output.

Limitations:

- Functions, channels and unsafe.Pointer are unsupported.
- The `typeutil.UnsafeResolver` does not work with gccgo (and probably not gollvm).
- Type resolution may fail if there are two types with the same package path,
name and string representation, e.g, two structs with the same name defined in
different functions.
- None of the functions in this package are concurrency-safe.
