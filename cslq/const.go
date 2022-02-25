package cslq

import "encoding/binary"

// tagCSLQ is the field's tag that will be used for options
const tagCSLQ = "cslq"

// tagSkip defines the keyword that marks a field that should not be encoded/decoded
const tagSkip = "skip"

var byteOrder = binary.BigEndian
