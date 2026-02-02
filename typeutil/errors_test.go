package typeutil

import (
	"errors"
	"testing"

	perr "github.com/outriggerlabs/errors"
	"github.com/stretchr/testify/assert"
)

func TestErrProps(t *testing.T) {
	var (
		errProps  = errProps()
		baseError = errors.New("test error")

		annotatedError = errProps.Annotate(baseError)
		extendedError  = errProps.Extend(baseError)
	)

	assert.Equal(t, "typeutil.TestErrProps", perr.Property[string](annotatedError, "op"))
	assert.Contains(t, annotatedError.Error(), "test error")

	assert.Equal(t, "typeutil.TestErrProps", perr.Property[string](extendedError, "op"))
	assert.Contains(t, extendedError.Error(), "test error")
}
