package astral

import (
	"io"
	"time"
)

type Time time.Time

func Now() Time {
	return Time(time.Now())
}

func (t Time) ObjectType() string { return "time" }

func (t Time) WriteTo(w io.Writer) (n int64, err error) {
	return Uint64(t.Time().UTC().UnixNano()).WriteTo(w)
}

func (t *Time) ReadFrom(r io.Reader) (n int64, err error) {
	var nsec Uint64
	n, err = (&nsec).ReadFrom(r)

	if err == nil {
		*t = Time(time.Unix(0, int64(nsec)))
	}

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
	return (*time.Time)(t).UnmarshalText(text)
}

func init() {
	_ = Add(&Time{})
}
