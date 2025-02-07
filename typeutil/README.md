# typeutil

Utilities for working with types

## `Resolver`

An interface for querying `reflect.Type` objects by name.

Implementations:

* `ChainResolver`: Returns the first type successfully returned from a chain of `Resolver` objects.
* `StaticResolver`: A manually-specified registry of types.
* `UnsafeResolver`: A (very unsafe) resolver with a registry constructed via go:linkname.
  * Does not work with gccgo or gollvm.
