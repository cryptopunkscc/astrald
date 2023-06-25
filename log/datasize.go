package log

import "fmt"

type DataSize uint64

const (
	B  DataSize = 1
	KB          = B << 10
	MB          = KB << 10
	GB          = MB << 10
	TB          = GB << 10
	PB          = TB << 10
	EB          = PB << 10
)

func (s DataSize) Bytes() uint64 {
	return uint64(s)
}

func (s DataSize) KBytes() float64 {
	v := s / KB
	r := s % KB
	return float64(v) + float64(r)/float64(KB)
}

func (s DataSize) MBytes() float64 {
	v := s / MB
	r := s % MB
	return float64(v) + float64(r)/float64(MB)
}

func (s DataSize) GBytes() float64 {
	v := s / GB
	r := s % GB
	return float64(v) + float64(r)/float64(GB)
}

func (s DataSize) TBytes() float64 {
	v := s / TB
	r := s % TB
	return float64(v) + float64(r)/float64(TB)
}

func (s DataSize) PBytes() float64 {
	v := s / PB
	r := s % PB
	return float64(v) + float64(r)/float64(PB)
}

func (s DataSize) EBytes() float64 {
	v := s / EB
	r := s % EB
	return float64(v) + float64(r)/float64(EB)
}

func (s DataSize) String() string {
	return s.HumanReadable()
}

func (s DataSize) HumanReadable() string {
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
