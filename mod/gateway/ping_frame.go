package gateway

import "io"

const (
	BytePing        = byte(0x00)
	BytePong        = byte(0x01)
	ByteSignalGo    = byte(0x02)
	ByteSignalReady = byte(0x03)
)

func WritePing(w io.Writer) error        { _, err := w.Write([]byte{BytePing}); return err }
func WritePong(w io.Writer) error        { _, err := w.Write([]byte{BytePong}); return err }
func WriteSignalGo(w io.Writer) error    { _, err := w.Write([]byte{ByteSignalGo}); return err }
func WriteSignalReady(w io.Writer) error { _, err := w.Write([]byte{ByteSignalReady}); return err }
