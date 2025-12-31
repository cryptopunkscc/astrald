package channel

type configFunc func(*channelConfig)

type channelConfig struct {
	fmtIn, fmtOut string
}

func InFmt(fmt string) func(*channelConfig) {
	return func(config *channelConfig) {
		config.fmtIn = fmt
	}
}

func OutFmt(fmt string) func(*channelConfig) {
	return func(config *channelConfig) {
		config.fmtOut = fmt
	}
}
