package channel

type ConfigFunc func(*Config)

type Config struct {
	fmtIn, fmtOut string
	allowUnparsed bool
}

func WithInputFormat(fmt string) func(*Config) {
	return func(config *Config) {
		config.fmtIn = fmt
	}
}

func WithOutputFormat(fmt string) func(*Config) {
	return func(config *Config) {
		config.fmtOut = fmt
	}
}

func WithFormats(fmtIn, fmtOut string) func(*Config) {
	return func(config *Config) {
		config.fmtIn = fmtIn
		config.fmtOut = fmtOut
	}
}

func WithFormat(fmt string) func(*Config) {
	return WithFormats(fmt, fmt)
}

func AllowUnparsed(b bool) func(*Config) {
	return func(config *Config) {
		config.allowUnparsed = b
	}
}
