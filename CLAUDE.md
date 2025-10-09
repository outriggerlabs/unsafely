# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go package for marshaling and unmarshaling Go values to/from JSON, including unexported fields, interface values, non-primitive map keys, and pointer reference tracking. **This package is intended for testing only, not production use.**

## Commands

### Testing
```bash
# Run all tests
go test ./...

# Run tests in a specific package
go test .
go test ./typeutil

# Run a specific test
go test -run TestMarshalJSON_Interface

# Run tests with verbose output
go test -v ./...
```

### Building
```bash
# Build (this is a library, no main binary)
go build ./...

# Check for compilation errors
go build .
```

### Documentation
```bash
# View package documentation
go doc

# View function documentation
go doc MarshalJSON
```

## Architecture

### Core Encoding/Decoding System

The package implements a custom JSON marshaling system with three main components:

1. **JSONEncoder** (`jsonEncoder.go`): Converts Go values to JSON
   - Maintains `pointerValues` map to track pointer references across the object graph
   - Tracks `pendingPointers` to detect circular references
   - Assigns unique `pointerIndex` numbers to shared pointers

2. **JSONDecoder** (`jsonDecoder.go`): Reconstructs Go values from JSON
   - Maintains `pointerValues` map to recreate shared pointer relationships
   - Uses `typeutil.Resolver` to reconstruct interface types by name

3. **Type Encoding** (`encodedTypeFor.go`): Dynamically creates encoded types
   - Builds mirror types with exported fields for unexported field access
   - Uses `typeCache` to avoid recreating types
   - Transforms interfaces, pointers, structs, maps, slices into encodable forms

### Special Value Types

The encoder uses wrapper structs to preserve Go semantics in JSON:

- **pointerValue** (`pointerValue.go`): Tracks pointer identity and enables pointer sharing
  - Contains `Pointer` field (reference number) and `Value` field (underlying data)
  - Prevents duplicating the same pointer multiple times

- **interfaceValue** (`interfaceValue.go`): Preserves type information for interface fields
  - Stores `PkgPath`, `TypeName`, `TypeString` for type reconstruction
  - Stores `PtrDepth` to handle pointers-to-interfaces correctly
  - JSON-encoded `Value` field contains the actual data

- **complexValue** (`complexValue.go`): Handles complex64/complex128 types
  - Standard JSON doesn't support complex numbers

### Type Resolution System (`typeutil/`)

Provides the `Resolver` interface for looking up `reflect.Type` objects by name, required for unmarshaling interface values.

Implementations:
- **StaticResolver** (`staticResolver.go`): Manual type registry
- **UnsafeResolver** (`unsafeResolver.go`): Uses `go:linkname` to access Go's internal type registry
  - Does NOT work with gccgo or gollvm
  - Uses reflection internals that may break in future Go versions
- **ChainResolver** (`chainResolver.go`): Tries multiple resolvers in sequence

### Public API

- `MarshalJSON(in any, options ...MarshalJSONOption) ([]byte, error)`: Main entry point for encoding
- `UnmarshalJSON(b []byte, outPtr any, options ...UnmarshalJSONOption) error`: Main entry point for decoding
- `NewJSONEncoder(options...)`: Create reusable encoder (shares pointer state across multiple Encode calls)
- `NewJSONDecoder(options...)`: Create reusable decoder (shares pointer state across multiple Decode calls)

Options:
- `WithPrefix(string)`, `WithIndent(string)`: Format JSON output
- `WithTypeResolver(typeutil.Resolver)`: Configure type resolution for interfaces

### JSON Format

All values are wrapped in `encodedJSONWrapper{Value: ...}` to reserve design space and discourage drop-in replacement of standard JSON libraries.

The `unsafely.json` struct tag overrides `json` tag behavior.
