package unsafely

import (
	"encoding/json"
)

// Wrapper that we're using to reserve an extra layer around the value. We may
// need to use this later to embed configuration options for unmarshaling.
//
// Also, the extra layer discourages attempts to use this as a drop-in for
// standard JSON marshaling.
type encodedJSONWrapper struct {
	Value json.RawMessage `json:"value"`
}

// MarshalJSON serializes the value to a JSON string, including unexported
// fields.
//
// See the package notes for restrictions, limitations and options.
func MarshalJSON(in any, options ...MarshalJSONOption) ([]byte, error) {
	return NewJSONEncoder(options...).Encode(in)
}

// Configuration options for MarshalJSON.
type marshalJSONConfig struct {
	prefix string
	indent string
}

// MarshalJSONOption is an option for modifying the behavior of MarshalJSON.
type MarshalJSONOption func(*marshalJSONConfig)

// WithPrefix sets the prefix for the JSON output.
func WithPrefix(prefix string) MarshalJSONOption {
	return func(config *marshalJSONConfig) {
		config.prefix = prefix
	}
}

// WithIndent sets the indent for the JSON output.
func WithIndent(indent string) MarshalJSONOption {
	return func(config *marshalJSONConfig) {
		config.indent = indent
	}
}
