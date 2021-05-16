package mux

// MaxPayloadSize maximum amount of data that can be transferred in a single mux frame.
// TODO: Maybe this should be adjusted so that our frame fits as a single brontide payload?
const MaxPayloadSize = 65535

// ControlStreamID is the stream ID that will be used for control frames.
const ControlStreamID = StreamID(0)

// MinStreamID is the lowest allowed stream ID.
const MinStreamID = StreamID(1)

// MaxStreamID is the highest allowed stream ID.
const MaxStreamID = StreamID(65535)
