package astral

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"time"
)

type Duration time.Duration

func (Duration) ObjectType() string {
	return "duration"
}

func (d Duration) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, ByteOrder, int64(d))
	if err == nil {
		n += 8
	}
	return
}

// binary

func (d *Duration) ReadFrom(r io.Reader) (n int64, err error) {
	var i int64
	err = binary.Read(r, ByteOrder, &i)
	if err != nil {
		return
	}
	n += 8
	*d = Duration(i)
	return
}

// json

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration())
}

func (d Duration) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, (*time.Duration)(&d))
	return err
}

// text

func (d Duration) MarshalText() (text []byte, err error) {
	return []byte(d.String()), nil
}

func (d *Duration) UnmarshalText(text []byte) (err error) {
	*(*time.Duration)(d), err = time.ParseDuration(string(text))
	return
}

// ...

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

func (d Duration) String() string {
	return time.Duration(d).String()
}

func init() {
	var d Duration
	_ = Add(&d)
}
