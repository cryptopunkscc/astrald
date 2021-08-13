package logfmt

func ID(s string) string {
	return s[len(s)-8:len(s)-4] + ":" + s[len(s)-4:]
}

func Dir(out bool) string {
	if out {
		return "out"
	}
	return "in"
}
