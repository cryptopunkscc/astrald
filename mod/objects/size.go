package objects

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Size uint64

const (
	B   Size = 1
	KB       = B * 1000
	MB       = KB * 1000
	GB       = MB * 1000
	TB       = GB * 1000
	PB       = TB * 1000
	EB       = PB * 1000
	KiB      = B << 10
	MiB      = KB << 10
	GiB      = MB << 10
	TiB      = GB << 10
	PiB      = TB << 10
	EiB      = PB << 10
)

func ParseSize(s string) (Size, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, errors.New("empty string")
	}

	var multiplier uint64 = 1000 // default decimal
	if strings.HasSuffix(s, "ib") {
		multiplier = 1024
		s = strings.TrimSuffix(s, "ib")
	} else if strings.HasSuffix(s, "b") {
		s = strings.TrimSuffix(s, "b")
	}

	var unit string
	switch {
	case strings.HasSuffix(s, "e"):
		unit = "e"
		multiplier *= 1000 * 1000 * 1000 * 1000 * 1000
	case strings.HasSuffix(s, "p"):
		unit = "p"
		multiplier *= 1000 * 1000 * 1000 * 1000
	case strings.HasSuffix(s, "t"):
		unit = "t"
		multiplier *= 1000 * 1000 * 1000
	case strings.HasSuffix(s, "g"):
		unit = "g"
		multiplier *= 1000 * 1000
	case strings.HasSuffix(s, "m"):
		unit = "m"
		multiplier *= 1000
	case strings.HasSuffix(s, "k"):
		unit = "k"
		multiplier *= 1
	default:
		unit = "" // no unit, just bytes
	}

	if unit != "" {
		s = strings.TrimSuffix(s, unit)
	}

	// Parse the numeric part (supports float for cases like "1.5MB")
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, err
	}
	if f < 0 {
		return 0, errors.New("negative size")
	}

	bytes := uint64(f * float64(multiplier))
	if float64(bytes)/float64(multiplier) != f {
		return 0, errors.New("overflow or precision loss")
	}
	return Size(bytes), nil
}

func (s Size) Bytes() uint64 {
	return uint64(s)
}

func (s Size) KBytes() float64 {
	v := s / KB
	r := s % KB
	return float64(v) + float64(r)/float64(KB)
}

func (s Size) KiBytes() float64 {
	v := s / KiB
	r := s % KiB
	return float64(v) + float64(r)/float64(KiB)
}

func (s Size) MBytes() float64 {
	v := s / MB
	r := s % MB
	return float64(v) + float64(r)/float64(MB)
}

func (s Size) MiBytes() float64 {
	v := s / MiB
	r := s % MiB
	return float64(v) + float64(r)/float64(MiB)
}

func (s Size) GBytes() float64 {
	v := s / GB
	r := s % GB
	return float64(v) + float64(r)/float64(GB)
}

func (s Size) GiBytes() float64 {
	v := s / GiB
	r := s % GiB
	return float64(v) + float64(r)/float64(GiB)
}

func (s Size) TBytes() float64 {
	v := s / TB
	r := s % TB
	return float64(v) + float64(r)/float64(TB)
}

func (s Size) TiBytes() float64 {
	v := s / TiB
	r := s % TiB
	return float64(v) + float64(r)/float64(TiB)
}

func (s Size) PBytes() float64 {
	v := s / PB
	r := s % PB
	return float64(v) + float64(r)/float64(PB)
}

func (s Size) PiBytes() float64 {
	v := s / PiB
	r := s % PiB
	return float64(v) + float64(r)/float64(PiB)
}

func (s Size) EBytes() float64 {
	v := s / EB
	r := s % EB
	return float64(v) + float64(r)/float64(EB)
}

func (s Size) EiBytes() float64 {
	v := s / EiB
	r := s % EiB
	return float64(v) + float64(r)/float64(EiB)
}

func (s Size) String() string {
	return s.HumanReadable()
}

func (s Size) HumanReadable() string {
	switch {
	case s > EB:
		return fmt.Sprintf("%.1fEB", s.EBytes())
	case s > PB:
		return fmt.Sprintf("%.1fPB", s.PBytes())
	case s > TB:
		return fmt.Sprintf("%.1fTB", s.TBytes())
	case s > GB:
		return fmt.Sprintf("%.1fGB", s.GBytes())
	case s > MB:
		return fmt.Sprintf("%.1fMB", s.MBytes())
	case s > KB:
		return fmt.Sprintf("%.1fKB", s.KBytes())
	default:
		return fmt.Sprintf("%dB", s)
	}
}

func (s Size) HumanReadableBinary() string {
	switch {
	case s > EiB:
		return fmt.Sprintf("%.1fEiB", s.EiBytes())
	case s > PiB:
		return fmt.Sprintf("%.1fPiB", s.PiBytes())
	case s > TiB:
		return fmt.Sprintf("%.1fTiB", s.TiBytes())
	case s > GiB:
		return fmt.Sprintf("%.1fGiB", s.GiBytes())
	case s > MiB:
		return fmt.Sprintf("%.1fMiB", s.MiBytes())
	case s > KiB:
		return fmt.Sprintf("%.1fKiB", s.KiBytes())
	default:
		return fmt.Sprintf("%dB", s)
	}
}
