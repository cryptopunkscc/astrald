package astral

import "encoding/binary"

// ByteOrder defines the byte order for data encoding across astral data structures and protocols. It affects
// everything, and changing this value completely breaks compatibility.
var ByteOrder = binary.BigEndian
