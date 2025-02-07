package unsafely

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests that recursive structures are supported, but cyclic data is not.
func TestMarshalJSON_DetectCycles(t *testing.T) {
	type node struct {
		next *node
	}

	t.Run("cycle", func(t *testing.T) {
		var (
			c = &node{}
			b = &node{next: c}
			a = &node{next: b}
		)
		c.next = a

		_, err := MarshalJSON(a)
		require.ErrorContains(t, err, "cycle detected")
	})

	t.Run("linear", func(t *testing.T) {
		var (
			c = &node{}
			b = &node{next: c}
			a = &node{next: b}
		)
		out, err := MarshalJSON(a, WithIndent("  "))
		require.NoError(t, err)
		assert.JSONEq(t, `
{
  "value": {
		"pointer": 4,
		"value": {
			"next": {
				"pointer": 3,
				"value": {
					"next": {
						"pointer": 2,
						"value": {
							"next": {
								"pointer": 1,
								"value": null
							}
						}
					}
				}
			}
		}
  }
}`,
			string(out))

		var decoded *node
		require.NoError(t, UnmarshalJSON(out, &decoded))
		assert.Equal(t, a, decoded)
	})
}
