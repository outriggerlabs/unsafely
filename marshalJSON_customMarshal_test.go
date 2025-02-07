package unsafely

import (
	"encoding/json"
	"strings"
	"testing"
)

// Tests that custom marshaling and unmarshaling is supported, and that we can
// work around ineligible fields using custom marshaling.
func TestMarshalJSON_CustomMarshal(t *testing.T) {
	type embeddedCustomJSON struct {
		customJSON
	}

	var tests = map[string]any{
		"customJSON": customJSON{value: "one"},
		"embedded":   embeddedCustomJSON{customJSON{value: "two"}},
	}

	testMarshalJSON(t, tests)
}

type customJSON struct {
	value       string
	unsupported func() // unsupported field to verify custom marshaling
}

func (s *customJSON) UnmarshalJSON(bytes []byte) error {
	var marshalled [2]string
	if err := json.Unmarshal(bytes, &marshalled); err != nil {
		return err
	}
	s.value = marshalled[0]
	return nil
}

func (s customJSON) MarshalJSON() ([]byte, error) {
	// Marshals the value to an array.
	marshalled := [2]string{s.value, strings.ToUpper(s.value)}
	return json.Marshal(marshalled)
}
