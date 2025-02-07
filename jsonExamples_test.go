package unsafely

import "fmt"

// Complex values are encoded, since they are not supported in JSON by default.
func ExampleMarshalJSON_complex() {
	var val = 1 + 2i

	b, err := MarshalJSON(val)
	check(err)
	fmt.Println(string(b))
	// Output: {"value":{"real":1,"imag":2}}
}

// Pointers contain a reference number to determine which pointers originally
// shared the same value.
//
// When unmarshaling, any fields with the same pointer references will also
// share the same value.
//
// Tip: Using a single JSONEncoder to marshal different objects with the same
// pointer will result in the same pointer reference numbers. These can then
// be unmarshaled with the same JSONDecoder to recreate objects that share
// the resulting pointers.
func ExampleMarshalJSON_pointers() {
	var val = 42

	b, err := MarshalJSON(&val)
	check(err)
	fmt.Println(string(b))
	// Output: {"value":{"pointer":1,"value":42}}
}

// Interface values are encoded with type information so that they concrete
// types can be reconstructed during unmarshaling.
//
// Note: Reconstruction during unmarshaling requires a typeutil.Resolver.
func ExampleMarshalJSON_interfaces() {
	type wrapper struct {
		v any
	}
	var val = wrapper{v: 42}

	b, err := MarshalJSON(val)
	check(err)
	fmt.Println(string(b))
	// Output: {"value":{"v":{"typeName":"int","value":42}}}
}

func ExampleMarshalJSON_genericInterface() {
	type generic[T any] struct {
		v T
	}
	type wrapper struct {
		internal any
	}
	var val = wrapper{internal: generic[int]{v: 42}}

	b, err := MarshalJSON(val)
	check(err)
	fmt.Println(string(b))
	// Output: {"value":{"internal":{"pkgPath":"github.com/outriggerlabs/unsafely","typeName":"generic[int]","value":{"v":42}}}}
}

// JSON maps only support primitive keys, so we marshal non-primitive types to
// JSON before using them as struct keys.
//
// (Yes, it can become rather verbose...)
func ExampleMarshalJSON_structMapKeys() {
	type point struct {
		x, y int
	}
	var val = map[point]int{
		{1, 2}: 3,
	}

	b, err := MarshalJSON(val)
	check(err)
	fmt.Println(string(b))
	// Output: {"value":{"{\"x\":1,\"y\":2}":3}}
}

// "json" tags are honoured, but can be overwritten with "unsafely.json" tags.
//
// Unsupported fields can be skipped using `unsafely.json:"-"`.
//
// To "cancel" a json tag, use `unsafely.json:","`.
func ExampleMarshalJSON_structTags() {
	type exampleWithTags struct {
		X    int    `json:"x"`
		Y    int    `json:"y,omitempty" unsafely.json:"y"`
		Temp int    `json:"-" unsafely.json:","` // cancels json tag
		Fn   func() `unsafely.json:"-"`
		Dash string `json:"-,"`
	}

	var val = exampleWithTags{
		X: 1,
		// Y is empty
		Temp: 3,
		Fn:   func() {},
		Dash: "dash",
	}

	b, err := MarshalJSON(val)
	check(err)
	fmt.Println(string(b))
	// Output: {"value":{"x":1,"y":0,"Temp":3,"-":"dash"}}
}

// The MarshalJSON and JSONEncoder.Encode functions support prefixes and
// indentation.
func ExampleMarshalJSON_prefixAndIndent() {
	type point struct {
		x, y int
	}

	type example struct {
		i     int
		p     point
		edges map[point]int
	}

	var val = example{
		i: 1,
		p: point{1, 2},
		edges: map[point]int{
			{2, 3}: 1,
		},
	}

	b, err := MarshalJSON(val, WithIndent("  "), WithPrefix(">>"))
	check(err)
	fmt.Println(string(b))
	// Output: {
	// >>  "value": {
	// >>    "i": 1,
	// >>    "p": {
	// >>      "x": 1,
	// >>      "y": 2
	// >>    },
	// >>    "edges": {
	// >>      "{\"x\":2,\"y\":3}": 1
	// >>    }
	// >>  }
	// >>}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
