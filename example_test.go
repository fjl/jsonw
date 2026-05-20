package jsonw_test

import (
	"fmt"

	"github.com/fjl/jsonw"
)

func Example() {
	// A zero value buffer is ready to use.
	var b jsonw.Buffer

	// Encode an object to the buffer.
	b.Object(func() {
		b.Key("a")
		b.Uint64(67)
		b.Key("b")
		b.Array(func() {
			b.String("hello")
		})
	})

	fmt.Println(string(b.Output()))
	// Output: {"a":67,"b":["hello"]}
}
