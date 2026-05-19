package jsonw

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"
)

type encoderTest struct {
	name        string
	fn          func(*Buffer)
	output      string
	expectPanic string
}

var encoderTests = []encoderTest{
	{
		name: "basic",
		fn: func(b *Buffer) {
			b.Object(func() {
				b.Key("abc")
				b.HexUint64(45)
				b.Key("array")
				b.Array(func() {
					b.HexUint64(1)
					b.HexBytes([]byte{1, 2, 3, 4, 5})
				})
			})
		},
		output: `{"abc":"0x2d","array":["0x1","0x0102030405"]}`,
	},
	{
		name: "array-1",
		fn: func(b *Buffer) {
			b.Array(func() {
				b.HexUint64(99)
			})
		},
		output: `["0x63"]`,
	},
	{
		name: "array-2null",
		fn: func(b *Buffer) {
			b.Array(func() {
				b.Null()
				b.Null()
			})
		},
		output: `[null,null]`,
	},
	{
		name: "array-nesting",
		fn: func(b *Buffer) {
			b.Array(func() {
				b.Object(func() {
					b.Key("a")
					b.Array(func() {
						b.Null()
						b.Array(func() {
							b.HexUint64(1)
							b.HexUint64(2)
						})
						b.HexUint64(3)
					})
					b.Key("b")
					b.Array(func() {
						b.HexUint64(4)
						b.HexUint64(5)
					})
				})
				b.Null()
				b.Array(func() {
					b.HexUint64(6)
				})
			})
		},
		output: `[{"a":[null,["0x1","0x2"],"0x3"],"b":["0x4","0x5"]},null,["0x6"]]`,
	},
	{
		name: "null",
		fn: func(b *Buffer) {
			b.Null()
		},
		output: `null`,
	},
	{
		name: "bool-true",
		fn: func(b *Buffer) {
			b.Bool(true)
		},
		output: `true`,
	},
	{
		name: "bool-false",
		fn: func(b *Buffer) {
			b.Bool(false)
		},
		output: `false`,
	},
	{
		name: "bool-as-object-value",
		fn: func(b *Buffer) {
			b.Object(func() {
				b.Key("k")
				b.Bool(true)
			})
		},
		output: `{"k":true}`,
	},
	{
		name: "hex-bytes-empty",
		fn: func(b *Buffer) {
			b.HexBytes(nil)
		},
		output: `"0x"`,
	},
	{
		name: "hex-bytes",
		fn: func(b *Buffer) {
			b.HexBytes([]byte{0x00, 0xff, 0xab, 0xcd})
		},
		output: `"0x00ffabcd"`,
	},
	{
		name: "hex-uint64-zero",
		fn: func(b *Buffer) {
			b.HexUint64(0)
		},
		output: `"0x0"`,
	},
	{
		name: "hex-uint64-max",
		fn: func(b *Buffer) {
			b.HexUint64(math.MaxUint64)
		},
		output: `"0xffffffffffffffff"`,
	},
	{
		name: "uint64",
		fn: func(b *Buffer) {
			b.Uint64(0)
		},
		output: `0`,
	},
	{
		name: "uint64-max",
		fn: func(b *Buffer) {
			b.Uint64(math.MaxUint64)
		},
		output: `18446744073709551615`,
	},
	{
		name: "int64-positive",
		fn: func(b *Buffer) {
			b.Int64(42)
		},
		output: `42`,
	},
	{
		name: "int64-negative",
		fn: func(b *Buffer) {
			b.Int64(-7)
		},
		output: `-7`,
	},
	{
		name: "int64-min",
		fn: func(b *Buffer) {
			b.Int64(math.MinInt64)
		},
		output: `-9223372036854775808`,
	},
	{
		name: "float64-zero",
		fn: func(b *Buffer) {
			b.Float64(0)
		},
		output: `0`,
	},
	{
		name: "float64-positive",
		fn: func(b *Buffer) {
			b.Float64(3.14)
		},
		output: `3.14`,
	},
	{
		name: "float64-negative",
		fn: func(b *Buffer) {
			b.Float64(-2.5)
		},
		output: `-2.5`,
	},
	{
		name: "float64-integer-value",
		fn: func(b *Buffer) {
			b.Float64(42)
		},
		output: `42`,
	},
	{
		name: "float64-small",
		fn: func(b *Buffer) {
			b.Float64(0.0001)
		},
		output: `0.0001`,
	},
	{
		name: "bigint",
		fn: func(b *Buffer) {
			v, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
			b.BigInt(v)
		},
		output: `123456789012345678901234567890`,
	},
	{
		name: "bigint-negative",
		fn: func(b *Buffer) {
			b.BigInt(big.NewInt(-42))
		},
		output: `-42`,
	},
	{
		name: "hex-bigint",
		fn: func(b *Buffer) {
			v, _ := new(big.Int).SetString("ffeeddccbbaa99887766", 16)
			b.HexBigInt(v)
		},
		output: `"0xffeeddccbbaa99887766"`,
	},
	{
		name: "hex-bigint-zero",
		fn: func(b *Buffer) {
			b.HexBigInt(big.NewInt(0))
		},
		output: `"0x0"`,
	},
	{
		name: "value-object",
		fn: func(b *Buffer) {
			b.MustValue(map[string]int{"a": 1})
		},
		output: `{"a":1}`,
	},
	{
		name: "value-in-array",
		fn: func(b *Buffer) {
			b.Array(func() {
				b.MustValue(1)
				b.MustValue("x")
				b.MustValue([]int{2, 3})
			})
		},
		output: `[1,"x",[2,3]]`,
	},
	{
		name: "value-error",
		fn: func(b *Buffer) {
			err := b.Value(make(chan int))
			if err == nil {
				panic("expected encode error")
			}
		},
		output: ``,
	},
	{
		name: "must-value-panic",
		fn: func(b *Buffer) {
			b.MustValue(make(chan int))
		},
		expectPanic: "json: unsupported type: chan int",
	},
	{
		name: "key-escaping",
		fn: func(b *Buffer) {
			b.Object(func() {
				b.Key(`a"b\c`)
				b.HexUint64(1)
				b.Key("unicode-€")
				b.HexUint64(2)
			})
		},
		output: `{"a\"b\\c":"0x1","unicode-€":"0x2"}`,
	},
	{
		name:   "string-empty",
		fn:     func(b *Buffer) { b.String("") },
		output: `""`,
	},
	{
		name:   "string-ascii",
		fn:     func(b *Buffer) { b.String("hello world") },
		output: `"hello world"`,
	},
	{
		name:   "string-quote-and-backslash",
		fn:     func(b *Buffer) { b.String(`a"b\c`) },
		output: `"a\"b\\c"`,
	},
	{
		name:   "string-short-escapes",
		fn:     func(b *Buffer) { b.String("\b\t\n\f\r") },
		output: `"\b\t\n\f\r"`,
	},
	{
		name:   "string-control-chars",
		fn:     func(b *Buffer) { b.String("\x00\x01\x1f") },
		output: `"\u0000\u0001\u001f"`,
	},
	{
		name: "string-all-controls",
		fn: func(b *Buffer) {
			var s [32]byte
			for i := range s {
				s[i] = byte(i)
			}
			b.String(string(s[:]))
		},
		output: `"\u0000\u0001\u0002\u0003\u0004\u0005\u0006\u0007\b\t\n\u000b\f\r\u000e\u000f\u0010\u0011\u0012\u0013\u0014\u0015\u0016\u0017\u0018\u0019\u001a\u001b\u001c\u001d\u001e\u001f"`,
	},
	{
		name:   "string-del-is-not-escaped",
		fn:     func(b *Buffer) { b.String("\x7f") },
		output: "\"\x7f\"",
	},
	{
		name:   "string-utf8-multibyte",
		fn:     func(b *Buffer) { b.String("naïve—€—\U0001F600") },
		output: "\"naïve—€—\U0001F600\"",
	},
	{
		name:   "string-invalid-utf8",
		fn:     func(b *Buffer) { b.String("\xff\xfe") },
		output: `"\ufffd\ufffd"`,
	},
	{
		name:   "string-mixed-invalid-utf8",
		fn:     func(b *Buffer) { b.String("ok\xffmid\xfe end") },
		output: `"ok\ufffdmid\ufffd end"`,
	},
	{
		name:   "string-long-safe-run",
		fn:     func(b *Buffer) { b.String("the quick brown fox jumps over the lazy dog 0123456789") },
		output: `"the quick brown fox jumps over the lazy dog 0123456789"`,
	},
	{
		name:   "string-long-with-escapes-and-utf8",
		fn:     func(b *Buffer) { b.String("aaaaaaaa\"bbbbbbbb\\cccccccc\nddddddddé€eeeeeeee") },
		output: `"aaaaaaaa\"bbbbbbbb\\cccccccc\nddddddddé€eeeeeeee"`,
	},
	{
		name: "string-in-array",
		fn: func(b *Buffer) {
			b.Array(func() {
				b.String("a")
				b.String("b")
				b.Null()
				b.String("c")
			})
		},
		output: `["a","b",null,"c"]`,
	},
	{
		name: "string-as-object-value",
		fn: func(b *Buffer) {
			b.Object(func() {
				b.Key("k")
				b.String("v\nv")
			})
		},
		output: `{"k":"v\nv"}`,
	},
	{
		name: "raw",
		fn: func(b *Buffer) {
			b.Object(func() {
				b.Key("a")
				b.RawValue([]byte(`{"a":1}` + "\n"))
				b.Key("b")
				b.RawValue([]byte("  2"))
			})
		},
		output: `{"a":{"a":1},"b":2}`,
	},
	{
		name: "empty-object",
		fn: func(b *Buffer) {
			b.Object(func() {})
		},
		output: `{}`,
	},
	{
		name: "empty-array",
		fn: func(b *Buffer) {
			b.Array(func() {})
		},
		output: `[]`,
	},
	{
		name: "object-value-panic",
		fn: func(b *Buffer) {
			b.Object(func() {
				b.Array(func() {
					b.HexUint64(99)
				})
			})
		},
		expectPanic: "writing value when object key expected",
	},
	{
		name: "object-key-panic",
		fn: func(b *Buffer) {
			b.Object(func() {
				b.Key("ab")
				b.Key("ac")
			})
		},
		expectPanic: "writing key when value expected",
	},
}

func TestReset(t *testing.T) {
	var b Buffer
	b.Array(func() {
		b.HexUint64(1)
		b.HexUint64(2)
	})
	if got := string(b.Output()); got != `["0x1","0x2"]` {
		t.Fatalf("unexpected output before reset: %s", got)
	}
	b.Reset()
	if got := string(b.Output()); got != `` {
		t.Fatalf("buffer not empty after reset: %s", got)
	}
	b.Object(func() {
		b.Key("k")
		b.Int64(7)
	})
	if got := string(b.Output()); got != `{"k":7}` {
		t.Fatalf("unexpected output after reset: %s", got)
	}
}

func TestOutput(t *testing.T) {
	for _, test := range encoderTests {
		t.Run(test.name, func(t *testing.T) {
			var b Buffer
			if test.expectPanic != "" {
				defer func() {
					p := recover()
					if p == nil {
						t.Error("expected panic but test ran successfully")
						t.Log("output: " + string(b.Output()))
					} else {
						msg := fmt.Sprint(p)
						if msg != test.expectPanic {
							t.Error("wrong panic message: " + msg)
						}
					}
				}()
			}
			test.fn(&b)
			output := string(b.Output())
			if output != test.output {
				t.Error("wrong output: " + output)
			}
		})
	}
}

// TestStringEscapeBoundary places an escape byte at every offset within a
// SWAR window to exercise the boundary between the 8-byte chunk loop and the
// scalar tail.
func TestStringEscapeBoundary(t *testing.T) {
	for _, esc := range []struct {
		in, out string
	}{
		{"\"", `\"`},
		{"\\", `\\`},
		{"\n", `\n`},
		{"\x01", `\u0001`},
	} {
		for off := 0; off <= 24; off++ {
			input := strings.Repeat("a", off) + esc.in + strings.Repeat("b", 24-off)
			want := `"` + strings.Repeat("a", off) + esc.out + strings.Repeat("b", 24-off) + `"`
			var b Buffer
			b.String(input)
			got := string(b.Output())
			if got != want {
				t.Errorf("offset %d, escape %q: got %s, want %s", off, esc.in, got, want)
			}
		}
	}
}

// TestStringMatchesEncodingJSON cross-checks our output against encoding/json
// (with HTML escaping disabled) for a wide range of inputs. RFC 8259 leaves
// some room for choice (e.g. which short forms to use), but on the inputs
// below the two implementations should agree byte-for-byte.
func TestStringMatchesEncodingJSON(t *testing.T) {
	inputs := []string{
		"",
		"hello",
		"\"\\",
		"\b\t\n\f\r",
		"\x00\x01\x02\x1e\x1f",
		"\x7f",
		"ünïcødé",
		"€",
		"\U0001F600",
		"mix \"quote\" and\nnewline and\tunicode €",
		strings.Repeat("a", 64),
		strings.Repeat("a", 64) + "\n",
		strings.Repeat("a", 64) + "\xff",
		"\xff\xfe\xfd",
		"a\xffb",
	}
	for _, in := range inputs {
		var b Buffer
		b.String(in)
		got := string(b.Output())
		want := encodingJSONString(t, in)
		if got != want {
			t.Errorf("input %q:\n got: %s\nwant: %s", in, got, want)
		}
	}
}

func encodingJSONString(t *testing.T, s string) string {
	t.Helper()
	var buf strings.Builder
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		t.Fatal(err)
	}
	return strings.TrimRight(buf.String(), "\n")
}
