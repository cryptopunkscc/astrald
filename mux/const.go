package mux

import (
	"math"
)

// MaxFrameSize is the maximum total size of a mux frame including the header
const MaxFrameSize = math.MaxUint16

// headerLen is the length of the frame header (uint16 for streamID and uint16 for payload size)
const headerLen = 4

// MaxPayload maximum payload size a single mux frame can carry
const MaxPayload = MaxFrameSize - headerLen

// MaxStreams number of streams in the multiplexer
const MaxStreams = 1 << 16

const controlStreamID = 0
