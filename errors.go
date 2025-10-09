package unsafely

import (
	"runtime"
	"strings"

	perr "github.com/outriggerlabs/errors"
)

// errProps returns a function that creates a perr.Builder with an "op" property
// automatically derived from the calling function's name.
//
// The "op" property format is "package.FunctionName" or "package.MethodName" for
// methods. This provides consistent operation tracking across the call stack.
//
// Usage:
//
//	func MyFunction() error {
//	    errProps := errProps()
//	    if err != nil {
//	        return errProps().Annotate(err)
//	    }
//	    // ... more code
//	}
func errProps() func() perr.Builder {
	// Capture the caller's program counter at the time errProps is called.
	// Skip 1 frame to get the function that called errProps().
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		// Fallback if we can't determine the caller.
		return func() perr.Builder {
			return perr.NewBuilder().WithProp("op", "unknown")
		}
	}

	// Return a function that lazily evaluates the operation name and creates a builder.
	return func() perr.Builder {
		funcName := runtime.FuncForPC(pc).Name()
		op := formatOpName(funcName)
		return perr.NewBuilder().WithProp("op", op)
	}
}

// formatOpName converts a fully-qualified function name from runtime.FuncForPC
// into a clean "package.FunctionName" format.
//
// Examples:
//   - "github.com/outriggerlabs/unsafely.MarshalJSON" -> "unsafely.MarshalJSON"
//   - "github.com/outriggerlabs/unsafely.(*JSONDecoder).Decode" -> "unsafely.Decode"
//   - "github.com/outriggerlabs/unsafely.(*JSONDecoder).Decode-fm" -> "unsafely.Decode"
func formatOpName(funcName string) string {
	// Remove any closure/wrapper suffixes like "-fm" or ".func1"
	if idx := strings.LastIndex(funcName, "-"); idx != -1 {
		funcName = funcName[:idx]
	}

	// Extract the package name from the full path.
	// e.g., "github.com/outriggerlabs/unsafely.(*JSONDecoder).Decode"
	lastSlash := strings.LastIndex(funcName, "/")
	if lastSlash == -1 {
		lastSlash = 0
	} else {
		lastSlash++ // Move past the slash
	}

	// Get everything after the last slash
	remainder := funcName[lastSlash:]

	// Split on dots to separate package.Type.Method
	parts := strings.Split(remainder, ".")

	if len(parts) == 0 {
		return "unknown"
	}

	// If there's only one part (no dots), return it as-is
	if len(parts) == 1 {
		return parts[0]
	}

	// Extract package name (first part)
	pkgName := parts[0]

	// Find the method/function name (last non-empty part)
	var funcShortName string
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		// Skip empty parts and receiver types like "(*JSONDecoder)"
		if part != "" && !strings.HasPrefix(part, "(*") && !strings.HasPrefix(part, "(") {
			funcShortName = part
			break
		}
	}

	if funcShortName == "" {
		return pkgName
	}

	return pkgName + "." + funcShortName
}

// opName is a convenience function that returns just the operation name string
// for the calling function, without creating a builder.
//
// This can be useful for logging or debugging purposes.
func opName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "unknown"
	}
	funcName := runtime.FuncForPC(pc).Name()
	return formatOpName(funcName)
}
