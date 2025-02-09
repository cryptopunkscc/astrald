package astral

import (
	"encoding/binary"
	"io"
	"time"
)

type Duration time.Duration

func (Duration) ObjectType() string {
	return "astral.duration"
}

func (d Duration) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, binary.BigEndian, int64(d))
	if err == nil {
		n += 8
	}
	return
}

func (d *Duration) ReadFrom(r io.Reader) (n int64, err error) {
	var i int64
	err = binary.Read(r, binary.BigEndian, &i)
	if err != nil {
		return
	}
	n += 8
	*d = Duration(i)
	return
}

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

func (d Duration) String() string {
	return time.Duration(d).String()
}

func (d *Duration) UnmarshalText(text []byte) (err error) {
	*(*time.Duration)(d), err = time.ParseDuration(string(text))
	return
}

func (d Duration) MarshalText() (text []byte, err error) {
	return []byte(d.String()), nil
}
