package typeutil

import (
	perr "github.com/outriggerlabs/errors"
	"github.com/outriggerlabs/errors/errorprops"
)

// errProps returns a function that creates a perr.Builder with an "op" property
// automatically derived from the calling function's name.
func errProps() perr.Builder {
	return perr.NewBuilder().WithPropEval(errorprops.CallerSkip("op", 2))
}
