# CSLQ Encoding

CSLQ is a very simple binary encoding with minimal syntax. Binary layout is described using a pattern
string. Byte order is big endian.

## Pattern format

Pattern syntax uses the following characters:

| Char  | Type                                     |
|-------|:-----------------------------------------|
| c     | uint8/byte/char                          |
| s     | uint16                                   |
| l     | uint32                                   |
| q     | uint64                                   |
| v     | custom Marshaler/Unmarshaler             |
| [x]y  | array of length x and elements of type y |
| {...} | structure                                |
| <...> | const                                    |

Whitespace characters (space, tab, newline) are ignored and can be used for visual formatting.

### Arrays

Arrays contain a fixed or dynamic number of elements of a single type. To describe an array
of 32 bytes, use:

`[32]c`

Since the array is always exactly 32 bytes long, its length doesn't need to be encoded. To describe
a dynamic length array, use one of uint types to describe how the length should be encoded. For example:

`[s]c`

Describes a dynamic array of bytes with its length encoded as an uint16 prefix. Array elements can be
any basic or complex type. For example, an array of strings can be described as:

`[s][s]c`

Or an array of structs:

`[s]{qq}`

You can use dynamic arrays of bytes to represent strings:

`[s]c`

This will correctly encode a string or decode to a *string.

### Structs

Structs represent an ordered group of types and are enclosed by curly brackets. They can be used
to describe an array of complex types or just to describe native structs. For example, this pattern:

`{qq}`

can be used to describe how to decode/encode a Go struct consisting of 2 64-bit integers, for example:

````go
type Coords struct {
	X uint64
	Y uint64
	SkipExample   `cslq:"skip"`   // this field will be ignored by the Encoder/Decoder
}
````

### Const

You can mark certain pattern fragments as constant values that will be written on Encode and verified on Decode:

````go
func main() {
	var buf = &bytes.Buffer{}
	var a, b int

	cslq.Encode(buf, "<[6]s>ss", "HEADER", 2, 3)

	r := bytes.NewReader(buf.Bytes())

	cslq.Decode(r, "<[6]s>ss", "HEADER", &a, &b)
}
````

The decoder will read the bytes matching the pattern between `<` and `>` and make sure it matches the respective
value.

## Examples

### Encode basic integers

````golang
package main

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"os"
)

func main() {
	cslq.Encode(os.Stdout, "cslq", 21, 420, 100000000, 1<<40)
}
````

Output:

````
$ go run example | hexdump -C
00000000  15 01 a4 05 f5 e1 00 00  00 01 00 00 00 00 00     |...............|
````

### Encode an array of 64-bit values and 16-bit length

````golang
package main

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"os"
)

func main() {
	cslq.Encode(os.Stdout, "[s]q", []int{21, 420, 100000000, 1<<40})
}
````

Output:

````
$ go run example | hexdump -C
00000000  00 04 00 00 00 00 00 00  00 15 00 00 00 00 00 00  |................|
00000010  01 a4 00 00 00 00 05 f5  e1 00 00 00 01 00 00 00  |................|
00000020  00 00                                             |..|
````

### Encode an array of coordinates

````golang
package main

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"os"
)

type Coords struct {
	X uint64
	Y uint64
}

func main() {
	cslq.Encode(os.Stdout, "[c]{qq}", []Coords{{98, 105}, {116, 99}, {111, 105}, {110, 100}})
}
````

Output:

````
$ go run example | hexdump -C
00000000  04 00 00 00 00 00 00 00  62 00 00 00 00 00 00 00  |........b.......|
00000010  69 00 00 00 00 00 00 00  74 00 00 00 00 00 00 00  |i.......t.......|
00000020  63 00 00 00 00 00 00 00  6f 00 00 00 00 00 00 00  |c.......o.......|
00000030  69 00 00 00 00 00 00 00  6e 00 00 00 00 00 00 00  |i.......n.......|
00000040  64                                                |d|
````

### Decoding

Decoding syntax is similar, just use pointers to vars where the decoded values should be stored (very
similar to how Go's Marshal/Unmarshal work). This decodes the combines output of the 3 encoding examples:

````golang
package main

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"os"
)

type Coords struct {
	X uint64
	Y uint64
}

func main() {
	var (
		u8     uint8
		u16    uint16
		u32    uint32
		u64    uint64
		ints   []int
		coords []Coords
	)
	err := cslq.Decode(os.Stdin, "cslq [s]q [c]{qq}", &u8, &u16, &u32, &u64, &ints, &coords)
	if err != nil {
		fmt.Println("decode error:", err)
	}
}
````

### Custom Marshaler/Unmarshaler

You can implement the Marshaler/Unmarshaler interfaces to include encoding logic directly into Go structs.


````golang
package main

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"os"
)

type Coords struct {
	X uint64
	Y uint64
}

func (c *Coords) UnmarshalCSLQ(dec *cslq.Decoder) error {
	return dec.Decode("qq", &c.X, &c.Y)
}

func (c Coords) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encode("qq", c.X, c.Y)
}

func main() {
	err := cslq.Encode(os.Stdout, "[s]v", []Coords{{98, 105}, {116, 99}, {111, 105}, {110, 100}})
	if err != nil {
		fmt.Println(err)
	}
}

func decode() error {
	var coords []Coords
	err := cslq.Decode(os.Stdin, "[s]v", &coords)
	if err != nil {
		fmt.Println(err)
	}
}
````