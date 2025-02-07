package unsafely

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests we honor "json" tags and can override them with "unsafely.json" tags.
func TestMarshalJSON_StructTags(t *testing.T) {
	tests := map[string]struct {
		inputType   any
		expectJSON  string
		expectError string
	}{
		"no tags": {
			inputType: struct {
				Name  string
				value int
			}{
				Name:  "a",
				value: 1,
			},
			expectJSON: `{
        "value": {
				  "Name": "a",
				  "value": 1
				}
			}`,
		},

		"json tags": {
			inputType: struct {
				X int `json:"x"`
				Y int `json:",omitempty"`
				Z int `json:"z,omitempty"`
			}{
				X: 1,
				Y: 2,
			},
			expectJSON: `{
        "value": {
				  "x": 1,
				  "Y": 2
        }
			}`,
		},
		"override tags": {
			inputType: struct {
				noTag           string
				WithTag         string `json:"jsonTag"`
				OverrideTag     string `json:"originalTag" unsafely.json:"overrideTag"`
				OverrideOptions string `json:",omitempty" unsafely.json:","`
				OverrideNoSkip  string `json:"-" unsafely.json:","`
				OverrideSkip    string `json:",omitempty" unsafely.json:"-"`
				unexported      string `unsafely.json:"last"`
			}{
				noTag:       "1",
				WithTag:     "2",
				OverrideTag: "3",
				// OverrideOptions: removed omitempty
				OverrideNoSkip: "5",
				OverrideSkip:   "6",
				unexported:     "7",
			},
			expectJSON: `{
        "value": {
					"noTag": "1",
					"jsonTag": "2",
					"overrideTag": "3",
					"OverrideOptions": "",
					"OverrideNoSkip": "5",
					"last": "7"
        }
			}`,
		},
		"duplicate json names": {
			inputType: struct {
				Field1 string `json:"same"`
				Field2 string `json:"same"`
			}{},
			expectError: `duplicate JSON field name "same" (struct fields "Field1" and "Field2")`,
		},
		"duplicate json name conflicts with override": {
			inputType: struct {
				Field1 string `json:"same"`
				Field2 string `unsafely.json:"same"`
			}{},
			expectError: `duplicate JSON field name "same" (struct fields "Field1" and "Field2")`,
		},
		"duplicate json names with override": {
			inputType: struct {
				Field1 string `json:"same"`
				same   string `unsafely.json:"different"`
			}{
				Field1: "a",
				same:   "b",
			},
			expectJSON: `{
				"value": {
				  "same": "a",
				  "different": "b" 	
				}
			}`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			encoded, err := MarshalJSON(tt.inputType)
			if tt.expectError != "" {
				require.Contains(t, err.Error(), tt.expectError)
			} else {
				require.NoError(t, err)
				assert.JSONEq(t, tt.expectJSON, string(encoded))
			}
		})
	}
}

// Specifically tests that we can marshal structs that would be otherwise be
// ineligible (because of the function) by skipping the field.
func TestMarshalJSON_StructTags_SkipFields(t *testing.T) {
	type skipFields struct {
		value int
		skip  func() `unsafely.json:"-"`
	}

	tests := map[string]any{
		"skip": skipFields{value: 1},
	}

	testMarshalJSON(t, tests)
}
