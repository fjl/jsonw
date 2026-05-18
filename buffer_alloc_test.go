package jsonw

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

var benchStrings = map[string]string{
	"ASCIIShort":  "hello world",
	"ASCIILong":   strings.Repeat("the quick brown fox jumps over the lazy dog ", 8),
	"WithEscapes": `key="value" with\backslashes` + "\nand\ta tab",
	"AllUnicode":  strings.Repeat("naïve—€—\U0001F600 ", 16),
	"Mixed":       strings.Repeat("ascii ", 8) + "naïve—€\n" + strings.Repeat("more ascii ", 8),
}

// These benchmarks verify that Array and Object do not allocate.
// The closures passed to them must not escape to the heap.

func TestNoAllocs(t *testing.T) {
	var buf Buffer
	// Warm up the buffer so growth allocations don't count.
	buf.Object(func() {
		buf.Key("arr")
		buf.Array(func() {
			buf.HexUint64(1)
			buf.HexUint64(2)
		})
		buf.Key("obj")
		buf.Object(func() {
			buf.Key("x")
			buf.Int64(7)
		})
	})
	// Make sure the backing buffer is large enough for every string case.
	for _, s := range benchStrings {
		buf.Reset()
		buf.String(s)
	}

	cases := []struct {
		name string
		fn   func()
	}{
		{"ArrayEmpty", func() { buf.Reset(); buf.Array(func() {}) }},
		{"ArrayElems", func() {
			buf.Reset()
			buf.Array(func() {
				buf.HexUint64(1)
				buf.HexUint64(2)
				buf.HexUint64(3)
			})
		}},
		{"ObjectEmpty", func() { buf.Reset(); buf.Object(func() {}) }},
		{"ObjectKeyValues", func() {
			buf.Reset()
			buf.Object(func() {
				buf.Key("a")
				buf.Int64(1)
				buf.Key("b")
				buf.Int64(2)
			})
		}},
		{"Nested", func() {
			buf.Reset()
			buf.Object(func() {
				buf.Key("arr")
				buf.Array(func() {
					buf.HexUint64(1)
					buf.HexUint64(2)
				})
				buf.Key("obj")
				buf.Object(func() {
					buf.Key("x")
					buf.Int64(7)
				})
			})
		}},
		{"StringASCIIShort", func() { buf.Reset(); buf.String(benchStrings["ASCIIShort"]) }},
		{"StringASCIILong", func() { buf.Reset(); buf.String(benchStrings["ASCIILong"]) }},
		{"StringWithEscapes", func() { buf.Reset(); buf.String(benchStrings["WithEscapes"]) }},
		{"StringUnicode", func() { buf.Reset(); buf.String(benchStrings["AllUnicode"]) }},
		{"StringMixed", func() { buf.Reset(); buf.String(benchStrings["Mixed"]) }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if n := testing.AllocsPerRun(100, c.fn); n != 0 {
				t.Errorf("got %v allocs/op, want 0", n)
			}
		})
	}
}

func BenchmarkArrayEmpty(b *testing.B) {
	var buf Buffer
	b.ReportAllocs()
	for b.Loop() {
		buf.Reset()
		buf.Array(func() {})
	}
}

func BenchmarkArrayElems(b *testing.B) {
	var buf Buffer
	b.ReportAllocs()
	for b.Loop() {
		buf.Reset()
		buf.Array(func() {
			buf.HexUint64(1)
			buf.HexUint64(2)
			buf.HexUint64(3)
		})
	}
}

func BenchmarkObjectEmpty(b *testing.B) {
	var buf Buffer
	b.ReportAllocs()
	for b.Loop() {
		buf.Reset()
		buf.Object(func() {})
	}
}

func BenchmarkObjectKeyValues(b *testing.B) {
	var buf Buffer
	b.ReportAllocs()
	for b.Loop() {
		buf.Reset()
		buf.Object(func() {
			buf.Key("a")
			buf.Int64(1)
			buf.Key("b")
			buf.Int64(2)
		})
	}
}

func BenchmarkNested(b *testing.B) {
	var buf Buffer
	b.ReportAllocs()
	for b.Loop() {
		buf.Reset()
		buf.Object(func() {
			buf.Key("arr")
			buf.Array(func() {
				buf.HexUint64(1)
				buf.HexUint64(2)
			})
			buf.Key("obj")
			buf.Object(func() {
				buf.Key("x")
				buf.Int64(7)
			})
		})
	}
}

func BenchmarkString(b *testing.B) {
	for name, s := range benchStrings {
		b.Run(name, func(b *testing.B) {
			var buf Buffer
			buf.String(s) // warm up backing buffer
			b.SetBytes(int64(len(s)))
			b.ReportAllocs()
			for b.Loop() {
				buf.Reset()
				buf.String(s)
			}
		})
	}
}

// BenchmarkStringStdlib runs encoding/json's string encoder on the same
// inputs, as a baseline for comparing against BenchmarkString.
func BenchmarkStringStdlib(b *testing.B) {
	for name, s := range benchStrings {
		b.Run(name, func(b *testing.B) {
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			enc.SetEscapeHTML(false)
			enc.Encode(s) // warm up
			b.SetBytes(int64(len(s)))
			b.ReportAllocs()
			for b.Loop() {
				buf.Reset()
				enc.Encode(s)
			}
		})
	}
}
