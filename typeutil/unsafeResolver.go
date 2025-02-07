//go:build !gccgo

// This is a heavily reduced version of https://github.com/modern-go/reflect2
// that only provides reflect.Type querying.
//
// The type map itself has been expanded to include non-struct types so we can
// query type aliases.

package typeutil

import (
	"reflect"
	"sync"
	"unsafe"
)

//go:linkname typelinks reflect.typelinks
func typelinks() (sections []unsafe.Pointer, offset [][]int32)

//go:linkname resolveTypeOff reflect.resolveTypeOff
func resolveTypeOff(rtype unsafe.Pointer, off int32) unsafe.Pointer

var (
	// Protects unsafeResolver.
	loadOnce       sync.Once
	unsafeResolver = NewStaticResolver()
)

func loadGoTypes() {
	var obj any = reflect.TypeOf(0)
	sections, offset := typelinks()
	for i, offs := range offset {
		rodata := sections[i]
		for _, off := range offs {
			(*emptyInterface)(unsafe.Pointer(&obj)).word = resolveTypeOff(rodata, off)
			typ := obj.(reflect.Type)
			if typ.Kind() != reflect.Pointer {
				continue
			}

			unsafeResolver.AddTypes(typ.Elem())
		}
	}
}

type emptyInterface struct {
	typ  unsafe.Pointer
	word unsafe.Pointer
}

// UnsafeResolver returns a type resolver that uses go:linkname to dynamically
// look up reflect.Type definitions.
//
// Given the recent efforts to lock down linkname handling in Go, it's unclear
// whether this will continue to work for long. Use at your own risk.
//
// This does not work with gccgo, and possibly not with gollvm.
//
// References:
// - https://github.com/modern-go/reflect2
// - https://github.com/golang/go/issues/67401
func UnsafeResolver() Resolver {
	loadOnce.Do(loadGoTypes)
	return unsafeResolver
}
