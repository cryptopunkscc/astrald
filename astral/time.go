package astral

import (
	"encoding/binary"
	"io"
	"time"
)

type Time time.Time

func Now() Time {
	return Time(time.Now())
}

func (t Time) ObjectType() string { return "astral.time" }

func (t Time) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, binary.BigEndian, int64(t.Time().UTC().UnixNano()))
	if err == nil {
		n += 8
	}
	return
}

func (t *Time) ReadFrom(r io.Reader) (n int64, err error) {
	var i int64
	err = binary.Read(r, binary.BigEndian, &i)
	if err != nil {
		return
	}
	n += 8
	*t = Time(time.Unix(0, i).UTC())
	return
}

func (t Time) Time() time.Time {
	return time.Time(t).UTC()
}

func (t Time) String() string {
	return t.Time().String()
}

func (t Time) MarshalJSON() ([]byte, error) {
	return t.Time().MarshalJSON()
}

func (t *Time) UnmarshalJSON(bytes []byte) (err error) {
	var tt time.Time
	err = tt.UnmarshalJSON(bytes)
	*t = Time(tt)
	return
}

func (t Time) MarshalText() (text []byte, err error) { return t.Time().MarshalText() }

func (t *Time) UnmarshalText(text []byte) (err error) {
	var tt time.Time
	err = tt.UnmarshalText(text)
	*t = Time(tt)
	return
}

func init() {
	DefaultBlueprints.Add(&Time{})
}
